package cmd

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"text/template"
)

// ConfigTemplate represents the structure for config file template
type ConfigTemplate struct {
	AdGuardURL           string
	Username             string
	Password             string
	Scheme               string
	Timeout              int
	LeasePath            string
	LeaseFormat          string
	PreserveDeletedHosts bool
	Debug                bool
	DryRun               bool
	LogLevel             string
	LogFile              string
	MaxLogSize           int
	MaxBackups           int
	MaxAge               int
	NoCompress           bool
}

// copyFile copies a file from src to ds
func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	return os.WriteFile(dst, input, 0644)
}

// installDryRun flag for dry run mode
var installDryRun bool

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the service and configuration files",
	Long: `Install the dhcp-adguard-sync binary, configuration, and service files.
This will:
1. Copy the binary to /usr/local/bin
2. Create a config file in /usr/local/etc/dhcp-adguard-sync with provided credentials
3. Install the rc.d service scrip
4. Set appropriate permissions

Use --dry-run to preview what would be written without making any changes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if running on FreeBSD
		if !installDryRun && runtime.GOOS != "freebsd" {
			return fmt.Errorf("installation is only supported on FreeBSD systems")
		}

		// Get the path of the current binary
		executable, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %w", err)
		}

		// Dry run: Print binary info
		if installDryRun {
			fmt.Println("\n=== Binary Installation (Dry Run) ===")
			fmt.Printf("Would copy binary from: %s\n", executable)
			fmt.Printf("Would copy binary to: %s\n", InstallPath)
			fmt.Printf("Would set permissions: 0755\n")
		} else {
			// Create installation directory and copy binary
			if err := os.MkdirAll(filepath.Dir(InstallPath), 0755); err != nil {
				return fmt.Errorf("failed to create installation directory: %w", err)
			}
			if executable != InstallPath {
				if err := copyFile(executable, InstallPath); err != nil {
					return fmt.Errorf("failed to copy binary: %w", err)
				}
			}
			if err := os.Chmod(InstallPath, 0755); err != nil {
				return fmt.Errorf("failed to set binary permissions: %w", err)
			}
		}

		// Generate config from template
		configTemplate := ConfigTemplate{
			AdGuardURL:           adguardURL,
			Username:             username,
			Password:             password,
			Scheme:               scheme,
			Timeout:              timeout,
			LeasePath:            leasePath,
			LeaseFormat:          leaseFormat,
			PreserveDeletedHosts: preserveDeletedHosts,
			Debug:                debug, // Added Debug field
			DryRun:               dryRun,
			LogLevel:             logLevel,
			LogFile:              logFile,
			MaxLogSize:           maxLogSize,
			MaxBackups:           maxBackups,
			MaxAge:               maxAge,
			NoCompress:           noCompress,
		}

		// Read template conten
		templateContent, err := templates.ReadFile("templates/config.yaml")
		if err != nil {
			return fmt.Errorf("failed to read config template: %w", err)
		}

		// Parse and execute template
		tmpl, err := template.New("config").Parse(string(templateContent))
		if err != nil {
			return fmt.Errorf("failed to parse config template: %w", err)
		}

		var configBuffer bytes.Buffer
		if err := tmpl.Execute(&configBuffer, configTemplate); err != nil {
			return fmt.Errorf("failed to generate config: %w", err)
		}

		// Dry run: Print config info
		if installDryRun {
			fmt.Println("\n=== Configuration File (Dry Run) ===")
			fmt.Printf("Would write to: %s\n", ConfigPath)
			fmt.Printf("Would set permissions: 0600\n")
			fmt.Println("Content would be:")
			fmt.Println("---")
			fmt.Println(configBuffer.String())
			fmt.Println("---")
		} else {
			// Create config directory and check for existing config
			configDir := filepath.Dir(ConfigPath)
			if err := os.MkdirAll(configDir, 0755); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}

			// Check if config file already exists
			configExists := false
			if _, err := os.Stat(ConfigPath); err == nil {
				configExists = true
				fmt.Printf("\nExisting configuration found at %s\n", ConfigPath)
				fmt.Println("Preserving existing configuration")
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("failed to check for existing config: %w", err)
			}

			// Only write config if it doesn't exis
			if !configExists {
				if err := os.WriteFile(ConfigPath, configBuffer.Bytes(), 0600); err != nil {
					return fmt.Errorf("failed to write config file: %w", err)
				}
				fmt.Printf("\nNew configuration written to %s\n", ConfigPath)
			}
		}

		// Read and process RC scrip
		rcContent, err := templates.ReadFile("templates/rc.script")
		if err != nil {
			return fmt.Errorf("failed to read rc.script template: %w", err)
		}

		// Dry run: Print RC script info
		if installDryRun {
			fmt.Println("\n=== RC Script (Dry Run) ===")
			fmt.Printf("Would write to: %s\n", RCPath)
			fmt.Printf("Would set permissions: 0755\n")
			fmt.Println("Content would be:")
			fmt.Println("---")
			fmt.Println(string(rcContent))
			fmt.Println("---")
			fmt.Println("\nService Installation (Dry Run):")
			fmt.Println("Would run: service dhcp-adguard-sync enable")
		} else {
			// Write RC scrip
			if err := os.WriteFile(RCPath, rcContent, 0755); err != nil {
				return fmt.Errorf("failed to create rc.d script: %w", err)
			}

			// Create OPNsense menu directory structure
			menuDir := filepath.Dir(MenuPath)
			if err := os.MkdirAll(menuDir, 0755); err != nil {
				return fmt.Errorf("failed to create menu directory: %w", err)
			}

			// Copy Menu.xml to OPNsense directory
			menuContent, err := templates.ReadFile("templates/Menu.xml")
			if err != nil {
				return fmt.Errorf("failed to read Menu.xml template: %w", err)
			}

			if err := os.WriteFile(MenuPath, menuContent, 0644); err != nil {
				return fmt.Errorf("failed to write Menu.xml: %w", err)
			}
			fmt.Printf("Menu file installed at %s\n", MenuPath)

			// Create ACL directory and copy ACL.xml
			aclDir := filepath.Dir(ACLPath)
			if err := os.MkdirAll(aclDir, 0755); err != nil {
				return fmt.Errorf("failed to create ACL directory: %w", err)
			}

			// Copy ACL.xml to OPNsense directory
			aclContent, err := templates.ReadFile("templates/ACL.xml")
			if err != nil {
				return fmt.Errorf("failed to read ACL.xml template: %w", err)
			}

			if err := os.WriteFile(ACLPath, aclContent, 0644); err != nil {
				return fmt.Errorf("failed to write ACL.xml: %w", err)
			}
			fmt.Printf("ACL file installed at %s\n", ACLPath)

			// Create model directory and copy model files
			modelDir := filepath.Dir(ModelXMLPath)
			if err := os.MkdirAll(modelDir, 0755); err != nil {
				return fmt.Errorf("failed to create model directory: %w", err)
			}

			// Copy model XML file
			modelXMLContent, err := templates.ReadFile("templates/DHCPAdGuardSync.xml")
			if err != nil {
				return fmt.Errorf("failed to read DHCPAdGuardSync.xml template: %w", err)
			}

			if err := os.WriteFile(ModelXMLPath, modelXMLContent, 0644); err != nil {
				return fmt.Errorf("failed to write DHCPAdGuardSync.xml: %w", err)
			}
			fmt.Printf("Model XML file installed at %s\n", ModelXMLPath)

			// Copy model PHP file
			modelPHPContent, err := templates.ReadFile("templates/DHCPAdGuardSync.php")
			if err != nil {
				return fmt.Errorf("failed to read DHCPAdGuardSync.php template: %w", err)
			}

			if err := os.WriteFile(ModelPHPPath, modelPHPContent, 0644); err != nil {
				return fmt.Errorf("failed to write DHCPAdGuardSync.php: %w", err)
			}
			fmt.Printf("Model PHP file installed at %s\n", ModelPHPPath)

			// Create API controllers directory and copy controller files
			controllersDir := filepath.Dir(SettingsControllerPath)
			if err := os.MkdirAll(controllersDir, 0755); err != nil {
				return fmt.Errorf("failed to create controllers directory: %w", err)
			}

			// Copy settings controller
			settingsContent, err := templates.ReadFile("templates/SettingsController.php")
			if err != nil {
				return fmt.Errorf("failed to read SettingsController.php template: %w", err)
			}

			if err := os.WriteFile(SettingsControllerPath, settingsContent, 0644); err != nil {
				return fmt.Errorf("failed to write SettingsController.php: %w", err)
			}
			fmt.Printf("Settings controller installed at %s\n", SettingsControllerPath)

			// Copy service controller
			serviceContent, err := templates.ReadFile("templates/ServiceController.php")
			if err != nil {
				return fmt.Errorf("failed to read ServiceController.php template: %w", err)
			}

			if err := os.WriteFile(ServiceControllerPath, serviceContent, 0644); err != nil {
				return fmt.Errorf("failed to write ServiceController.php: %w", err)
			}
			fmt.Printf("Service controller installed at %s\n", ServiceControllerPath)

			// Create view directory and copy view file
			viewDir := filepath.Dir(ViewPath)
			if err := os.MkdirAll(viewDir, 0755); err != nil {
				return fmt.Errorf("failed to create view directory: %w", err)
			}

			// Copy view file
			viewContent, err := templates.ReadFile("templates/index.volt")
			if err != nil {
				return fmt.Errorf("failed to read index.volt template: %w", err)
			}

			if err := os.WriteFile(ViewPath, viewContent, 0644); err != nil {
				return fmt.Errorf("failed to write index.volt: %w", err)
			}
			fmt.Printf("View file installed at %s\n", ViewPath)

			// Create form directory and copy form file
			formDir := filepath.Dir(FormPath)
			if err := os.MkdirAll(formDir, 0755); err != nil {
				return fmt.Errorf("failed to create form directory: %w", err)
			}

			// Copy form file
			formContent, err := templates.ReadFile("templates/dialogSettings.xml")
			if err != nil {
				return fmt.Errorf("failed to read dialogSettings.xml template: %w", err)
			}

			if err := os.WriteFile(FormPath, formContent, 0644); err != nil {
				return fmt.Errorf("failed to write dialogSettings.xml: %w", err)
			}
			fmt.Printf("Form file installed at %s\n", FormPath)

			// Enable the service
			if err := exec.Command("service", "dhcp-adguard-sync", "enable").Run(); err != nil {
				return fmt.Errorf("failed to enable service: %w", err)
			}
		}

		if installDryRun {
			fmt.Println("\n=== Dry Run Complete ===")
			fmt.Println("No changes were made to your system.")
		} else {
			fmt.Println("\nInstallation completed successfully!")
			// Message about configuration was already printed earlier
			fmt.Println("Start the service with: service dhcp-adguard-sync start")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	serveCmd.MarkFlagRequired("username")
	serveCmd.MarkFlagRequired("password")

	// Add dry-run flag
	installCmd.Flags().BoolVar(&installDryRun, "dry-run", false, "Show what would be installed without making any changes")
}
