package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var installPluginCmd = &cobra.Command{
	Use:   "install-plugin",
	Short: "Install the OPNsense web UI plugin (stub for future implementation)",
	Long: `Install the OPNsense web UI plugin.

This command is currently a stub for future implementation.

For now, install the plugin manually using:
1. Sync plugin files with dev-sync.sh (for development)
2. Or follow the manual installation instructions in the README`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Plugin installation is not yet implemented.")
		fmt.Println("")
		fmt.Println("For development, use the dev-sync.sh script to sync plugin files.")
		fmt.Println("For production, follow the manual plugin installation instructions in the README.")
		fmt.Println("")
		fmt.Println("Plugin installation via this command will be implemented in a future version.")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(installPluginCmd)
}
