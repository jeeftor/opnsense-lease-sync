// cmd/paths.go
package cmd

const (
	// FreeBSD standard paths
	InstallPath = "/usr/local/bin/dhcpsync"
	ConfigPath  = "/usr/local/etc/dhcpsync/config.env"
	RCPath      = "/usr/local/etc/rc.d/dhcpsync"
	SyslogPath  = "/usr/local/etc/syslog.d/dhcpsync.conf"

	// OPNsenseBasePath is the base path for OPNsense files
	OPNsenseBasePath = "/usr/local/opnsense"

	// MenuPath is the path to the OPNsense menu file
	MenuPath = "/usr/local/opnsense/mvc/app/models/OPNsense/Dhcpsync/Menu/Menu.xml"

	// ACLPath is the path to the OPNsense ACL file
	ACLPath = "/usr/local/opnsense/mvc/app/models/OPNsense/Dhcpsync/ACL/ACL.xml"

	// ModelXMLPath is the path to the OPNsense model XML file
	ModelXMLPath = "/usr/local/opnsense/mvc/app/models/OPNsense/Dhcpsync/Dhcpsync.xml"

	// ModelPHPPath is the path to the OPNsense model PHP file
	ModelPHPPath = "/usr/local/opnsense/mvc/app/models/OPNsense/Dhcpsync/Dhcpsync.php"

	// SettingsControllerPath is the path to the OPNsense settings controller
	SettingsControllerPath = "/usr/local/opnsense/mvc/app/controllers/OPNsense/Dhcpsync/Api/SettingsController.php"

	// ServiceControllerPath is the path to the OPNsense service controller
	ServiceControllerPath = "/usr/local/opnsense/mvc/app/controllers/OPNsense/Dhcpsync/Api/ServiceController.php"

	// ViewPath is the path to the OPNsense view file
	ViewPath = "/usr/local/opnsense/mvc/app/views/OPNsense/Dhcpsync/index.volt"

	// FormPath is the path to the OPNsense form file
	FormPath = "/usr/local/opnsense/mvc/app/controllers/OPNsense/Dhcpsync/forms/dialogSettings.xml"
)
