package dhcp

import (
	"bufio"
	"os"
	"strings"
)

type Lease struct {
	IP       string
	Hostname string
	MAC      string
	IsActive bool
}

func ParseLeaseFile(path string) (map[string]Lease, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	leases := make(map[string]Lease)
	scanner := bufio.NewScanner(file)

	var currentLease Lease
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
				leases[currentLease.IP] = currentLease
			}
			currentLease = Lease{} // Reset for next lease
		} else if inLeaseBlock {
			// Parse details within the lease block
			if strings.HasPrefix(line, "starts") {
				currentLease.IsActive = true // Simplified, in real scenario you'd check if it's in the future
			} else if strings.HasPrefix(line, "hardware ethernet") {
				parts := strings.Fields(line)
				if len(parts) > 2 {
					currentLease.MAC = parts[2]
				}
			} else if strings.HasPrefix(line, "client-hostname") {
				parts := strings.Fields(line)
				if len(parts) > 1 {
					currentLease.Hostname = strings.Trim(parts[1], "\"")
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return leases, nil
}
