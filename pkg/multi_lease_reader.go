// pkg/multi_lease_reader.go
package pkg

import (
	"fmt"
	"os"
)

// MultiLeaseReader combines multiple lease readers into one
type MultiLeaseReader struct {
	readers []LeaseReader
	logger  Logger
	debug   bool
}

// NewMultiLeaseReader creates a new multi-lease reader
func NewMultiLeaseReader(readers []LeaseReader, logger Logger, debug bool) *MultiLeaseReader {
	return &MultiLeaseReader{
		readers: readers,
		logger:  logger,
		debug:   debug,
	}
}

// Path returns a comma-separated list of paths
func (m *MultiLeaseReader) Path() string {
	paths := make([]string, 0, len(m.readers))
	for _, reader := range m.readers {
		paths = append(paths, reader.Path())
	}
	return fmt.Sprintf("Multiple paths: %v", paths)
}

// GetLeases reads leases from all configured sources and merges them
func (m *MultiLeaseReader) GetLeases() (map[string]ISCDHCPLease, error) {
	allLeases := make(map[string]ISCDHCPLease)

	for _, reader := range m.readers {
		// Skip readers with non-existent files
		if _, err := os.Stat(reader.Path()); os.IsNotExist(err) {
			if m.debug {
				m.logger.Info(fmt.Sprintf("Lease file not found, skipping: %s", reader.Path()))
			}
			continue
		}

		leases, err := reader.GetLeases()
		if err != nil {
			m.logger.Error(fmt.Sprintf("Error reading leases from %s: %v", reader.Path(), err))
			continue
		}

		// Merge leases, newer leases (from later readers) will overwrite older ones
		for mac, lease := range leases {
			if m.debug {
				m.logger.Info(fmt.Sprintf("Found lease for %s: %s (%s) from %s",
					mac, lease.IP, lease.Hostname, reader.Path()))
			}
			allLeases[mac] = lease
		}
	}

	if m.debug {
		m.logger.Info(fmt.Sprintf("Combined %d leases from all sources", len(allLeases)))
	}

	return allLeases, nil
}
