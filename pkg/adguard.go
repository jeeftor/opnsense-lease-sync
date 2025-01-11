// pkg/adguard.go
package pkg

import (
	"fmt"
	"github.com/gmichels/adguard-client-go"
	"strings"
)

type AdGuard struct {
	client *adguard.ADG
}

func NewAdGuard(cfg Config) (*AdGuard, error) {
	// Set defaults if not provided
	scheme := "http" // Since AdGuard is running locally on OPNsense
	if cfg.Scheme != "" {
		scheme = cfg.Scheme
	}

	timeout := 10
	if cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}

	// Extract host from URL (remove scheme if present)
	host := cfg.AdGuardURL
	host = strings.TrimPrefix(host, "http://")
	host = strings.TrimPrefix(host, "https://")

	// Create new AdGuard client
	client, err := adguard.NewClient(
		&host,
		&cfg.Username,
		&cfg.Password,
		&scheme,
		&timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("creating AdGuard client: %w", err)
	}

	return &AdGuard{
		client: client,
	}, nil

}

// GetClients retrieves all clients from AdGuard Home
func (a *AdGuard) GetClients() ([]adguard.Client, error) {
	allClients, err := a.client.GetAllClients()
	if err != nil {
		return nil, fmt.Errorf("getting clients: %w", err)
	}
	return allClients.Clients, nil
}

// GetClientByMAC finds a client by MAC address from the clients list
func (a *AdGuard) GetClientByMAC(mac string) (*adguard.Client, error) {
	allClients, err := a.client.GetAllClients()
	if err != nil {
		return nil, err
	}

	for _, client := range allClients.Clients {
		for _, id := range client.Ids {
			if id == mac {
				return &client, nil
			}
		}
	}
	return nil, nil
}

// AddClient creates a new client in AdGuard Home
func (a *AdGuard) AddClient(name, mac, ip string) error {
	client := adguard.Client{
		Name: name,
		Ids:  []string{mac, ip},
		// Set sensible defaults for AdGuard Home client
		UseGlobalSettings:        true,
		UseGlobalBlockedServices: true,
		FilteringEnabled:         true,
		ParentalEnabled:          false,
		SafebrowsingEnabled:      false,

		SafeSearch: adguard.SafeSearchConfig{
			Enabled: false,
		},
	}

	_, err := a.client.CreateClient(client)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}
	return nil
}

// UpdateClient updates an existing client in AdGuard Home
func (a *AdGuard) UpdateClient(name, mac, ip string) error {
	clientUpdate := adguard.ClientUpdate{
		Name: mac, // Use MAC as identifier
		Data: adguard.Client{
			Name:                     name,
			Ids:                      []string{mac, ip},
			UseGlobalSettings:        true,
			UseGlobalBlockedServices: true,
			FilteringEnabled:         true,
			ParentalEnabled:          false,
			SafebrowsingEnabled:      false,
			SafeSearch: adguard.SafeSearchConfig{
				Enabled: false,
			},
		},
	}

	_, err := a.client.UpdateClient(clientUpdate)
	if err != nil {
		return fmt.Errorf("updating client: %w", err)
	}
	return nil
}

// RemoveClient removes a client from AdGuard Home
func (a *AdGuard) RemoveClient(name string) error {
	clientDelete := adguard.ClientDelete{
		Name: name,
	}

	err := a.client.DeleteClient(clientDelete)
	if err != nil {
		return fmt.Errorf("deleting client: %w", err)
	}
	return nil
}
