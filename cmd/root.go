// cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Only truly global flags should be here.
	logLevel   string
	logFile    string
	syslogOnly bool
	bsdFormat  bool
	maxLogSize int
	maxBackups int
	maxAge     int
	noCompress bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "dhcpsync",
	Short: "Sync DHCP leases to AdGuard Home",
	Long: `A service/CLI tool that synchronizes DHCP leases from OPNsense
to AdGuard Home, keeping client configurations in sync automatically.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// This function will be simplified further as logic moves to subcommands.
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func init() {
	// Add all commands to the root command here. This ensures they are all discovered.
	rootCmd.AddCommand(installCmd)
	rootCmd.AddCommand(uninstallCmd)
	rootCmd.AddCommand(syncCmd)
	rootCmd.AddCommand(serveCmd)
	rootCmd.AddCommand(installPluginCmd)
	rootCmd.AddCommand(uninstallPluginCmd)
	rootCmd.AddCommand(versionCmd)

	// Only define truly global flags here.
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Log level (error, warn, info, debug)")
	rootCmd.PersistentFlags().StringVar(&logFile, "log-file", "", "Log file path (default: stdout+syslog)")
	rootCmd.PersistentFlags().BoolVar(&syslogOnly, "syslog-only", false, "Log to syslog only")
	rootCmd.PersistentFlags().BoolVar(&bsdFormat, "bsd", false, "Use BSD syslog format for file logging")
	rootCmd.PersistentFlags().IntVar(&maxLogSize, "max-log-size", 10, "Max log file size (MB)")
	rootCmd.PersistentFlags().IntVar(&maxBackups, "max-backups", 3, "Max backup log files")
	rootCmd.PersistentFlags().IntVar(&maxAge, "max-age", 28, "Max age of log files (days)")
	rootCmd.PersistentFlags().BoolVar(&noCompress, "no-compress", false, "Disable log compression")
}
