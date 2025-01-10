package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	b64auth    string
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
	rootCmd.PersistentFlags().StringVar(&b64auth, "b64auth", "", "Base64 encoded credentials for AdGuard Home API (username:password)")
	rootCmd.PersistentFlags().StringVar(&adguardURL, "adguard-url", "", "AdGuard Home URL")
	rootCmd.PersistentFlags().StringVar(&leasePath, "lease-path", "", "Path to DHCP leases file")
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Dry run mode")
}
