// cmd/plugin.go
package cmd

import (
	"github.com/spf13/cobra"
)

// pluginCmd represents the plugin command
var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage the OPNsense web UI plugin",
	Long:  `Install or uninstall the OPNsense web UI plugin.`,
}

func init() {
	rootCmd.AddCommand(pluginCmd)
}
