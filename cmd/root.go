package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	adguardURL string
	leasePath  string
	dryRun     bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "opnsense-lease-sync",
	Short: "Sync ISC DHCP leases to AdGuard Home",
	Long: `A service/CLI tool that synchronizes ISC DHCP leases from OPNsense
to AdGuard Home, keeping client configurations in sync automatically.

Can be run either as a one-time sync (CLI mode) or as a persistent service
that watches for lease file changes.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&adguardURL, "adguard", "http://localhost:3000", "AdGuard Home API URL")
	rootCmd.PersistentFlags().StringVar(&leasePath, "lease-file", "/var/dhcpd/var/db/dhcpd.leases", "Path to DHCP lease file")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Print changes that would be made without actually making them")
}
