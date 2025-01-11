package pkg

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gmichels/adguard-client-go"
)

// SyncService represents the DHCP to AdGuard sync service

// NewSyncService creates a new sync service instance
func NewSyncService(cfg Config) (*SyncService, error) {
	dhcpLeaseWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating file dhcpLeaseWatcher: %w", err)
	}

	adguardClient, err := NewAdGuard(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating AdGuard client: %w", err)
	}

	// Create NDP watcher
	ndpWatcher, err := NewNDPTableWatcher(NDPTableWatcherConfig{
		UpdateInterval: cfg.NDPUpdateInterval,
		Debug:          cfg.Debug,
		Logger:         cfg.Logger,
	})
	if err != nil {
		return nil, fmt.Errorf("creating NDP table watcher: %w", err)

	}

	service := &SyncService{
		adguard:              adguardClient,
		leases:               NewDHCP(cfg.LeasePath),
		logger:               cfg.Logger,
		dhcpLeaseWatcher:     dhcpLeaseWatcher,
		ndpWatcher:           ndpWatcher,
		done:                 make(chan bool),
		dryRun:               cfg.DryRun,
		preserveDeletedHosts: cfg.PreserveDeletedHosts,
		debug:                cfg.Debug,
	}
	// Register callback for NDP table updates
	ndpWatcher.AddCallback(service.handleNDPUpdate)

	if service.debug {
		service.logger.Info("Created new SyncService with config:")
		service.logger.Info(fmt.Sprintf("- Lease path: %s", cfg.LeasePath))
		service.logger.Info(fmt.Sprintf("- Dry run: %v", cfg.DryRun))
		service.logger.Info(fmt.Sprintf("- Preserve deleted hosts: %v", cfg.PreserveDeletedHosts))
		service.logger.Info(fmt.Sprintf("- Debug mode: enabled"))
		service.logger.Info(fmt.Sprintf("- NDP update interval: %v", cfg.NDPUpdateInterval))

	}

	return service, nil
}

// handleNDPUpdate is called when the NDP table changes
func (s *SyncService) handleNDPUpdate(ndpTable map[string][]string) {
	if s.debug {
		s.logger.Info("NDP table update detected")
	}

	// Trigger a sync when NDP table changes
	if err := s.Sync(); err != nil {
		s.logger.Error(fmt.Sprintf("Sync after NDP update failed: %v", err))
	}
}
func (s *SyncService) addClientWithRetry(action *AdguardUpdateAction) error {
	if s.debug {
		s.logger.Info(fmt.Sprintf("Attempting to add client - hostname: %s, MAC: %s, IDs: %v",
			action.Hostname, action.MAC, action.IDs))
	}

	maxRetries := 10
	hostname := action.Hostname

	// First try without suffix
	err := s.adguard.AddClient(hostname, action.MAC, action.IDs)
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

		err = s.adguard.AddClient(newName, action.MAC, action.IDs)
		if err == nil {
			s.logger.Info(fmt.Sprintf("Successfully added client with modified name: %s", newName))
			return nil
		}

		if s.debug {
			s.logger.Info(fmt.Sprintf("Retry %d failed: %v", i, err))
		}

		// Only continue retrying if it's a name conflict
		if !strings.Contains(strings.ToLower(err.Error()), "uses the same name") {
			return err
		}
	}

	return fmt.Errorf("failed to add client after %d retries: %v", maxRetries, err)
}

//ips := []string{ip}
//
//// Try to get IPv6 addresses
//ip6Addresses, err := s.ndpWatcher.GetIP6forMAC(mac)
//if err == nil && len(ip6Addresses) > 0 {
//	// Append all IPv6 addresses to our IPs array
//	ips = append(ips, ip6Addresses...)
//}
//
//if s.debug {
//	s.logger.Info(fmt.Sprintf("Attempting to add client - hostname: %s, MAC: %s, IPs: %s", hostname, mac, ips))
//}
//
//maxRetries := 10
//
//// First try without suffix
//err = s.adguard.AddClient(hostname, mac, ips)
//if err == nil {
//	if s.debug {
//		s.logger.Info(fmt.Sprintf("Successfully added client %s on first attempt", hostname))
//	}
//	return nil
//}
//
//if s.debug {
//	s.logger.Info(fmt.Sprintf("Initial add attempt failed: %v", err))
//}
//
//// If error is not name conflict, return immediately
//if !strings.Contains(strings.ToLower(err.Error()), "uses the same name") {
//	return err
//}
//
//// Try with incremental suffixes
//for i := 1; i <= maxRetries; i++ {
//	newName := fmt.Sprintf("%s-%d", hostname, i)
//	if s.debug {
//		s.logger.Info(fmt.Sprintf("Attempting retry %d with name: %s", i, newName))
//	}
//
//	err = s.adguard.AddClient(newName, mac, ips)
//	if err == nil {
//		s.logger.Info(fmt.Sprintf("Successfully added client with modified name: %s", newName))
//		return nil
//	}
//
//	if s.debug {
//		s.logger.Info(fmt.Sprintf("Retry %d failed: %v", i, err))
//	}
//
//	if !strings.Contains(strings.ToLower(err.Error()), "uses the same name") {
//		return err
//	}
//}
//
//return fmt.Errorf("failed to add client after %d retries: %v", maxRetries, err)

