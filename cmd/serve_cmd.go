// cmd/serve_cmd.go
package cmd

import (
	"context"
	"dhcpsync/pkg"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Run as a service, watching for lease file changes",
	Long: `Runs as a persistent service that watches the lease file for changes
and automatically syncs them to AdGuard Home. This is the recommended
mode for production use.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get flags specific to this command
		username, _ := cmd.Flags().GetString("username")
		password, _ := cmd.Flags().GetString("password")
		adguardURL, _ := cmd.Flags().GetString("adguard-url")
		leasePath, _ := cmd.Flags().GetString("lease-path")
		leaseFormat, _ := cmd.Flags().GetString("lease-format")
		scheme, _ := cmd.Flags().GetString("scheme")
		timeout, _ := cmd.Flags().GetInt("timeout")
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		preserveDeletedHosts, _ := cmd.Flags().GetBool("preserve-deleted-hosts")

		// Determine log file path - check environment variable if flag not set
		logFilePath := logFile
		if logFilePath == "" {
			if envLogFile := os.Getenv("LOG_FILE"); envLogFile != "" {
				logFilePath = envLogFile
			}
		}

		// Create log configuration from global flags
		logConfig := pkg.LogConfig{
			Level:      pkg.ParseLogLevel(logLevel),
			FilePath:   logFilePath,
			SyslogOnly: syslogOnly,
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

		// Print debug info if log level is debug
		if logLevel == "debug" {
			logger.Info("=== Configuration Debug Info ===")
			logger.Info(fmt.Sprintf("AdGuard URL: %s", adguardURL))
			logger.Info(fmt.Sprintf("AdGuard Scheme: %s", scheme))
			logger.Info(fmt.Sprintf("AdGuard Username: %s", username))
			logger.Info(fmt.Sprintf("AdGuard Password: %s", "[REDACTED]"))
			logger.Info(fmt.Sprintf("AdGuard Timeout: %d", timeout))
			logger.Info(fmt.Sprintf("DHCP Lease Path: %s", leasePath))
			logger.Info(fmt.Sprintf("DHCP Lease Format: %s", leaseFormat))
			logger.Info(fmt.Sprintf("Dry Run: %t", dryRun))
			logger.Info(fmt.Sprintf("Preserve Deleted Hosts: %t", preserveDeletedHosts))
			logger.Info(fmt.Sprintf("Log Level: %s", logLevel))
			logger.Info(fmt.Sprintf("Log File: %s", logFilePath))
			logger.Info("===============================")
		}

		// Create a context that we'll cancel on shutdown
		_, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Set up signal handling
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		// Create error channel for service errors
		errChan := make(chan error, 1)

		// Convert string lease format to LeaseFormat type
		var leaseFormatType pkg.LeaseFormat
		if leaseFormat == "dnsmasq" {
			leaseFormatType = pkg.DNSMasqFormat
		} else {
			leaseFormatType = pkg.ISCDHCPFormat
		}

		syncService, err := pkg.NewSyncService(pkg.Config{
			AdGuardURL:           adguardURL,
			LeasePath:            leasePath,
			LeaseFormat:          leaseFormatType,
			DryRun:               dryRun,
			Username:             username,
			Password:             password,
			Scheme:               scheme,
			Timeout:              timeout,
			Logger:               logger,
			PreserveDeletedHosts: preserveDeletedHosts,
			Debug:                logLevel == "debug",
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
	// Add flags specific to the serve command
	serveCmd.Flags().String("username", "", "AdGuard Home username")
	serveCmd.Flags().String("password", "", "AdGuard Home password")
	serveCmd.Flags().String("adguard-url", "127.0.0.1:3000", "AdGuard Home host:port")
	serveCmd.Flags().String("lease-path", "/var/db/dnsmasq.leases", "Path to DHCP leases file")
	serveCmd.Flags().String("lease-format", "dnsmasq", "DHCP lease file format (isc or dnsmasq)")
	serveCmd.Flags().String("scheme", "http", "Connection scheme (http/https)")
	serveCmd.Flags().Int("timeout", 10, "API timeout in seconds")
	serveCmd.Flags().Bool("dry-run", false, "Dry run mode")
	serveCmd.Flags().Bool("preserve-deleted-hosts", false, "Preserve deleted hosts")

	// Mark required flags
	serveCmd.MarkFlagRequired("username")
	serveCmd.MarkFlagRequired("password")
}
