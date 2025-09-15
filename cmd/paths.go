// cmd/paths.go
package cmd

const (
	// FreeBSD standard paths
	InstallPath = "/usr/local/bin/dhcp-adguard-sync"
	ConfigPath  = "/usr/local/etc/dhcp-adguard-sync/config.yaml"
	RCPath      = "/usr/local/etc/rc.d/dhcp-adguard-sync"

	// OPNsenseBasePath is the base path for OPNsense files
	OPNsenseBasePath = "/usr/local/opnsense"

	// MenuPath is the path to the OPNsense menu file
	MenuPath = "/usr/local/opnsense/mvc/app/models/OPNsense/DHCPAdGuardSync/Menu/Menu.xml"

	// ACLPath is the path to the OPNsense ACL file
	ACLPath = "/usr/local/opnsense/mvc/app/models/OPNsense/DHCPAdGuardSync/ACL/ACL.xml"
)
