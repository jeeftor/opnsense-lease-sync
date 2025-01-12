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
	PreserveDeletedHosts bool
	Debug                bool // Added Debug field
	DryRun               bool
	LogLevel             string
	LogFile              string
	MaxLogSize           int
	MaxBackups           int
	MaxAge               int
	NoCompress           bool
}

// copyFile copies a file from src to dst
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
3. Install the rc.d service script
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
			if err := copyFile(executable, InstallPath); err != nil {
				return fmt.Errorf("failed to copy binary: %w", err)
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

		// Read template content
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
			// Create config directory and write config
			configDir := filepath.Dir(ConfigPath)
			if err := os.MkdirAll(configDir, 0755); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}
			if err := os.WriteFile(ConfigPath, configBuffer.Bytes(), 0600); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}
		}

		// Read and process RC script
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
			// Write RC script
			if err := os.WriteFile(RCPath, rcContent, 0755); err != nil {
				return fmt.Errorf("failed to create rc.d script: %w", err)
			}

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
			fmt.Printf("Configuration has been written to %s\n", ConfigPath)
			fmt.Println("Start the service with: service dhcp-adguard-sync start")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)

	// Add dry-run flag
	installCmd.Flags().BoolVar(&installDryRun, "dry-run", false, "Show what would be installed without making any changes")
}
