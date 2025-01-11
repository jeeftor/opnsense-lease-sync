package pkg

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gmichels/adguard-client-go"
)

// SyncService represents the DHCP to AdGuard sync service

// NewSyncService creates a new sync service instance
func NewSyncService(cfg Config) (*SyncService, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating file watcher: %w", err)
	}

	adguardClient, err := NewAdGuard(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating AdGuard client: %w", err)
	}

	service := &SyncService{
		adguard:              adguardClient,
		leases:               NewDHCP(cfg.LeasePath),
		logger:               cfg.Logger,
		watcher:              watcher,
		done:                 make(chan bool),
		dryRun:               cfg.DryRun,
		preserveDeletedHosts: cfg.PreserveDeletedHosts,
		debug:                cfg.Debug,
	}

	if service.debug {
		service.logger.Info("Created new SyncService with config:")
		service.logger.Info(fmt.Sprintf("- Lease path: %s", cfg.LeasePath))
		service.logger.Info(fmt.Sprintf("- Dry run: %v", cfg.DryRun))
		service.logger.Info(fmt.Sprintf("- Preserve deleted hosts: %v", cfg.PreserveDeletedHosts))
		service.logger.Info(fmt.Sprintf("- Debug mode: enabled"))
	}

	return service, nil
}

func (s *SyncService) addClientWithRetry(hostname, mac, ip string) error {
	if s.debug {
		s.logger.Info(fmt.Sprintf("Attempting to add client - hostname: %s, MAC: %s, IP: %s", hostname, mac, ip))
	}

	maxRetries := 10

	// First try without suffix
	err := s.adguard.AddClient(hostname, mac, ip)
	if err == nil {
		if s.debug {
			s.logger.Info(fmt.Sprintf("Successfully added client %s on first attempt", hostname))
		}
		return nil
	}

	if s.debug {
		s.logger.Info(fmt.Sprintf("Initial add attempt failed: %v", err))
	}

	// If error is not name conflict, return immediately
	if !strings.Contains(strings.ToLower(err.Error()), "uses the same name") {
		return err
	}

	// Try with incremental suffixes
	for i := 1; i <= maxRetries; i++ {
		newName := fmt.Sprintf("%s-%d", hostname, i)
		if s.debug {
			s.logger.Info(fmt.Sprintf("Attempting retry %d with name: %s", i, newName))
		}

		err = s.adguard.AddClient(newName, mac, ip)
		if err == nil {
			s.logger.Info(fmt.Sprintf("Successfully added client with modified name: %s", newName))
			return nil
		}

		if s.debug {
			s.logger.Info(fmt.Sprintf("Retry %d failed: %v", i, err))
		}

		if !strings.Contains(strings.ToLower(err.Error()), "uses the same name") {
			return err
		}
	}

	return fmt.Errorf("failed to add client after %d retries: %v", maxRetries, err)
}

func (s *SyncService) updateClientWithRetry(existingClient *adguard.Client, lease ISCDHCPLease, mac string) error {
	if s.debug {
		s.logger.Info(fmt.Sprintf("Attempting to update client - MAC: %s, Current name: %s, New name: %s",
			mac, existingClient.Name, lease.Hostname))
	}

	maxRetries := 10

	// First try with original name
	updatedClient := *existingClient
	updatedClient.Name = lease.Hostname
	updatedClient.Ids = []string{mac, lease.IP}

	clientUpdate := adguard.ClientUpdate{
		Name: existingClient.Name,
		Data: updatedClient,
	}

	if s.debug {
		s.logger.Info(fmt.Sprintf("Attempting initial update with original name: %s", lease.Hostname))
	}

	_, err := s.adguard.client.UpdateClient(clientUpdate)
	if err == nil {
		if s.debug {
			s.logger.Info(fmt.Sprintf("Successfully updated client on first attempt"))
		}
		return nil
	}

	if s.debug {
		s.logger.Info(fmt.Sprintf("Initial update attempt failed: %v", err))
	}

	if !strings.Contains(strings.ToLower(err.Error()), "uses the same name") {
		return err
	}

	// Try with incremental suffixes
	for i := 1; i <= maxRetries; i++ {
		newName := fmt.Sprintf("%s-%d", lease.Hostname, i)
		if s.debug {
			s.logger.Info(fmt.Sprintf("Attempting retry %d with name: %s", i, newName))
		}

		updatedClient.Name = newName
		clientUpdate.Data = updatedClient

		_, err = s.adguard.client.UpdateClient(clientUpdate)
		if err == nil {
			s.logger.Info(fmt.Sprintf("Successfully updated client with modified name: %s", newName))
			return nil
		}

		if s.debug {
			s.logger.Info(fmt.Sprintf("Retry %d failed: %v", i, err))
		}

		if !strings.Contains(strings.ToLower(err.Error()), "uses the same name") {
			return err
		}
	}

	return fmt.Errorf("failed to update client after %d retries: %v", maxRetries, err)
}

