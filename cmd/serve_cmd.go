// cmd/serve_cmd.go
package cmd

import (
	"context"
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

		// Create a context that we'll cancel on shutdown
		_, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Set up signal handling
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Create error channel for service errors
		errChan := make(chan error, 1)

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

		// Start service in a goroutine
		go func() {
			logger.Info("Starting service...")
			if err := syncService.Run(); err != nil {
				logger.Error(fmt.Sprintf("Service error: %v", err))
				errChan <- err
			}
		}()

		// Wait for either:
		// - A signal (SIGINT/SIGTERM)
		// - An error from the service
		select {
		case sig := <-sigChan:
			logger.Info(fmt.Sprintf("Received signal: %v", sig))
			logger.Info("Initiating graceful shutdown...")

			// Cancel context to notify service to stop
			cancel()

			// Call Stop() to cleanup
			if err := syncService.Stop(); err != nil {
				logger.Error(fmt.Sprintf("Error during shutdown: %v", err))
				return fmt.Errorf("error stopping service: %w", err)
			}

			logger.Info("Service stopped successfully")
			return nil

		case err := <-errChan:
			logger.Error(fmt.Sprintf("Service failed: %v", err))
			return fmt.Errorf("service failed: %w", err)
		}
	},
}

func init() {
	serveCmd.MarkFlagRequired("username")
	serveCmd.MarkFlagRequired("password")

	rootCmd.AddCommand(serveCmd)
}
