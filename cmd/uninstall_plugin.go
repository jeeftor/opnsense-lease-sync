// cmd/uninstall_plugin.go
package cmd

import (
	"dhcpsync/pkg/plugin"

	"github.com/spf13/cobra"
)

var uninstallPluginCmd = &cobra.Command{
	Use:          "uninstall-plugin",
	Short:        "Uninstall the OPNsense web UI plugin",
	SilenceUsage: true,
	Long: `Uninstall the OPNsense web UI plugin.

This command removes the plugin files from the OPNsense installation.

Use --prefix to specify a custom installation directory (e.g., --prefix=/tmp for testing).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		prefix, _ := cmd.Flags().GetString("prefix")
		return plugin.UninstallPlugin(prefix)
	},
}

func init() {
	uninstallPluginCmd.Flags().String("prefix", "", "Custom installation prefix directory (default: /usr/local)")
}
