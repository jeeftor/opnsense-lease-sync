package pkg

import "github.com/fsnotify/fsnotify"

// DHCP represents the DHCP lease file reader
type DHCP struct {
	path string
}

// SyncService represents the DHCP to AdGuard sync service
type SyncService struct {
	adguard              *AdGuard
	leases               *DHCP
	logger               Logger
	watcher              *fsnotify.Watcher
	done                 chan bool
	dryRun               bool
	preserveDeletedHosts bool
	debug                bool
}

// ISCDHCPLease represents a lease from ISC DHCP server's lease file
type ISCDHCPLease struct {
	IP       string
	Hostname string
	MAC      string
	IsActive bool
}

// AdGuardDHCPLease represents a current DHCP lease from AdGuard
type AdGuardDHCPLease struct {
	IP          string `json:"ip"`
	MAC         string `json:"mac"`
	Hostname    string `json:"hostname"`
	IsStatic    bool   `json:"static"`
	Online      bool   `json:"online"`
	LastSeen    string `json:"last_seen"`
	Expiry      string `json:"expires"`
	DisplayName string `json:"display_name,omitempty"`
}

// StaticDHCPLease represents a static DHCP lease configuration
type StaticDHCPLease struct {
	MAC      string `json:"mac"`
	IP       string `json:"ip"`
	Hostname string `json:"hostname"`
}

// AdGuardDHCPResponse represents the response from AdGuard's DHCP leases endpoint
type AdGuardDHCPResponse struct {
	Leases []AdGuardDHCPLease `json:"leases"`
}

// StaticDHCPResponse represents the response from AdGuard's static leases endpoint
type StaticDHCPResponse struct {
	Leases []StaticDHCPLease `json:"static_leases"`
}
