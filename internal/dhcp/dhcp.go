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
	var currentMAC string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// [Previous lease parsing logic]
	}

	return leases, nil
}
