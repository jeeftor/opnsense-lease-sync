package pkg

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type SyncService struct {
	adguard *AdGuard
	leases  *DHCP
	logger  Logger
	watcher *fsnotify.Watcher
	done    chan bool
	dryRun  bool
}

func NewSyncService(cfg Config) (*SyncService, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating file watcher: %w", err)
	}

	adguard := NewAdGuard(cfg.AdGuardURL, cfg.B64Auth)
	return &SyncService{
		adguard: adguard,
		leases:  NewDHCP(cfg.LeasePath),
		logger:  cfg.Logger,
		watcher: watcher,
		done:    make(chan bool),
		dryRun:  cfg.DryRun,
	}, nil
}

func (s *SyncService) logAPICall(method, endpoint string, payload interface{}) {
	s.logger.Info(fmt.Sprintf("DRY-RUN: Would call %s %s/control/%s with Authorization: Basic ****** and payload: %+v",
		method, s.adguard.BaseURL, endpoint, payload))
}

func (s *SyncService) getAdGuardClients() (map[string]*AdGuardClient, error) {
	clients, err := s.adguard.GetClients()
	if err != nil {
		if s.dryRun {
			s.logger.Error(fmt.Sprintf("Unable to connect to AdGuard (continuing in dry-run mode): %v", err))
			s.logAPICall("GET", "clients", nil)
			return make(map[string]*AdGuardClient), nil
		}
		return nil, fmt.Errorf("getting AdGuard clients: %w", err)
	}

	currentClients := make(map[string]*AdGuardClient)
	for _, client := range clients {
		if len(client.IDs) > 0 {
			clientCopy := client
			currentClients[client.IDs[0]] = &clientCopy
		}
	}
	return currentClients, nil
}

func (s *SyncService) handleClientUpdate(lease DHCPLease, mac string, existing *AdGuardClient) error {
	if existing.Name == lease.Hostname && existing.IP == lease.IP {
		return nil
	}

	payload := AdGuardClient{
		Name:                existing.Name,
		IDs:                 []string{mac},
		IP:                  lease.IP,
		UseGlobalSettings:   existing.UseGlobalSettings,
		FilteringEnabled:    existing.FilteringEnabled,
		ParentalEnabled:     existing.ParentalEnabled,
		SafebrowsingEnabled: existing.SafebrowsingEnabled,
		SafeSearch:          existing.SafeSearch,
	}

	action := fmt.Sprintf("Updating client %s (%s) with IP %s", lease.Hostname, mac, lease.IP)
	if s.dryRun {
		s.logger.Info("DRY-RUN: " + action)
		s.logAPICall("PUT", "clients/update", payload)
		return nil
	}

	s.logger.Info(action)
	if err := s.adguard.UpdateClient(lease.Hostname, lease.IP, mac, existing); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to update client: %v", err))
		return err
	}
	return nil
}

func (s *SyncService) handleClientAdd(lease DHCPLease, mac string) error {
	payload := AdGuardClient{
		Name:                lease.Hostname,
		IDs:                 []string{mac},
		IP:                  lease.IP,
		UseGlobalSettings:   true,
		FilteringEnabled:    true,
		ParentalEnabled:     false,
		SafebrowsingEnabled: false,
		SafeSearch:          false,
	}

	action := fmt.Sprintf("Adding new client %s (%s) with IP %s", lease.Hostname, mac, lease.IP)
	if s.dryRun {
		s.logger.Info("DRY-RUN: " + action)
		s.logAPICall("POST", "clients/add", payload)
		return nil
	}

	s.logger.Info(action)
	if err := s.adguard.AddClient(lease.Hostname, lease.IP, mac); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to add client: %v", err))
		return err
	}
	return nil
}

func (s *SyncService) handleClientRemove(client *AdGuardClient, mac string) error {
	action := fmt.Sprintf("Removing client %s (%s)", client.Name, mac)
	if s.dryRun {
		s.logger.Info("DRY-RUN: " + action)
		s.logAPICall("DELETE", fmt.Sprintf("clients/delete?mac=%s", mac), nil)
		return nil
	}

	s.logger.Info(action)
	if err := s.adguard.RemoveClient(mac); err != nil {
		s.logger.Error(fmt.Sprintf("Failed to remove client: %v", err))
		return err
	}
	return nil
}

func (s *SyncService) Sync() error {
	s.logger.Info("Starting sync")

	currentClients, err := s.getAdGuardClients()
	if err != nil {
		return err
	}

	leases, err := s.leases.GetLeases()
	if err != nil {
		return fmt.Errorf("getting DHCP leases: %w", err)
	}

	// Process active leases
	for mac, lease := range leases {
		if !lease.IsActive {
			continue
		}

		if existing, exists := currentClients[mac]; exists {
			if err := s.handleClientUpdate(lease, mac, existing); err != nil {
				s.logger.Error(fmt.Sprintf("Error updating client %s: %v", mac, err))
			}
			delete(currentClients, mac)
		} else {
			if err := s.handleClientAdd(lease, mac); err != nil {
				s.logger.Error(fmt.Sprintf("Error adding client %s: %v", mac, err))
			}
		}
	}

	// Remove stale clients
	for mac, client := range currentClients {
		if err := s.handleClientRemove(client, mac); err != nil {
			s.logger.Error(fmt.Sprintf("Error removing client %s: %v", mac, err))
		}
	}

	return nil
}

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
