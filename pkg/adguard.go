// pkg/adguard.go
package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type AdGuard struct {
	BaseURL    string
	authHeader string
}

type AdGuardClient struct {
	Name                string   `json:"name"`
	IDs                 []string `json:"ids"`
	IP                  string   `json:"ip"`
	UseGlobalSettings   bool     `json:"use_global_settings"`
	FilteringEnabled    bool     `json:"filtering_enabled"`
	ParentalEnabled     bool     `json:"parental_enabled"`
	SafebrowsingEnabled bool     `json:"safebrowsing_enabled"`
	SafeSearch          bool     `json:"safe_search"`
}

type AdGuardResponse struct {
	Clients     []AdGuardClient `json:"clients"`
	AutoClients []AutoClient    `json:"auto_clients"`
}

type AutoClient struct {
	WhoisInfo map[string]string `json:"whois_info"`
	IP        string            `json:"ip"`
	Name      string            `json:"name"`
	Source    string            `json:"source"`
}

func NewAdGuard(baseURL string, b64auth string) *AdGuard {
	return &AdGuard{
		BaseURL:    baseURL,
		authHeader: fmt.Sprintf("Basic %s", b64auth),
	}
}

func (a *AdGuard) makeRequest(method, path string, body interface{}) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", a.BaseURL, path)

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if a.authHeader != "" {
		req.Header.Add("Authorization", a.authHeader)
	}
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

func (a *AdGuard) GetClients() ([]AdGuardClient, error) {
	resp, err := a.makeRequest(http.MethodGet, "/control/clients", nil)
	if err != nil {
		return nil, fmt.Errorf("getting clients: %w", err)
	}
	defer resp.Body.Close()

	var response AdGuardResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	// If clients is nil, initialize empty slice
	if response.Clients == nil {
		response.Clients = []AdGuardClient{}
	}

	return response.Clients, nil
}

func (a *AdGuard) UpdateClient(name, ip, mac string, existingClient *AdGuardClient) error {
	client := AdGuardClient{
		Name: name,
		IDs:  []string{mac},
		IP:   ip,
		// Preserve existing settings
		UseGlobalSettings:   existingClient.UseGlobalSettings,
		FilteringEnabled:    existingClient.FilteringEnabled,
		ParentalEnabled:     existingClient.ParentalEnabled,
		SafebrowsingEnabled: existingClient.SafebrowsingEnabled,
		SafeSearch:          existingClient.SafeSearch,
	}

	_, err := a.makeRequest(http.MethodPut, "/control/clients/update", client)
	if err != nil {
		return fmt.Errorf("updating client: %w", err)
	}
	return nil
}

func (a *AdGuard) AddClient(name, ip, mac string) error {
	client := AdGuardClient{
		Name:                name,
		IDs:                 []string{mac},
		IP:                  ip,
		UseGlobalSettings:   true,
		FilteringEnabled:    true,
		ParentalEnabled:     false,
		SafebrowsingEnabled: false,
		SafeSearch:          false,
	}

	_, err := a.makeRequest(http.MethodPost, "/control/clients/add", client)
	if err != nil {
		return fmt.Errorf("adding client: %w", err)
	}
	return nil
}

func (a *AdGuard) RemoveClient(mac string) error {
	params := url.Values{}
	params.Add("mac", mac)

	_, err := a.makeRequest(http.MethodDelete, "/control/clients/delete?"+params.Encode(), nil)
	if err != nil {
		return fmt.Errorf("removing client: %w", err)
	}
	return nil
}
