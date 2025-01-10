package adguard

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Client struct {
	BaseURL string
}

type AdGuardClient struct {
	Name              string   `json:"name"`
	IDs               []string `json:"ids"` // MAC addresses
	IP                string   `json:"ip"`
	UseGlobalSettings bool     `json:"use_global_settings"`
}

func New(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
}

func (c *Client) GetClients() ([]AdGuardClient, error) {
	resp, err := http.Get(fmt.Sprintf("%s/control/clients", c.BaseURL))
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

func (c *Client) UpdateClient(name, ip, mac string) error {
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

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/control/clients", c.BaseURL), bytes.NewBuffer(jsonData))
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

func (c *Client) RemoveClient(clientID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/control/clients/%s", c.BaseURL, clientID), nil)
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

func (c *Client) FindClientByMAC(mac string) (*AdGuardClient, error) {
	clients, err := c.GetClients()
	if err != nil {
		return nil, fmt.Errorf("getting clients: %w", err)
	}

	for _, client := range clients {
		for _, id := range client.IDs {
			if id == mac {
				return &client, nil
			}
		}
	}

	return nil, nil // Not found but not an error
}
