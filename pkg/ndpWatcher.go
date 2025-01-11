package pkg

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// NDPTableReader defines the interface for reading NDP table data
type NDPTableReader interface {
	// GetTable returns the current NDP table mapping MAC addresses to IPv6 addresses
	GetTable() map[string][]string
	// GetIP6forMAC returns IPv6 addresses for a specific MAC address
	GetIP6forMAC(mac string) ([]string, error)
}

// NDPTableWatcher implements NDPTableReader and provides background updates
type NDPTableWatcher struct {
	table     map[string][]string
	mu        sync.RWMutex
	done      chan bool
	interval  time.Duration
	debug     bool
	logger    Logger
	callbacks []func(map[string][]string)
}

// NDPTableWatcherConfig holds configuration for the NDP table watcher
type NDPTableWatcherConfig struct {
	UpdateInterval time.Duration
	Debug          bool
	Logger         Logger
}

// NewNDPTableWatcher creates a new NDP table watcher instance
func NewNDPTableWatcher(cfg NDPTableWatcherConfig) (*NDPTableWatcher, error) {
	if cfg.UpdateInterval == 0 {
		cfg.UpdateInterval = 30 * time.Second
	}

	watcher := &NDPTableWatcher{
		table:     make(map[string][]string),
		done:      make(chan bool),
		interval:  cfg.UpdateInterval,
		debug:     cfg.Debug,
		logger:    cfg.Logger,
		callbacks: make([]func(map[string][]string), 0),
	}

	// Perform initial table update
	if err := watcher.updateTable(); err != nil {
		return nil, fmt.Errorf("initial NDP table update failed: %w", err)
	}

	return watcher, nil
}

// Start begins the background NDP table monitoring
func (w *NDPTableWatcher) Start() {
	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if err := w.updateTable(); err != nil && w.debug {
					w.logger.Error(fmt.Sprintf("NDP table update failed: %v", err))
				}
			case <-w.done:
				if w.debug {
					w.logger.Info("NDP table watcher stopping")
				}
				return
			}
		}
	}()
}

// Stop terminates the background NDP table monitoring
func (w *NDPTableWatcher) Stop() {
	close(w.done)
}

// AddCallback registers a function to be called when the NDP table is updated
func (w *NDPTableWatcher) AddCallback(cb func(map[string][]string)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks = append(w.callbacks, cb)
}

// GetTable returns the current NDP table
func (w *NDPTableWatcher) GetTable() map[string][]string {
	w.mu.RLock()
	defer w.mu.RUnlock()

	// Create a copy of the table to prevent external modification
	tableCopy := make(map[string][]string, len(w.table))
	for mac, ips := range w.table {
		ipsCopy := make([]string, len(ips))
		copy(ipsCopy, ips)
		tableCopy[mac] = ipsCopy
	}

	return tableCopy
}

// GetIP6forMAC returns IPv6 addresses for a specific MAC address
func (w *NDPTableWatcher) GetIP6forMAC(mac string) ([]string, error) {
	mac = strings.ToUpper(mac)

	w.mu.RLock()
	defer w.mu.RUnlock()

	if ips, exists := w.table[mac]; exists {
		// Return a copy of the slice to prevent external modification
		result := make([]string, len(ips))
		copy(result, ips)
		return result, nil
	}
	return []string{}, nil
}

// updateTable refreshes the NDP table data
func (w *NDPTableWatcher) updateTable() error {
	//if w.debug {
	//	w.logger.Info("Updating NDP table")
	//}

	cmd := exec.Command("ndp", "-an")
	var out bytes.Buffer
	cmd.Stdout = &out

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("executing ndp command: %w", err)
	}

	newTable := make(map[string][]string)

	scanner := bufio.NewScanner(&out)
	// Skip header line
	if scanner.Scan() {
		_ = scanner.Text()
	}

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) >= 3 {
			ipv6 := fields[0]
			mac := strings.ToUpper(fields[1])
			newTable[mac] = append(newTable[mac], ipv6)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scanning ndp output: %w", err)
	}

	// Check for changes before updating
	if w.hasChanges(newTable) {
		w.mu.Lock()
		w.table = newTable
		callbacks := make([]func(map[string][]string), len(w.callbacks))
		copy(callbacks, w.callbacks)
		w.mu.Unlock()

		// Execute callbacks with a copy of the new table
		tableCopy := make(map[string][]string, len(newTable))
		for mac, ips := range newTable {
			ipsCopy := make([]string, len(ips))
			copy(ipsCopy, ips)
			tableCopy[mac] = ipsCopy
		}

		for _, cb := range callbacks {
			cb(tableCopy)
		}
	}

	return nil
}

// hasChanges compares the new table with the current one to detect changes
func (w *NDPTableWatcher) hasChanges(newTable map[string][]string) bool {
	w.mu.RLock()
	defer w.mu.RUnlock()

	if len(w.table) != len(newTable) {
		if w.debug {
			w.logger.Info(fmt.Sprintf("NDP table size changed: old=%d new=%d", len(w.table), len(newTable)))
		}
		return true
	}

	// Track removed MACs
	if w.debug {
		for mac := range w.table {
			if _, exists := newTable[mac]; !exists {
				w.logger.Info(fmt.Sprintf("MAC removed from NDP table: %s (had IPs: %v)", mac, w.table[mac]))
			}
		}
	}

	for mac, newIPs := range newTable {
		currentIPs, exists := w.table[mac]
		if !exists {
			if w.debug {
				w.logger.Info(fmt.Sprintf("New MAC in NDP table: %s (IPs: %v)", mac, newIPs))
			}
			return true
		}

		if len(currentIPs) != len(newIPs) {
			if w.debug {
				w.logger.Info(fmt.Sprintf("IP count changed for MAC %s: old=%d new=%d", mac, len(currentIPs), len(newIPs)))
			}
			return true
		}

		// Create maps for easier comparison
		currentIPMap := make(map[string]bool)
		for _, ip := range currentIPs {
			currentIPMap[ip] = true
		}

		for _, ip := range newIPs {
			if !currentIPMap[ip] {
				if w.debug {
					w.logger.Info(fmt.Sprintf("New IP for MAC %s: %s", mac, ip))
				}
				return true
			}
		}

		// Check for removed IPs
		if w.debug {
			newIPMap := make(map[string]bool)
			for _, ip := range newIPs {
				newIPMap[ip] = true
			}
			for _, ip := range currentIPs {
				if !newIPMap[ip] {
					w.logger.Info(fmt.Sprintf("IP removed from MAC %s: %s", mac, ip))
				}
			}
		}
	}

	return false
}

// IsValidMAC checks if a string is a valid MAC address
func IsValidMAC(addr string) bool {
	clean := strings.ReplaceAll(strings.ReplaceAll(addr, ":", ""), "-", "")

	if len(clean) != 12 {
		return false
	}

	for _, char := range clean {
		if !((char >= '0' && char <= '9') ||
			(char >= 'a' && char <= 'f') ||
			(char >= 'A' && char <= 'F')) {
			return false
		}
	}

	return true
}