func (s *SyncService) handleLeaseUpdate(existingClient *adguard.Client, lease ISCDHCPLease, mac string) error {
	hostname := lease.Hostname
	// Try RDNS if hostname is empty
	if hostname == "" {
		names, err := net.LookupAddr(lease.IP)
		if err == nil && len(names) > 0 {
			hostname = strings.TrimSuffix(names[0], ".")
			if s.debug {
				s.logger.Info(fmt.Sprintf("Using RDNS hostname %s for update of MAC %s", hostname, mac))
			}
		} else {
			s.logger.Info(fmt.Sprintf("No hostname and no RDNS hostname available for MAC %s", mac))
			return fmt.Errorf("no hostname available for update")
		}
	}

	// Create a modified lease with the RDNS hostname if needed
	updatedLease := lease
	updatedLease.Hostname = hostname

	action := fmt.Sprintf("Updating client [%s] (%s) with IP %s", hostname, mac, lease.IP)

	if s.dryRun {
		s.logger.Info("DRY-RUN: " + action)
		return nil
	}

	s.logger.Info(action)
	if err := s.updateClientWithRetry(existingClient, updatedLease, mac); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to update client: %v", err))
		return err
	}
	return nil
}

func (s *SyncService) Sync() error {
	s.logger.Info("Starting sync")

	if s.debug {
		s.logger.Info("Fetching current clients from AdGuard")
	}

	// Get current clients from AdGuard
	currentClients, err := s.adguard.GetClients()
	if err != nil {
		return fmt.Errorf("getting AdGuard clients: %w", err)
	}

	if s.debug {
		s.logger.Info(fmt.Sprintf("Retrieved %d clients from AdGuard", len(currentClients)))
	}

	// Create maps for easier lookup
	currentClientsMap := make(map[string]*adguard.Client)
	for _, client := range currentClients {
		for _, id := range client.Ids {
			if strings.Contains(id, ":") { // MAC address check
				clientCopy := client
				currentClientsMap[id] = &clientCopy
				if s.debug {
					s.logger.Info(fmt.Sprintf("Mapped client %s to MAC %s", client.Name, id))
				}
				break
			}
		}
	}

	if s.debug {
		s.logger.Info("Fetching current DHCP leases")
	}

	// Get current DHCP leases
	iscLeases, err := s.leases.GetLeases()
	if err != nil {
		return fmt.Errorf("getting DHCP leases: %w", err)
	}

	if s.debug {
		s.logger.Info(fmt.Sprintf("Retrieved %d DHCP leases", len(iscLeases)))
	}

	// Track processed MACs to identify stale entries
	processedMACs := make(map[string]bool)

	// Process active leases
	for mac, lease := range iscLeases {
		if !lease.IsActive {
			if s.debug {
				s.logger.Info(fmt.Sprintf("Skipping inactive lease for MAC %s", mac))
			}
			continue
		}

		processedMACs[mac] = true

		if s.debug {
			s.logger.Info(fmt.Sprintf("Processing lease - MAC: %s, Hostname: %s, IP: %s",
				mac, lease.Hostname, lease.IP))
		}

		if existing := currentClientsMap[mac]; existing != nil {
			if s.debug {
				s.logger.Info(fmt.Sprintf("Found existing client for MAC %s", mac))
			}

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

			if s.debug {
				s.logger.Info(fmt.Sprintf("Update check - IP found: %v, Needs update: %v", ipFound, needsUpdate))
			}

			// Update needed if hostname changed or IP not found
			if needsUpdate || !ipFound {
				if err := s.handleLeaseUpdate(existing, lease, mac); err != nil {
					s.logger.Error(fmt.Sprintf("Error updating lease %s: %v", mac, err))
				}
			}
		} else {
			hostname := lease.Hostname

			if hostname == "" {
				// Try reverse lookup
				// Try reverse DNS lookup if no hostname provided
				names, err := net.LookupAddr(lease.IP)
				if err == nil && len(names) > 0 {
					// Remove trailing dot from hostname if present
					hostname = strings.TrimSuffix(names[0], ".")

					if s.debug {
						s.logger.Info(fmt.Sprintf("No existing client found for MAC %s, using RDNS hostname %s", mac, hostname))
					}
				} else {
					s.logger.Info(fmt.Sprintf("No existing client found for MAC %s and no RDNS hostname available", mac))
					continue
				}
			}
			err = s.addClientWithRetry(hostname, mac, lease.IP)
			if err != nil {
				s.logger.Error(fmt.Sprintf("Error adding lease %s: %v", mac, err))
			}

		}
	}

	// Handle stale clients only if deletion is not preserved
	if !s.preserveDeletedHosts {
		if s.debug {
			s.logger.Info("Checking for stale clients")
		}

		for mac, client := range currentClientsMap {
			if !processedMACs[mac] {
				action := fmt.Sprintf("Removing stale client %s (%s)", client.Name, mac)
				if s.dryRun {
					s.logger.Info("DRY-RUN: " + action)
					continue
				}

				if s.debug {
					s.logger.Info(fmt.Sprintf("Found stale client - MAC: %s, Name: %s", mac, client.Name))
				}

				s.logger.Info(action)
				if err := s.adguard.RemoveClient(client.Name); err != nil {
					s.logger.Error(fmt.Sprintf("Error removing stale client %s: %v", mac, err))
				}
			}
		}
	} else if s.debug {
		s.logger.Info("Skipping stale client removal (preserveDeletedHosts is enabled)")
	}

	s.logger.Info("Sync completed")
	return nil
}

