// cmd/sync.go
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"opnsense-lease-sync/pkg"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Run a one-time sync of DHCP leases to AdGuard",
	Long: `Performs a single synchronization of DHCP leases to AdGuard Home
and then exits. This is useful for testing or manual synchronization.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, err := pkg.NewLogger()
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		syncService, err := pkg.NewSyncService(pkg.Config{
			AdGuardURL: adguardURL,
			LeasePath:  leasePath,
			DryRun:     dryRun,
			Username:   username,
			Password:   password,
			Scheme:     scheme,
			Timeout:    timeout,
			Logger:     logger,
			Debug:      debug,
		})
		if err != nil {
			return fmt.Errorf("failed to create service: %w", err)
		}

		// Run one sync and exit
		if err := syncService.Sync(); err != nil {
			return fmt.Errorf("sync failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(syncCmd)
}
