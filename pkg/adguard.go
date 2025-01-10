package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// AdGuard represents the AdGuard Home API client
type AdGuard struct {
	baseURL string
}

// AdGuardClient represents a client in AdGuard Home
type AdGuardClient struct {
	Name              string   `json:"name"`
	IDs               []string `json:"ids"` // MAC addresses
	IP                string   `json:"ip"`
	UseGlobalSettings bool     `json:"use_global_settings"`
}

func NewAdGuard(baseURL string) *AdGuard {
	return &AdGuard{baseURL: baseURL}
}

func (c *AdGuard) GetClients() ([]AdGuardClient, error) {
	resp, err := http.Get(fmt.Sprintf("%s/control/clients", c.baseURL))
	if err != nil {
		return nil, fmt.Errorf("getting clients: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get clients: status %d", resp.StatusCode)
	}

	var clients []AdGuardClient
	if err := json.NewDecoder(resp.Body).Decode(&clients); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return clients, nil
}

func (c *AdGuard) UpdateClient(name, ip, mac string) error {
	client := AdGuardClient{
		Name:              name,
		IDs:               []string{mac},
		IP:                ip,
		UseGlobalSettings: true,
	}

	jsonData, err := json.Marshal(client)
	if err != nil {
		return fmt.Errorf("marshal JSON: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/control/clients", c.baseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update client: status %d", resp.StatusCode)
	}

	return nil
}

func (c *AdGuard) RemoveClient(clientID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/control/clients/%s", c.baseURL, clientID), nil)
	if err != nil {
		return fmt.Errorf("creating delete request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("sending delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to remove client: status %d", resp.StatusCode)
	}

	return nil
}

func (c *AdGuard) FindClientByMAC(mac string) (*AdGuardClient, error) {
	clients, err := c.GetClients()
	if err != nil {
		return nil, fmt.Errorf("getting clients: %w", err)
	}

	for _, client := range clients {
		for _, id := range client.IDs {
			if id == mac {
				clientCopy := client // Create copy to avoid pointer to loop variable
				return &clientCopy, nil
			}
		}
	}

	return nil, nil // Not found but not an error
}
