// pkg/sync.go
// pkg/sync.go
package pkg

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gmichels/adguard-client-go"
)

func NewSyncService(cfg Config) (*SyncService, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating file watcher: %w", err)
	}

	adguardClient, err := NewAdGuard(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating AdGuard client: %w", err)
	}

	return &SyncService{
		adguard:              adguardClient,
		leases:               NewDHCP(cfg.LeasePath),
		logger:               cfg.Logger,
		watcher:              watcher,
		done:                 make(chan bool),
		dryRun:               cfg.DryRun,
		preserveDeletedHosts: cfg.PreserveDeletedHosts,
	}, nil
}

func (s *SyncService) addClientWithRetry(hostname, mac, ip string) error {
	maxRetries := 10 // Maximum number of retries with different suffixes

	// First try without suffix
	err := s.adguard.AddClient(hostname, mac, ip)
	if err == nil {
		return nil
	}

	// If error is not name conflict, return immediately
	if !strings.Contains(strings.ToLower(err.Error()), "uses the same name") {
		return err
	}

	// Try with incremental suffixes
	for i := 1; i <= maxRetries; i++ {
		newName := fmt.Sprintf("%s-%d", hostname, i)
		err = s.adguard.AddClient(newName, mac, ip)
		if err == nil {
			s.logger.Info(fmt.Sprintf("Successfully added client with modified name: %s", newName))
			return nil
		}

		// If error is not name conflict, return immediately
		if !strings.Contains(strings.ToLower(err.Error()), "uses the same name") {
			return err
		}
	}

	return fmt.Errorf("failed to add client after %d retries: %v", maxRetries, err)
}

func (s *SyncService) updateClientWithRetry(existingClient *adguard.Client, lease ISCDHCPLease, mac string) error {
	maxRetries := 10 // Maximum number of retries with different suffixes

	// First try with original name
	updatedClient := *existingClient // Create a copy to preserve settings
	updatedClient.Name = lease.Hostname
	updatedClient.Ids = []string{mac, lease.IP}

	clientUpdate := adguard.ClientUpdate{
		Name: existingClient.Name, // Use existing name as identifier
		Data: updatedClient,
	}

	_, err := s.adguard.client.UpdateClient(clientUpdate)
	if err == nil {
		return nil
	}

	// If error is not name conflict, return immediately
	if !strings.Contains(strings.ToLower(err.Error()), "uses the same name") {
		return err
	}

	// Try with incremental suffixes
	for i := 1; i <= maxRetries; i++ {
		newName := fmt.Sprintf("%s-%d", lease.Hostname, i)
		updatedClient.Name = newName

		clientUpdate := adguard.ClientUpdate{
			Name: existingClient.Name, // Still use existing name as identifier
			Data: updatedClient,
		}

		_, err = s.adguard.client.UpdateClient(clientUpdate)
		if err == nil {
			s.logger.Info(fmt.Sprintf("Successfully updated client with modified name: %s", newName))
			return nil
		}

		// If error is not name conflict, return immediately
		if !strings.Contains(strings.ToLower(err.Error()), "uses the same name") {
			return err
		}
	}

	return fmt.Errorf("failed to update client after %d retries: %v", maxRetries, err)
}

func (s *SyncService) handleLeaseUpdate(existingClient *adguard.Client, lease ISCDHCPLease, mac string) error {
	action := fmt.Sprintf("Updating client %s (%s) with IP %s", lease.Hostname, mac, lease.IP)
	if s.dryRun {
		s.logger.Info("DRY-RUN: " + action)
		return nil
	}

	s.logger.Info(action)
	if err := s.updateClientWithRetry(existingClient, lease, mac); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to update client: %v", err))
		return err
	}
	return nil
}

func (s *SyncService) Sync() error {
	s.logger.Info("Starting sync")

	// Get current clients from AdGuard
	currentClients, err := s.adguard.GetClients()
	if err != nil {
		return fmt.Errorf("getting AdGuard clients: %w", err)
	}

	// Create maps for easier lookup
	currentClientsMap := make(map[string]*adguard.Client)
	for _, client := range currentClients {
		for _, id := range client.Ids {
			if strings.Contains(id, ":") { // MAC address check
				clientCopy := client
				currentClientsMap[id] = &clientCopy
				break
			}
		}
	}

	// Get current DHCP leases
	iscLeases, err := s.leases.GetLeases()
	if err != nil {
		return fmt.Errorf("getting DHCP leases: %w", err)
	}

	// Track processed MACs to identify stale entries
	processedMACs := make(map[string]bool)

	// Process active leases
	for mac, lease := range iscLeases {
		if !lease.IsActive {
			continue
		}

		processedMACs[mac] = true

		if existing := currentClientsMap[mac]; existing != nil {
			// Check if update needed
			needsUpdate := false
			ipFound := false

			for _, id := range existing.Ids {
				if id == lease.IP {
					ipFound = true
					needsUpdate = existing.Name != lease.Hostname
					break
				}
			}

			// Update needed if hostname changed or IP not found
			if needsUpdate || !ipFound {
				if err := s.handleLeaseUpdate(existing, lease, mac); err != nil {
					s.logger.Error(fmt.Sprintf("Error updating lease %s: %v", mac, err))
				}
			}
		} else if lease.Hostname != "" {
			// Add new client
			// Try adding client with incremental suffix if name conflict occurs
			err := s.addClientWithRetry(lease.Hostname, mac, lease.IP)
			if err != nil {
				s.logger.Error(fmt.Sprintf("Error adding lease %s: %v", mac, err))
			}
		}
	}

	// Handle stale clients only if deletion is not preserved
	if !s.preserveDeletedHosts {
		for mac, client := range currentClientsMap {
			if !processedMACs[mac] {
				action := fmt.Sprintf("Removing stale client %s (%s)", client.Name, mac)
				if s.dryRun {
					s.logger.Info("DRY-RUN: " + action)
					continue
				}

				s.logger.Info(action)
				if err := s.adguard.RemoveClient(mac); err != nil {
					s.logger.Error(fmt.Sprintf("Error removing stale client %s: %v", mac, err))
				}
			}
		}
	}

	return nil
}

// Run and Stop methods remain the same
func (s *SyncService) Run() error {
	s.logger.Info("Starting DHCP to AdGuard Home sync service")

	if err := s.Sync(); err != nil {
		s.logger.Error(fmt.Sprintf("Initial sync failed: %v", err))
	}

	leaseDir := filepath.Dir(s.leases.Path())
	if err := s.watcher.Add(leaseDir); err != nil {
		return fmt.Errorf("watching lease directory: %w", err)
	}

	var debounceTimer *time.Timer
	const debounceDelay = 2 * time.Second

	go func() {
		for {
			select {
			case event, ok := <-s.watcher.Events:
				if !ok {
					return
				}

				if event.Name != s.leases.Path() {
					continue
				}

				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(debounceDelay, func() {
					if err := s.Sync(); err != nil {
						s.logger.Error(fmt.Sprintf("Sync after file change failed: %v", err))
					}
				})

			case err, ok := <-s.watcher.Errors:
				if !ok {
					return
				}
				s.logger.Error(fmt.Sprintf("Watcher error: %v", err))

			case <-s.done:
				return
			}
		}
	}()

	return nil
}

func (s *SyncService) Stop() error {
	s.logger.Info("Stopping sync service")
	close(s.done)
	if err := s.watcher.Close(); err != nil {
		return fmt.Errorf("closing watcher: %w", err)
	}
	return nil
}