func (s *SyncService) updateClient(existingClient *adguard.Client, action *AdguardUpdateAction) error {
	if s.debug {
		s.logger.Info(fmt.Sprintf("Attempting to update client - MAC: %s, Current name: %s, New name: %s",
			action.MAC, existingClient.Name, action.Hostname))
	}

	hostname := action.Hostname

	// First try with original name
	updatedClient := *existingClient
	updatedClient.Name = hostname
	updatedClient.Ids = action.IDs

	clientUpdate := adguard.ClientUpdate{
		Name: existingClient.Name,
		Data: updatedClient,
	}

	if s.debug {
		s.logger.Info(fmt.Sprintf("Updating: %s", hostname))
	}

	_, err := s.adguard.client.UpdateClient(clientUpdate)
	if err == nil {
		if s.debug {
			s.logger.Info(fmt.Sprintf("Successfully updated client"))
		}
		return nil
	}

	return fmt.Errorf("failed to update client: %v", err)
}

//func (s *SyncService) handleLeaseUpdate(existingClient *adguard.Client, lease ISCDHCPLease, mac string) error {
//
//	// Build an update action
//	action := &AdguardUpdateAction{
//		Type:     Update,
//		Hostname: lease.Hostname,
//		MAC:      mac,
//	}
//
//	// Get IPv6 IDs
//	ipv6IDs, err := s.ndpWatcher.GetIP6forMAC(mac)
//	if err != nil {
//		ipv6IDs = []string{}
//	}
//	action.IDs = ipv6IDs
//	action.IDs = append(action.IDs, lease.IP)
//	//
//	//
//	//
//	//// Try RDNS if hostname is empty
//	//if hostname == "" {
//	//	names, err := net.LookupAddr(lease.IP)
//	//	if err == nil && len(names) > 0 {
//	//		hostname = strings.TrimSuffix(names[0], ".")
//	//		if s.debug {
//	//			s.logger.Info(fmt.Sprintf("Using RDNS hostname %s for update of MAC %s", hostname, mac))
//	//		}
//	//	} else {
//	//		s.logger.Info(fmt.Sprintf("No hostname and no RDNS hostname available for MAC %s", mac))
//	//		return fmt.Errorf("no hostname available for update")
//	//	}
//	//}
//	//
//	//// Create a modified lease with the RDNS hostname if needed
//	//updatedLease := lease
//	//updatedLease.Hostname = hostname
//
//	//ip6Addresses, _ := s.ndpWatcher.GetIP6forMAC(mac)
//	//
//	//action := fmt.Sprintf("Updating client [%s] (%s) with IP4 [%s] IP6 %s", hostname, mac, lease.IP, ip6Addresses)
//	//
//	//if s.dryRun {
//	//	s.logger.Info("DRY-RUN: " + action)
//	//	return nil
//	//}
//	//
//	//s.logger.Info(action)
//	if err := s.updateClient(existingClient, action); err != nil {
//		s.logger.Error(fmt.Sprintf("Failed to update client: %v", err))
//		return err
//	}
//	return nil
//}

// determineUpdateAction checks if and what kind of update is needed for a given lease
func (s *SyncService) determineUpdateAction(lease ISCDHCPLease, mac string, existing *adguard.Client) (*AdguardUpdateAction, error) {
	action := &AdguardUpdateAction{
		Type:     NoUpdate,
		Hostname: lease.Hostname,
		MAC:      mac,
	}

	// Get IPv6 IDs
	ipv6IDs, err := s.ndpWatcher.GetIP6forMAC(mac)
	if err != nil {
		ipv6IDs = []string{}
	}

	// Calculate RDNS
	//rdnsNames, err := net.LookupAddr(lease.IP)

	// Build wanted IDs list
	action.IDs = ipv6IDs
	//action.IDs = append(action.IDs, mac) // MAC is needed for an update - but seems to automatically be added to an Add
	action.IDs = append(action.IDs, lease.IP)
	// Skip RDNS
	//if err == nil && len(rdnsNames) > 0 {
	//	action.IDs = append(action.IDs, strings.Split(strings.TrimSuffix(rdnsNames[0], "."), ".")[0])
	//}

	// If no existing client, this is an Add
	if existing == nil {
		if action.Hostname == "" {
			return action, fmt.Errorf("no hostname available for new client")
		}

		action.Type = Add
		action.Reason = "new client"
		return action, nil
	}

	// Compare existing vs wanted IDs
	existingIDsMap := make(map[string]bool)
	for _, id := range existing.Ids {
		if id != mac { // exclude mac address (I think)
			existingIDsMap[id] = true
		}
	}

	// Ensure we also copy the mac address into this
	wantedIDsMap := make(map[string]bool)
	for _, id := range action.IDs {
		wantedIDsMap[id] = true
	}

	// Check for missing IDs
	for id := range wantedIDsMap {
		if !existingIDsMap[id] {
			if s.debug {
				s.logger.Info(fmt.Sprintf("Missing ID found: %s for %s", id, mac))
			}
			action.NeedsUpdate = true
			action.Reason = fmt.Sprintf("missing ID: %s", id)
			break
		}
	}

	// Check for extra IDs
	if !action.NeedsUpdate {
		for id := range existingIDsMap {
			if !wantedIDsMap[id] {
				if s.debug {
					s.logger.Info(fmt.Sprintf("Extra ID found: %s", id))
				}
				action.NeedsUpdate = true
				action.Reason = fmt.Sprintf("extra ID: %s", id)
				break
			}
		}
	}

	// Check if IP is found
	for _, id := range action.IDs {
		if id == lease.IP {
			action.IPFound = true
			break
		}
	}

	if action.NeedsUpdate || !action.IPFound {
		action.Type = Update
		// Update actions require the MAC address to be added
		action.IDs = append(action.IDs, mac)
	}

	return action, nil
}

