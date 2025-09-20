package cmd

import (
	"dhcpsync/pkg/plugin"

	"github.com/spf13/cobra"
)

var (
	prefixFlag string
	forceFlag  bool
)

var installPluginCmd = &cobra.Command{
	Use:          "install-plugin",
	Short:        "Install the OPNsense web UI plugin",
	SilenceUsage: true,
	Long: `Install the OPNsense web UI plugin.

This command copies plugin files from the opnsense-plugin directory to the
appropriate locations in the OPNsense installation.

Use --prefix to specify a custom installation directory (e.g., --prefix=/tmp for testing).
Use --force to overwrite existing files without prompting.

The plugin files are embedded in the binary and will be extracted to the
appropriate OPNsense directories during installation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return plugin.InstallPlugin(prefixFlag, forceFlag)
	},
}

func init() {
	installPluginCmd.Flags().StringVar(&prefixFlag, "prefix", "", "Custom installation prefix directory (default: /usr/local)")
	installPluginCmd.Flags().BoolVar(&forceFlag, "force", false, "Overwrite existing files without prompting")
}
