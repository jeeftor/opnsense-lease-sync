package cmd

import (
	"fmt"
	"opnsense-lease-sync/internal/logger"
	"opnsense-lease-sync/internal/service"
	_ "os"

	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Perform a one-time sync of DHCP leases to AdGuard",
	Long: `Syncs the current DHCP leases to AdGuard Home once and exits.
This is useful for testing or for running as a scheduled task.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, err := logger.NewLogger()
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		syncService, err := service.New(service.Config{
			AdGuardURL: adguardURL,
			LeasePath:  leasePath,
			DryRun:     dryRun,
			Logger:     logger,
		})
		if err != nil {
			return fmt.Errorf("failed to create service: %w", err)
		}

		if err := syncService.Run(); err != nil {
			return fmt.Errorf("sync failed: %w", err)
		}

		// For sync command, we want to run once and exit
		if err := syncService.Stop(); err != nil {
			return fmt.Errorf("error stopping service: %w", err)
		}

		return nil
	},
}