func (s *SyncService) Run() error {
	s.logger.Info("Starting DHCP to AdGuard Home sync service")

	// Convert to absolute path
	absPath, err := filepath.Abs(s.leases.Path())
	if err != nil {
		return fmt.Errorf("getting absolute path: %w", err)
	}

	if s.debug {
		s.logger.Info(fmt.Sprintf("Lease file absolute path: %s", absPath))
	}

	// Verify the lease file exists
	if _, err := os.Stat(absPath); err != nil {
		return fmt.Errorf("accessing lease file: %w", err)
	}

	// Get the directory from the absolute path
	leaseDir := filepath.Dir(absPath)
	if s.debug {
		s.logger.Info(fmt.Sprintf("Setting up watcher for directory: %s", leaseDir))
	}

	if err := s.watcher.Add(leaseDir); err != nil {
		return fmt.Errorf("watching lease directory: %w", err)
	}

	if s.debug {
		s.logger.Info("File watcher setup complete")
	}

	// Perform initial sync
	if err := s.Sync(); err != nil {
		s.logger.Error(fmt.Sprintf("Initial sync failed: %v", err))
	}

	var debounceTimer *time.Timer
	const debounceDelay = 2 * time.Second

	go func() {
		for {
			select {
			case event, ok := <-s.watcher.Events:
				if !ok {
					s.logger.Info("Watcher events channel closed")
					return
				}

				if s.debug {
					s.logger.Info(fmt.Sprintf("Received file event: %s on %s", event.Op, event.Name))
				}

				// Get absolute path for comparison
				eventPath, err := filepath.Abs(event.Name)
				if err != nil {
					s.logger.Error(fmt.Sprintf("Failed to get absolute path for event: %v", err))
					continue
				}

				if eventPath != absPath {
					if s.debug {
						s.logger.Info(fmt.Sprintf("Ignoring event for non-target file - Event: %s, Target: %s",
							eventPath, absPath))
					}
					continue
				}

				if debounceTimer != nil {
					if s.debug {
						s.logger.Info("Stopping previous debounce timer")
					}
					debounceTimer.Stop()
				}

				if s.debug {
					s.logger.Info(fmt.Sprintf("Starting debounce timer (%s)", debounceDelay))
				}

				debounceTimer = time.AfterFunc(debounceDelay, func() {
					if s.debug {
						s.logger.Info("Debounce timer expired, starting sync")
					}
					if err := s.Sync(); err != nil {
						s.logger.Error(fmt.Sprintf("Sync after file change failed: %v", err))
					}
				})

			case err, ok := <-s.watcher.Errors:
				if !ok {
					s.logger.Info("Watcher errors channel closed")
					return
				}
				s.logger.Error(fmt.Sprintf("Watcher error: %v", err))

			case <-s.done:
				if s.debug {
					s.logger.Info("Received shutdown signal")
				}
				return
			}
		}
	}()

	return nil
}

func (s *SyncService) Stop() error {
	if s.debug {
		s.logger.Info("Stop requested - shutting down sync service")
	}
	s.logger.Info("Stopping sync service")
	close(s.done)
	if err := s.watcher.Close(); err != nil {
		return fmt.Errorf("closing watcher: %w", err)
	}
	return nil
}
