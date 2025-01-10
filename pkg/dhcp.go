package pkg

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// DHCP represents the DHCP lease file reader
type DHCP struct {
	path string
}

// DHCPLease represents a DHCP lease entry
type DHCPLease struct {
	IP       string
	Hostname string
	MAC      string
	IsActive bool
}

func NewDHCP(path string) *DHCP {
	return &DHCP{path: path}
}

func (d *DHCP) Path() string {
	return d.path
}

func (d *DHCP) GetLeases() (map[string]DHCPLease, error) {
	file, err := os.Open(d.path)
	if err != nil {
		return nil, fmt.Errorf("opening lease file: %w", err)
	}
	defer file.Close()

	leases := make(map[string]DHCPLease)
	scanner := bufio.NewScanner(file)

	var currentLease DHCPLease
	var inLeaseBlock bool

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "lease") {
			// New lease entry
			inLeaseBlock = true
			parts := strings.Fields(line)
			if len(parts) > 1 {
				currentLease.IP = parts[1]
			}
		} else if strings.HasPrefix(line, "}") {
			// End of lease entry
			inLeaseBlock = false
			if currentLease.MAC != "" {
				// Store lease by MAC address for easier lookup
				leases[currentLease.MAC] = currentLease
			}
			currentLease = DHCPLease{} // Reset for next lease
		} else if inLeaseBlock {
			// Parse details within the lease block
			if strings.HasPrefix(line, "binding state active") {
				currentLease.IsActive = true
			} else if strings.HasPrefix(line, "hardware ethernet") {
				parts := strings.Fields(line)
				if len(parts) > 2 {
					currentLease.MAC = strings.TrimSuffix(parts[2], ";")
				}
			} else if strings.HasPrefix(line, "client-hostname") {
				parts := strings.Fields(line)
				if len(parts) > 1 {
					currentLease.Hostname = strings.Trim(strings.TrimSuffix(parts[1], ";"), "\"")
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning lease file: %w", err)
	}

	return leases, nil
}
