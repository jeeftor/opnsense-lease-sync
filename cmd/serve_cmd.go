// cmd/serve_cmd.go
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
		// Create log configuration
		logConfig := pkg.LogConfig{
			Level:      pkg.ParseLogLevel(logLevel),
			FilePath:   logFile,
			MaxSize:    maxLogSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   !noCompress,
		}

		// Initialize logger
		logger, err := pkg.NewLogger(logConfig)
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
			PreserveDeletedHosts: preserveDeletedHosts,
			Debug:                debug,
			LogConfig:            logConfig,
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
}
