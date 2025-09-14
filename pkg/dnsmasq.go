// pkg/dnsmasq.go
package pkg

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// DNSMasq represents the DNSMasq lease file reader
type DNSMasq struct {
	path string
}

// NewDNSMasq creates a new DNSMasq lease reader
func NewDNSMasq(path string) *DNSMasq {
	return &DNSMasq{path: path}
}

// Path returns the path to the lease file
func (d *DNSMasq) Path() string {
	return d.path
}

// GetLeases reads the DNSMasq lease file and returns a map of MAC addresses to lease information
// DNSMasq lease format:
// <expiry timestamp> <MAC address> <IP address> <hostname> <client identifier>
func (d *DNSMasq) GetLeases() (map[string]ISCDHCPLease, error) {
	file, err := os.Open(d.path)
	if err != nil {
		return nil, fmt.Errorf("opening DNSMasq lease file: %w", err)
	}
	defer file.Close()

	leases := make(map[string]ISCDHCPLease)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue // Skip empty lines and comments
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue // Skip invalid lines
		}

		// DNSMasq lease format: <expiry timestamp> <MAC address> <IP address> <hostname> <client identifier>
		mac := parts[1]
		ip := parts[2]

		// Hostname might be "*" or an actual hostname
		hostname := ""
		if len(parts) > 3 && parts[3] != "*" {
			hostname = parts[3]
		}

		// Create a lease entry that matches the ISCDHCPLease format
		lease := ISCDHCPLease{
			IP:       ip,
			MAC:      mac,
			Hostname: hostname,
			IsActive: true, // Assume all leases in the file are active
		}

		leases[mac] = lease
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning DNSMasq lease file: %w", err)
	}

	return leases, nil
}
