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

	// ModelXMLPath is the path to the OPNsense model XML file
	ModelXMLPath = "/usr/local/opnsense/mvc/app/models/OPNsense/DHCPAdGuardSync/DHCPAdGuardSync.xml"

	// ModelPHPPath is the path to the OPNsense model PHP file
	ModelPHPPath = "/usr/local/opnsense/mvc/app/models/OPNsense/DHCPAdGuardSync/DHCPAdGuardSync.php"

	// SettingsControllerPath is the path to the OPNsense settings controller
	SettingsControllerPath = "/usr/local/opnsense/mvc/app/controllers/OPNsense/DHCPAdGuardSync/Api/SettingsController.php"

	// ServiceControllerPath is the path to the OPNsense service controller
	ServiceControllerPath = "/usr/local/opnsense/mvc/app/controllers/OPNsense/DHCPAdGuardSync/Api/ServiceController.php"

	// ViewPath is the path to the OPNsense view file
	ViewPath = "/usr/local/opnsense/mvc/app/views/OPNsense/DHCPAdGuardSync/index.volt"

	// FormPath is the path to the OPNsense form file
	FormPath = "/usr/local/opnsense/mvc/app/controllers/OPNsense/DHCPAdGuardSync/forms/dialogSettings.xml"
)
