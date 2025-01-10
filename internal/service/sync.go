package service

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"opnsense-lease-sync/internal/adguard"
	"opnsense-lease-sync/internal/dhcp"
	"opnsense-lease-sync/internal/logger"
	"path/filepath"
	"time"
)

type SyncService struct {
	adguardClient *adguard.Client
	leasePath     string
	logger        logger.Logger
	watcher       *fsnotify.Watcher
	done          chan bool
}

func New(adguardURL, leasePath string, logger logger.Logger) (*SyncService, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating file watcher: %w", err)
	}

	return &SyncService{
		adguardClient: adguard.New(adguardURL),
		leasePath:     leasePath,
		logger:        logger,
		watcher:       watcher,
		done:          make(chan bool),
	}, nil
}

// initialSync gets the current state from both AdGuard and the lease file
// and performs a full sync
func (s *SyncService) initialSync() error {
	s.logger.Info("Performing initial sync")

	// Get current AdGuard clients
	adguardClients, err := s.adguardClient.GetClients()
	if err != nil {
		return fmt.Errorf("getting AdGuard clients: %w", err)
	}

	// Create a map of MAC -> Client for easy lookup
	currentAdGuardClients := make(map[string]adguard.Client)
	for _, client := range adguardClients {
		// Assuming the first ID is the MAC address
		if len(client.IDs) > 0 {
			currentAdGuardClients[client.IDs[0]] = client
		}
	}

	// Get current DHCP leases
	dhcpLeases, err := dhcp.ParseLeaseFile(s.leasePath)
	if err != nil {
		return fmt.Errorf("parsing lease file: %w", err)
	}

	// Sync the states
	for mac, lease := range dhcpLeases {
		// Check if lease exists in AdGuard
		if adguardClient, exists := currentAdGuardClients[mac]; exists {
			// Update if different
			if adguardClient.Name != lease.Hostname || adguardClient.IP != lease.IP {
				if err := s.adguardClient.UpdateClient(lease.Hostname, lease.IP, lease.MAC); err != nil {
					s.logger.Error(fmt.Sprintf("Error updating lease in AdGuard: %v", err))
				}
			}
			// Remove from map to track what's been processed
			delete(currentAdGuardClients, mac)
		} else {
			// Add new lease to AdGuard
			if err := s.adguardClient.UpdateClient(lease.Hostname, lease.IP, lease.MAC); err != nil {
				s.logger.Error(fmt.Sprintf("Error adding lease to AdGuard: %v", err))
			}
		}
	}

	// Remove any AdGuard clients that don't have active DHCP leases
	for mac := range currentAdGuardClients {
		if err := s.adguardClient.RemoveClient(mac); err != nil {
			s.logger.Error(fmt.Sprintf("Error removing client from AdGuard: %v", err))
		}
	}

	return nil
}

func (s *SyncService) Run() error {
	s.logger.Info("DHCP to AdGuard Home sync service starting")

	// Perform initial sync
	if err := s.initialSync(); err != nil {
		s.logger.Error(fmt.Sprintf("Initial sync failed: %v", err))
	}

	// Watch the directory containing the lease file
	leaseDir := filepath.Dir(s.leasePath)
	if err := s.watcher.Add(leaseDir); err != nil {
		return fmt.Errorf("watching lease directory: %w", err)
	}

	// Debounce timer to prevent multiple rapid updates
	var debounceTimer *time.Timer
	const debounceDelay = 2 * time.Second

	go func() {
		for {
			select {
			case event, ok := <-s.watcher.Events:
				if !ok {
					return
				}

				// Only process events for our lease file
				if event.Name != s.leasePath {
					continue
				}

				// Reset or start the debounce timer
				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(debounceDelay, func() {
					// On file change, just do a full sync again
					if err := s.initialSync(); err != nil {
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
