// cmd/serve.go
package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
	"opnsense-lease-sync/pkg"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run as a service, watching for lease file changes",
	Long: `Runs as a persistent service that watches the lease file for changes
and automatically syncs them to AdGuard Home. This is the recommended
mode for production use.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger, err := pkg.NewLogger()
		if err != nil {
			return fmt.Errorf("failed to initialize logger: %w", err)
		}

		syncService, err := pkg.NewSyncService(pkg.Config{
			AdGuardURL:           adguardURL,
			LeasePath:            leasePath,
			DryRun:               dryRun,
			Username:             username,
			Password:             password,
			Scheme:               scheme,
			Timeout:              timeout,
			Logger:               logger,
			PreserveDeletedHosts: preserveDeletedHosts, // Add this if you want to expose it as a flag
			Debug:                debug,
		})
		if err != nil {
			return fmt.Errorf("failed to create service: %w", err)
		}

		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		if err := syncService.Run(); err != nil {
			return fmt.Errorf("service failed: %w", err)
		}

		// Wait for shutdown signal
		<-sigChan
		logger.Info("Shutting down...")

		if err := syncService.Stop(); err != nil {
			return fmt.Errorf("error stopping service: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	// If you want to add preserve-deleted-hosts flag only for serve command
	serveCmd.Flags().BoolVar(&preserveDeletedHosts, "preserve-deleted-hosts", false,
		"Don't remove AdGuard clients when their DHCP leases expire")
}
