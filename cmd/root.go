// cmd/root.go
package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	username             string
	password             string
	adguardURL           string
	leasePath            string
	leaseFormat          string
	dryRun               bool
	scheme               string
	timeout              int
	preserveDeletedHosts bool
	debug                bool

	// Logging configuration
	logLevel   string
	logFile    string
	maxLogSize int
	maxBackups int
	maxAge     int
	noCompress bool
)

// validateAdGuardFlags checks AdGuard-specific flags
func validateAdGuardFlags(cmd *cobra.Command) error {
	// Skip validation for commands that don't need AdGuard credentials
	if cmd.Name() == "install" || cmd.Name() == "uninstall" || cmd.Name() == "version" {
		return nil
	}

	// Check for environment variables
	if envUser := os.Getenv("ADGUARD_USERNAME"); envUser != "" && !cmd.Flags().Changed("username") {
		username = envUser
	}
	if envPass := os.Getenv("ADGUARD_PASSWORD"); envPass != "" && !cmd.Flags().Changed("password") {
		password = envPass
	}

	// Validate required flags for AdGuard interaction
	if username == "" {
		return fmt.Errorf("username is required for %s command. Set via --username flag or ADGUARD_USERNAME environment variable", cmd.Name())
	}
	if password == "" {
		return fmt.Errorf("password is required for %s command. Set via --password flag or ADGUARD_PASSWORD environment variable", cmd.Name())
	}

	return nil
}

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "dhcp-adguard-sync",
	Short: "Sync ISC DHCP leases to AdGuard Home",
	Long: `A service/CLI tool that synchronizes ISC DHCP leases from OPNsense
to AdGuard Home, keeping client configurations in sync automatically.

Can be run either as a one-time sync (CLI mode) or as a persistent service
that watches for lease file changes.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Check for other environment variables
		if envURL := os.Getenv("ADGUARD_URL"); envURL != "" && !cmd.Flags().Changed("adguard-url") {
			adguardURL = envURL
		}
		if envLease := os.Getenv("DHCP_LEASE_PATH"); envLease != "" && !cmd.Flags().Changed("lease-path") {
			leasePath = envLease
		}
		if envLeaseFormat := os.Getenv("LEASE_FORMAT"); envLeaseFormat != "" && !cmd.Flags().Changed("lease-format") {
			leaseFormat = envLeaseFormat
		}
		if envScheme := os.Getenv("ADGUARD_SCHEME"); envScheme != "" && !cmd.Flags().Changed("scheme") {
			scheme = envScheme
		}
		if envTimeout := os.Getenv("ADGUARD_TIMEOUT"); envTimeout != "" && !cmd.Flags().Changed("timeout") {
			if t, err := strconv.Atoi(envTimeout); err == nil {
				timeout = t
			}
		}
		if envDryRun := os.Getenv("DRY_RUN"); envDryRun != "" && !cmd.Flags().Changed("dry-run") {
			dryRun = envDryRun == "true" || envDryRun == "1"
		}
		if envPreserve := os.Getenv("PRESERVE_DELETED_HOSTS"); envPreserve != "" && !cmd.Flags().Changed("preserve-deleted-hosts") {
			preserveDeletedHosts = envPreserve == "true" || envPreserve == "1"
		}

		// Check for logging environment variables
		if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" && !cmd.Flags().Changed("log-level") {
			logLevel = envLogLevel
		}
		if envLogFile := os.Getenv("LOG_FILE"); envLogFile != "" && !cmd.Flags().Changed("log-file") {
			logFile = envLogFile
		}
		if envMaxSize := os.Getenv("MAX_LOG_SIZE"); envMaxSize != "" && !cmd.Flags().Changed("max-log-size") {
			if size, err := strconv.Atoi(envMaxSize); err == nil {
				maxLogSize = size
			}
		}
		if envMaxBackups := os.Getenv("MAX_BACKUPS"); envMaxBackups != "" && !cmd.Flags().Changed("max-backups") {
			if backups, err := strconv.Atoi(envMaxBackups); err == nil {
				maxBackups = backups
			}
		}
		if envMaxAge := os.Getenv("MAX_AGE"); envMaxAge != "" && !cmd.Flags().Changed("max-age") {
			if age, err := strconv.Atoi(envMaxAge); err == nil {
				maxAge = age
			}
		}
		if envNoCompress := os.Getenv("NO_COMPRESS"); envNoCompress != "" && !cmd.Flags().Changed("no-compress") {
			noCompress = envNoCompress == "true" || envNoCompress == "1"
		}

		// Validate AdGuard flags conditionally
		if err := validateAdGuardFlags(cmd); err != nil {
			return err
		}

		// Always validate other settings
		if timeout <= 0 {
			return fmt.Errorf("timeout must be greater than 0")
		}

		if scheme != "http" && scheme != "https" {
			return fmt.Errorf("scheme must be either 'http' or 'https'")
		}

		if maxLogSize <= 0 {
			return fmt.Errorf("max-log-size must be greater than 0")
		}
		if maxBackups < 0 {
			return fmt.Errorf("max-backups cannot be negative")
		}
		if maxAge < 0 {
			return fmt.Errorf("max-age cannot be negative")
		}

		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Define flags with defaults
	rootCmd.PersistentFlags().StringVar(&username, "username", "", "AdGuard Home username")
	rootCmd.PersistentFlags().StringVar(&password, "password", "", "AdGuard Home password")
	rootCmd.PersistentFlags().StringVar(&adguardURL, "adguard-url", "127.0.0.1:3000", "AdGuard Home host:port")
	rootCmd.PersistentFlags().StringVar(&leasePath, "lease-path", "/var/dhcpd/var/db/dhcpd.leases", "Path to DHCP leases file")
	rootCmd.PersistentFlags().StringVar(&leaseFormat, "lease-format", "isc", "DHCP lease file format (isc or dnsmasq)")
	rootCmd.PersistentFlags().StringVar(&scheme, "scheme", "http", "Connection scheme (http/https)")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 10, "API timeout in seconds")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Dry run mode (print actions instead of executing)")
	rootCmd.PersistentFlags().BoolVar(&preserveDeletedHosts, "preserve-deleted-hosts", false, "Don't remove AdGuard clients when their DHCP leases expire")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Show debug info")

	// Add logging flags
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (error, warn, info, debug)")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Log file path (default: syslog for service, stdout for CLI)")
	rootCmd.PersistentFlags().IntVar(&maxLogSize, "max-log-size", 100, "Maximum log file size in megabytes before rotation")
	rootCmd.PersistentFlags().IntVar(&maxBackups, "max-backups", 3, "Maximum number of old log files to retain")
	rootCmd.PersistentFlags().IntVar(&maxAge, "max-age", 28, "Maximum number of days to retain old log files")
	rootCmd.PersistentFlags().BoolVar(&noCompress, "no-compress", false, "Disable compression of old log files")

	// Note: Required flags are now handled per-command in validateAdGuardFlags
}
