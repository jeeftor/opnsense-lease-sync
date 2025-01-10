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

func (s *SyncService) syncLeases(currentLeases, previousLeases map[string]dhcp.Lease) error {
	// Check for new or updated leases
	for mac, lease := range currentLeases {
		prev, exists := previousLeases[mac]
		if !exists || prev.IP != lease.IP || prev.Hostname != lease.Hostname {
			if err := s.adguardClient.UpdateClient(lease.Hostname, lease.IP, lease.MAC); err != nil {
				s.logger.Error(fmt.Sprintf("Error updating lease in AdGuard: %v", err))
			}
		}
	}

	// Check for expired leases
	for mac, lease := range previousLeases {
		if _, exists := currentLeases[mac]; !exists {
			if err := s.adguardClient.RemoveClient(lease.MAC); err != nil {
				s.logger.Error(fmt.Sprintf("Error removing lease from AdGuard: %v", err))
			}
		}
	}

	return nil
}

func (s *SyncService) processLeaseFile() (map[string]dhcp.Lease, error) {
	currentLeases, err := dhcp.ParseLeaseFile(s.leasePath)
	if err != nil {
		return nil, fmt.Errorf("parsing lease file: %w", err)
	}
	return currentLeases, nil
}

func (s *SyncService) Run() error {
	s.logger.Info("DHCP to AdGuard Home sync service starting")

	// Watch the directory containing the lease file
	leaseDir := filepath.Dir(s.leasePath)
	if err := s.watcher.Add(leaseDir); err != nil {
		return fmt.Errorf("watching lease directory: %w", err)
	}

	// Initial read of leases
	previousLeases, err := s.processLeaseFile()
	if err != nil {
		s.logger.Error(fmt.Sprintf("Initial lease file read failed: %v", err))
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
					s.logger.Info("Lease file changed, processing updates")

					currentLeases, err := s.processLeaseFile()
					if err != nil {
						s.logger.Error(fmt.Sprintf("Failed to process lease file: %v", err))
						return
					}

					if err := s.syncLeases(currentLeases, previousLeases); err != nil {
						s.logger.Error(fmt.Sprintf("Failed to sync leases: %v", err))
						return
					}

					previousLeases = currentLeases
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
