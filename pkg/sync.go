package pkg

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type SyncService struct {
	adguard *AdGuard // Changed from *AdGuardClient to *AdGuard
	leases  *DHCP
	logger  Logger
	watcher *fsnotify.Watcher
	done    chan bool
	dryRun  bool
}

type Config struct {
	AdGuardURL string
	LeasePath  string
	DryRun     bool
	Logger     Logger
}

func NewSyncService(cfg Config) (*SyncService, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("creating file watcher: %w", err)
	}

	return &SyncService{
		adguard: NewAdGuard(cfg.AdGuardURL),
		leases:  NewDHCP(cfg.LeasePath),
		logger:  cfg.Logger,
		watcher: watcher,
		done:    make(chan bool),
		dryRun:  cfg.DryRun,
	}, nil
}

func (s *SyncService) sync() error {
	s.logger.Info("Starting sync")

	clients, err := s.adguard.GetClients()
	if err != nil {
		return fmt.Errorf("getting AdGuard clients: %w", err)
	}

	// Map for tracking current AdGuard clients by MAC
	currentClients := make(map[string]*AdGuardClient)
	for _, client := range clients {
		if len(client.IDs) > 0 {
			clientCopy := client
			currentClients[client.IDs[0]] = &clientCopy
		}
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

		if client, exists := currentClients[mac]; exists {
			// Update existing client if needed
			if client.Name != lease.Hostname || client.IP != lease.IP {
				action := fmt.Sprintf("Updating client %s (%s) with IP %s", lease.Hostname, mac, lease.IP)
				if s.dryRun {
					s.logger.Info("DRY-RUN: " + action)
				} else {
					s.logger.Info(action)
					if err := s.adguard.UpdateClient(lease.Hostname, lease.IP, mac); err != nil {
						s.logger.Error(fmt.Sprintf("Failed to update client: %v", err))
					}
				}
			}
			delete(currentClients, mac)
		} else {
			// Add new client
			action := fmt.Sprintf("Adding new client %s (%s) with IP %s", lease.Hostname, mac, lease.IP)
			if s.dryRun {
				s.logger.Info("DRY-RUN: " + action)
			} else {
				s.logger.Info(action)
				if err := s.adguard.UpdateClient(lease.Hostname, lease.IP, mac); err != nil {
					s.logger.Error(fmt.Sprintf("Failed to add client: %v", err))
				}
			}
		}
	}

	// Remove stale clients
	for mac, client := range currentClients {
		action := fmt.Sprintf("Removing client %s (%s)", client.Name, mac)
		if s.dryRun {
			s.logger.Info("DRY-RUN: " + action)
		} else {
			s.logger.Info(action)
			if err := s.adguard.RemoveClient(mac); err != nil {
				s.logger.Error(fmt.Sprintf("Failed to remove client: %v", err))
			}
		}
	}

	return nil
}

func (s *SyncService) Run() error {
	s.logger.Info("Starting DHCP to AdGuard Home sync service")

	// Initial sync
	if err := s.sync(); err != nil {
		s.logger.Error(fmt.Sprintf("Initial sync failed: %v", err))
	}

	// Watch lease file directory
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
					if err := s.sync(); err != nil {
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
