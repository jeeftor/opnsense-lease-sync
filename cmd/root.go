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
	dryRun               bool
	scheme               string
	timeout              int
	preserveDeletedHosts bool
	debug                bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "opnsense-lease-sync",
	Short: "Sync ISC DHCP leases to AdGuard Home",
	Long: `A service/CLI tool that synchronizes ISC DHCP leases from OPNsense
to AdGuard Home, keeping client configurations in sync automatically.

Can be run either as a one-time sync (CLI mode) or as a persistent service
that watches for lease file changes.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Check for environment variables
		if envUser := os.Getenv("ADGUARD_USERNAME"); envUser != "" && !cmd.Flags().Changed("username") {
			username = envUser
		}
		if envPass := os.Getenv("ADGUARD_PASSWORD"); envPass != "" && !cmd.Flags().Changed("password") {
			password = envPass
		}
		if envURL := os.Getenv("ADGUARD_URL"); envURL != "" && !cmd.Flags().Changed("adguard-url") {
			adguardURL = envURL
		}
		if envLease := os.Getenv("DHCP_LEASE_PATH"); envLease != "" && !cmd.Flags().Changed("lease-path") {
			leasePath = envLease
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

		// Validate required flags
		if username == "" {
			return fmt.Errorf("username is required. Set via --username flag or ADGUARD_USERNAME environment variable")
		}
		if password == "" {
			return fmt.Errorf("password is required. Set via --password flag or ADGUARD_PASSWORD environment variable")
		}

		// Validate timeout
		if timeout <= 0 {
			return fmt.Errorf("timeout must be greater than 0")
		}

		// Validate scheme
		if scheme != "http" && scheme != "https" {
			return fmt.Errorf("scheme must be either 'http' or 'https'")
		}

		if envDebug := os.Getenv("DEBUG"); envDebug != "" && !cmd.Flags().Changed("debug") {
			debug = envDebug == "true" || envDebug == "1"
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
	rootCmd.PersistentFlags().StringVar(&scheme, "scheme", "http", "Connection scheme (http/https)")
	rootCmd.PersistentFlags().IntVar(&timeout, "timeout", 10, "API timeout in seconds")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Dry run mode (print actions instead of executing)")
	rootCmd.PersistentFlags().BoolVar(&preserveDeletedHosts, "preserve-deleted-hosts", false, "Don't remove AdGuard clients when their DHCP leases expire")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Show debug info")

	// Add flag usage examples to help text
	rootCmd.Example = `  # Run with username and password
  opnsense-lease-sync sync --username admin --password mypassword

  # Run with environment variables
  export ADGUARD_USERNAME=admin
  export ADGUARD_PASSWORD=mypassword
  opnsense-lease-sync sync

  # Run as a service with custom settings
  opnsense-lease-sync serve \
    --username admin \
    --password mypassword \
    --adguard-url 192.168.1.1:3000 \
    --scheme https \
    --preserve-deleted-hosts`

	// Mark required flags
	rootCmd.MarkPersistentFlagRequired("username")
	rootCmd.MarkPersistentFlagRequired("password")
}