// buildClientMap creates a map of MAC addresses to AdGuard clients for efficient lookup
func (s *SyncService) buildClientMap(clients []adguard.Client) map[string]*adguard.Client {
	if s.debug {
		s.logger.Info("Building client MAC address map")
	}

	currentClientsMap := make(map[string]*adguard.Client)
	for _, client := range clients {
		for _, id := range client.Ids {
			if IsValidMAC(id) {
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
		s.logger.Info(fmt.Sprintf("Built map with %d clients", len(currentClientsMap)))
	}

	return currentClientsMap
}

func (s *SyncService) Sync() error {
	s.logger.Info("Starting sync")

	// Get current clients from AdGuard
	currentClients, err := s.adguard.GetClients()
	if err != nil {
		return fmt.Errorf("getting AdGuard clients: %w", err)
	}

	// Create MAC address lookup map
	currentClientsMap := s.buildClientMap(currentClients)

	// Get current DHCP leases
	iscLeases, err := s.leases.GetLeases()
	if err != nil {
		return fmt.Errorf("getting DHCP leases: %w", err)
	}

	// Track processed MACs
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
		existing := currentClientsMap[mac]

		// Retrive the update action
		action, err := s.determineUpdateAction(lease, mac, existing)
		if err != nil {
			s.logger.Error(fmt.Sprintf("Error determining update action for %s: %v", mac, err))
			continue
		}

		switch action.Type {
		case Update:

			if err := s.updateClient(existing, action); err != nil {
				s.logger.Error(fmt.Sprintf("Error updating lease %s: %v", mac, err))
			}
		case Add:
			if err := s.addClientWithRetry(action); err != nil {
				s.logger.Error(fmt.Sprintf("Error adding lease %s: %v", mac, err))
			}
		}
	}

	// Handle stale clients
	if err := s.handleStaleClients(currentClientsMap, processedMACs); err != nil {
		s.logger.Error(fmt.Sprintf("Error handling stale clients: %v", err))
	}

	s.logger.Info("Sync completed")
	return nil
}

func (s *SyncService) handleStaleClients(currentClients map[string]*adguard.Client, processedMACs map[string]bool) error {
	if s.preserveDeletedHosts {
		if s.debug {
			s.logger.Info("Skipping stale client removal (preserveDeletedHosts is enabled)")
		}
		return nil
	}

	if s.debug {
		s.logger.Info("Checking for stale clients")
	}

	for mac, client := range currentClients {
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
				return fmt.Errorf("removing stale client %s: %w", mac, err)
			}
		}
	}

	return nil
}

func (s *SyncService) Run() error {
	s.logger.Info("Starting DHCP to AdGuard Home sync service")

	// Start the NDP Watcher
	s.ndpWatcher.Start()

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
		s.logger.Info(fmt.Sprintf("Setting up dhcpLeaseWatcher for directory: %s", leaseDir))
	}

	if err := s.dhcpLeaseWatcher.Add(leaseDir); err != nil {
		return fmt.Errorf("watching lease directory: %w", err)
	}

	if s.debug {
		s.logger.Info("File dhcpLeaseWatcher setup complete")
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
			case event, ok := <-s.dhcpLeaseWatcher.Events:
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

			case err, ok := <-s.dhcpLeaseWatcher.Errors:
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

	// Stop NDP watcher first
	s.ndpWatcher.Stop()

	// Signal the main loop to stop
	close(s.done)

	// Close the DHCP lease watcher
	if err := s.dhcpLeaseWatcher.Close(); err != nil {
		return fmt.Errorf("closing dhcp lease watcher: %w", err)
	}

	return nil
}
