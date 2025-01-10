package adguard

import (
	"encoding/json"
	"fmt"
)

type Client struct {
	BaseURL string
}

type AdGuardClient struct {
	Name              string   `json:"name"`
	IDs               []string `json:"ids"`
	IP                string   `json:"ip"`
	UseGlobalSettings bool     `json:"use_global_settings"`
}

func New(baseURL string) *Client {
	return &Client{BaseURL: baseURL}
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

	// [Previous update logic]
	return nil
}

func (c *Client) RemoveClient(clientID string) error {
	// [Previous remove logic]
	return nil
}
