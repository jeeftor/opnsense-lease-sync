package pkg

// Use the ndp -an command to generate a list of
import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// GetNDPTable executes 'ndp -rn' command and returns a map of MAC addresses to IPv6 addresses
func GetNDPTable() (map[string][]string, error) {
	fmt.Println("Getting NDP table")

	// Initialize the result map
	ndpTable := make(map[string][]string)

	// Execute the ndp -an command (note: changed from -rn to -an to match actual output)
	cmd := exec.Command("ndp", "-an")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		return nil, err
	}

	// Process the output line by line
	scanner := bufio.NewScanner(&out)
	// Skip the header line
	if scanner.Scan() {
		_ = scanner.Text()
	}

	// Process each line
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)

		// Format from output:
		// Neighbor                             Linklayer Address  Netif Expire    1s 5s
		if len(fields) >= 3 {
			ipv6 := fields[0]
			mac := fields[1]
			//iface := fields[2]

			// Convert MAC address to uppercase for consistency
			mac = strings.ToUpper(mac)

			// Create a structured entry with both IPv6 and interface
			//entry := fmt.Sprintf("%s (via %s)", ipv6, iface)
			entry := ipv6
			ndpTable[mac] = append(ndpTable[mac], entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return ndpTable, nil
}

// GetIP6forMAC returns an array of IPv6 addresses associated with the given MAC address.
// Returns empty array if no IPv6 addresses are found or if there's an error getting the NDP table.
func GetIP6forMAC(mac string) ([]string, error) {
	// Normalize the MAC address to uppercase for consistency
	mac = strings.ToUpper(mac)

	// Get the NDP table
	ndpTable, err := GetNDPTable()
	if err != nil {
		return nil, err
	}

	// Return the IPv6 addresses for this MAC, or empty array if none found
	if ips, exists := ndpTable[mac]; exists {
		return ips, nil
	}
	return []string{}, nil
}
