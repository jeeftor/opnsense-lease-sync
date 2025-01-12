// cmd/uninstall.go changes
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

var (
	removeConfig bool // Flag to remove config directory
	forceful     bool // Flag to ignore errors
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the service and optionally remove configuration",
	Long: `Uninstall dhcp-adguard-sync service and files.
This will:
1. Stop and disable the service
2. Remove the rc.d service script
3. Remove the binary from /usr/local/bin
4. Optionally remove configuration files (with --remove-config flag)`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// First stop and disable the service
		fmt.Println("Stopping and disabling service...")
		if err := exec.Command("service", "dhcp-adguard-sync", "stop").Run(); err != nil && !forceful {
			return fmt.Errorf("failed to stop service: %w", err)
		}

		if err := exec.Command("service", "dhcp-adguard-sync", "disable").Run(); err != nil && !forceful {
			return fmt.Errorf("failed to disable service: %w", err)
		}

		// Remove rc.d script
		fmt.Println("Removing rc.d script...")
		if err := os.Remove(RCPath); err != nil && !os.IsNotExist(err) && !forceful {
			return fmt.Errorf("failed to remove rc.d script: %w", err)
		}

		// Remove binary
		fmt.Println("Removing binary...")
		if err := os.Remove(InstallPath); err != nil && !os.IsNotExist(err) && !forceful {
			return fmt.Errorf("failed to remove binary: %w", err)
		}

		// Optionally remove config directory
		if removeConfig {
			fmt.Println("Removing configuration directory...")
			configDir := filepath.Dir(ConfigPath)
			if err := os.RemoveAll(configDir); err != nil && !os.IsNotExist(err) && !forceful {
				return fmt.Errorf("failed to remove config directory: %w", err)
			}
		} else {
			fmt.Println("Configuration directory preserved at:", filepath.Dir(ConfigPath))
		}

		fmt.Println("Uninstallation completed successfully!")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(uninstallCmd)

	// Add flags
	uninstallCmd.Flags().BoolVar(&removeConfig, "remove-config", false, "Remove configuration directory")
	uninstallCmd.Flags().BoolVar(&forceful, "force", false, "Continue even if errors occur")
}
