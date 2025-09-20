// cmd/sync_cmd.go
package cmd

import (
	"dhcpsync/pkg"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Run a one-time sync of DHCP leases to AdGuard",
	Long: `Performs a single synchronization of DHCP leases to AdGuard Home
and then exits. This is useful for testing or manual synchronization.`,
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
			BSDFormat:  bsdFormat,
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

		// Run one sync and exit
		if err := syncService.Sync(); err != nil {
			return fmt.Errorf("sync failed: %w", err)
		}

		return nil
	},
}

func init() {
	// Add flags specific to the sync command
	syncCmd.Flags().String("username", "", "AdGuard Home username")
	syncCmd.Flags().String("password", "", "AdGuard Home password")
	syncCmd.Flags().String("adguard-url", "127.0.0.1:3000", "AdGuard Home host:port")
	syncCmd.Flags().String("lease-path", "/var/db/dnsmasq.leases", "Path to DHCP leases file")
	syncCmd.Flags().String("lease-format", "dnsmasq", "DHCP lease file format (isc or dnsmasq)")
	syncCmd.Flags().String("scheme", "http", "Connection scheme (http/https)")
	syncCmd.Flags().Int("timeout", 10, "API timeout in seconds")
	syncCmd.Flags().Bool("dry-run", false, "Dry run mode")
	syncCmd.Flags().Bool("preserve-deleted-hosts", false, "Preserve deleted hosts")

	// Mark required flags
	syncCmd.MarkFlagRequired("username")
	syncCmd.MarkFlagRequired("password")
}
