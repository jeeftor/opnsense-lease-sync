// pkg/lease_reader.go
package pkg

import "strings"

// LeaseReader defines the interface for reading DHCP lease files
type LeaseReader interface {
	// Path returns the path to the lease file
	Path() string

	// GetLeases reads the lease file and returns a map of MAC addresses to lease information
	GetLeases() (map[string]ISCDHCPLease, error)
}

// DetectLeaseFileFormat examines a file path and returns the appropriate lease reader
func DetectLeaseFileFormat(path string) LeaseReader {
	// If the path contains "dnsmasq", use the DNSMasq reader
	if contains(path, "dnsmasq") {
		return NewDNSMasq(path)
	}

	// Default to ISC DHCP format
	return NewDHCP(path)
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
