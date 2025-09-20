package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"text/template"

	"github.com/spf13/cobra"
	"golang.org/x/term"
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

// detectDHCPService attempts to detect which DHCP service is running and returns appropriate config
func detectDHCPService() (leasePath, leaseFormat string) {
	// Check for DNSMasq (OPNsense default)
	if _, err := os.Stat("/var/db/dnsmasq.leases"); err == nil {
		return "/var/db/dnsmasq.leases", "dnsmasq"
	}

	// Check for ISC DHCP v4
	if _, err := os.Stat("/var/dhcpd/var/db/dhcpd.leases"); err == nil {
		return "/var/dhcpd/var/db/dhcpd.leases", "isc"
	}

	// Fallback to DNSMasq (OPNsense default)
	return "/var/db/dnsmasq.leases", "dnsmasq"
}

// promptForInput prompts user for input with a message
func promptForInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

// promptForPassword prompts for password without echoing
func promptForPassword(prompt string) string {
	fmt.Print(prompt)
	bytePassword, _ := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return string(bytePassword)
}

var installCmd = &cobra.Command{
	Use:          "install",
	Short:        "Install the service and configuration files",
	SilenceUsage: true,
	Long: `Install the dhcpsync binary, configuration, and service files.
This will:
1. Copy the binary to /usr/local/bin
2. Create a config file in /usr/local/etc/dhcpsync with provided credentials
3. Install the rc.d service script
4. Create log directory for dhcpsync.log
5. Set appropriate permissions

Use --dry-run to preview what would be written without making any changes.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		installDryRun, _ := cmd.Flags().GetBool("dry-run")

		// Check if running on FreeBSD
		if !installDryRun && runtime.GOOS != "freebsd" {
			return fmt.Errorf("installation is only supported on FreeBSD systems")
		}

		// Get flags for this command
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		adguardURL, _ := cmd.Flags().GetString("adguard-url")
		leasePath, _ := cmd.Flags().GetString("lease-path")
		leaseFormat, _ := cmd.Flags().GetString("lease-format")
		scheme, _ := cmd.Flags().GetString("scheme")
		timeout, _ := cmd.Flags().GetInt("timeout")
		preserveDeletedHosts, _ := cmd.Flags().GetBool("preserve-deleted-hosts")
		logLevel, _ := cmd.Flags().GetString("log-level")
		logFile, _ := cmd.Flags().GetString("log-file")
		maxLogSize, _ := cmd.Flags().GetInt("max-log-size")
		maxBackups, _ := cmd.Flags().GetInt("max-backups")
		maxAge, _ := cmd.Flags().GetInt("max-age")
		noCompress, _ := cmd.Flags().GetBool("no-compress")

		// Prompt for username if not provided
		if username == "" {
			username = promptForInput("AdGuard Home Username: ")
			if username == "" {
				return fmt.Errorf("username is required")
			}
		}

		// Prompt for password if not provided
		if password == "" {
			password = promptForPassword("AdGuard Home Password: ")
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
			fmt.Printf("Would copy binary to: %s\n", "/usr/local/bin/dhcpsync")
			fmt.Printf("Would set permissions: 0755\n")
		} else {
			// Create installation directory and copy binary
			if err := os.MkdirAll(filepath.Dir("/usr/local/bin/dhcpsync"), 0755); err != nil {
				return fmt.Errorf("failed to create installation directory: %w", err)
			}
			if executable != "/usr/local/bin/dhcpsync" {
				if err := copyFile(executable, "/usr/local/bin/dhcpsync"); err != nil {
					return fmt.Errorf("failed to copy binary: %w", err)
				}
			}
			if err := os.Chmod("/usr/local/bin/dhcpsync", 0755); err != nil {
				return fmt.Errorf("failed to set binary permissions: %w", err)
			}
		}

		// Auto-detect DHCP service if not overridden by flags
		detectedLeasePath, detectedLeaseFormat := detectDHCPService()

		// Use detected values if flags weren't explicitly set
		finalLeasePath := leasePath
		finalLeaseFormat := leaseFormat
		if !cmd.Flags().Changed("lease-path") {
			finalLeasePath = detectedLeasePath
		}
		if !cmd.Flags().Changed("lease-format") {
			finalLeaseFormat = detectedLeaseFormat
		}

		if !installDryRun {
			fmt.Printf("Auto-detected DHCP service: %s (lease file: %s)\n", finalLeaseFormat, finalLeasePath)
		}

		// Generate config from template
		configTemplate := ConfigTemplate{
			AdGuardURL:           adguardURL,
			Username:             username,
			Password:             password,
			Scheme:               scheme,
			Timeout:              timeout,
			LeasePath:            finalLeasePath,
			LeaseFormat:          finalLeaseFormat,
			PreserveDeletedHosts: preserveDeletedHosts,
			Debug:                logLevel == "debug",
			DryRun:               installDryRun,
			LogLevel:             logLevel,
			LogFile:              logFile,
			MaxLogSize:           maxLogSize,
			MaxBackups:           maxBackups,
			MaxAge:               maxAge,
			NoCompress:           noCompress,
		}

		// Read template content
		// Parse and execute template from the embedded filesystem
		tmpl, err := template.New("config.env").ParseFS(templates, "templates/config.env")
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
			fmt.Printf("Would write to: %s\n", "/usr/local/etc/dhcpsync/config.env")
			fmt.Printf("Would set permissions: 0600\n")
			fmt.Println("Content would be:")
			fmt.Println("---")
			fmt.Println(configBuffer.String())
			fmt.Println("---")
		} else {
			// Create config directory and check for existing config
			configDir := filepath.Dir("/usr/local/etc/dhcpsync/config.env")
			if err := os.MkdirAll(configDir, 0755); err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}

			// Check if config file already exists
			configExists := false
			if _, err := os.Stat("/usr/local/etc/dhcpsync/config.env"); err == nil {
				configExists = true
				fmt.Printf("\nExisting configuration found at %s\n", "/usr/local/etc/dhcpsync/config.env")
				fmt.Println("Preserving existing configuration")
			} else if !os.IsNotExist(err) {
				return fmt.Errorf("failed to check for existing config: %w", err)
			}

			// Only write config if it doesn't exist
			if !configExists {
				if err := os.WriteFile("/usr/local/etc/dhcpsync/config.env", configBuffer.Bytes(), 0600); err != nil {
					return fmt.Errorf("failed to write config file: %w", err)
				}
				fmt.Printf("\nNew configuration written to %s\n", "/usr/local/etc/dhcpsync/config.env")
			}
		}

		// Read and process RC script from the embedded filesystem
		rcContent, err := templates.ReadFile("templates/rc.script")
		if err != nil {
			return fmt.Errorf("failed to read rc.script template: %w", err)
		}

		// Dry run: Print RC script info
		if installDryRun {
			fmt.Println("\n=== RC Script (Dry Run) ===")
			fmt.Printf("Would write to: %s\n", "/usr/local/etc/rc.d/dhcpsync")
			fmt.Printf("Would set permissions: 0755\n")
			fmt.Println("Content would be:")
			fmt.Println("---")
			fmt.Println(string(rcContent))
			fmt.Println("---")
			fmt.Println("\nService Installation (Dry Run):")
			fmt.Println("Would run: service dhcpsync enable")
		} else {
			// Write RC script
			if err := os.WriteFile("/usr/local/etc/rc.d/dhcpsync", rcContent, 0755); err != nil {
				return fmt.Errorf("failed to create rc.d script: %w", err)
			}

			// Enable the service
			cmd := exec.Command("service", "dhcpsync", "enable")
			output, err := cmd.CombinedOutput()
			if err != nil {
				return fmt.Errorf("failed to enable service: %w\nOutput: %s", err, string(output))
			}
		}

		// Create log directory for direct file logging
		logDir := "/var/log"
		if installDryRun {
			fmt.Println("\n=== Log Directory Setup (Dry Run) ===")
			fmt.Printf("Would ensure directory exists: %s\n", logDir)
			fmt.Printf("Log file will be: %s\n", "/var/log/dhcpsync.log")
			fmt.Println("Logs will be written directly to file (no syslog configuration needed)")
		} else {
			// Ensure log directory exists (should already exist on FreeBSD/OPNsense)
			if err := os.MkdirAll(logDir, 0755); err != nil {
				return fmt.Errorf("failed to ensure log directory exists: %w", err)
			}
			fmt.Printf("Log directory ready: %s\n", logDir)
		}
		if installDryRun {
			fmt.Println("\n=== Dry Run Complete ===")
			fmt.Println("No changes were made to your system.")
		} else {
			fmt.Println("\nInstallation completed successfully!")
			// Message about configuration was already printed earlier
			fmt.Println("Start the service with: service dhcpsync start")
		}
		return nil
	},
}

func init() {
	installCmd.Flags().String("username", "", "AdGuard Home username")
	installCmd.Flags().String("password", "", "AdGuard Home password")
	installCmd.Flags().String("adguard-url", "127.0.0.1:3000", "AdGuard Home host:port")
	installCmd.Flags().String("lease-path", "", "Path to DHCP leases file (will be auto-detected if empty)")
	installCmd.Flags().String("lease-format", "", "DHCP lease file format (isc or dnsmasq, will be auto-detected if empty)")
	installCmd.Flags().String("scheme", "http", "Connection scheme (http/https)")
	installCmd.Flags().Int("timeout", 10, "API timeout in seconds")
	installCmd.Flags().Bool("dry-run", false, "Show what would be installed without making any changes")
	installCmd.Flags().Bool("preserve-deleted-hosts", false, "Preserve deleted hosts")
	installCmd.Flags().String("log-level", "info", "Log level (debug, info, warn, error, fatal)")
	installCmd.Flags().String("log-file", "/var/log/dhcpsync.log", "Log file path")
	installCmd.Flags().Int("max-log-size", 100, "Maximum log file size in megabytes")
	installCmd.Flags().Int("max-backups", 5, "Maximum number of log file backups")
	installCmd.Flags().Int("max-age", 28, "Maximum age of log file backups in days")
	installCmd.Flags().Bool("no-compress", false, "Do not compress log file backups")
}
