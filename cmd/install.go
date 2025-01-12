// cmd/install.go
package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the service and configuration files",
	Long: `Install the dhcp-adguard-sync binary, configuration, and service files.
This will:
1. Copy the binary to /usr/local/bin
2. Create a config file in /usr/local/etc/dhcp-adguard-sync
3. Install the rc.d service script
4. Set appropriate permissions`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if running on FreeBSD
		if runtime.GOOS != "freebsd" {
			return fmt.Errorf("installation is only supported on FreeBSD systems")
		}

		// Get the path of the current binary
		executable, err := os.Executable()
		if err != nil {
			return fmt.Errorf("failed to get executable path: %w", err)
		}

		// Create installation directory if it doesn't exist
		if err := os.MkdirAll(filepath.Dir(InstallPath), 0755); err != nil {
			return fmt.Errorf("failed to create installation directory: %w", err)
		}

		// Copy binary
		if err := copyFile(executable, InstallPath); err != nil {
			return fmt.Errorf("failed to copy binary: %w", err)
		}
		if err := os.Chmod(InstallPath, 0755); err != nil {
			return fmt.Errorf("failed to set binary permissions: %w", err)
		}

		// Create config directory
		configDir := filepath.Dir(ConfigPath)
		if err := os.MkdirAll(configDir, 0755); err != nil {
			return fmt.Errorf("failed to create config directory: %w", err)
		}

		// Extract config file if it doesn't exist
		if _, err := os.Stat(ConfigPath); os.IsNotExist(err) {
			if err := extractFile("config.yaml", ConfigPath, 0600); err != nil {
				return fmt.Errorf("failed to create config file: %w", err)
			}
		}

		// Extract rc.d script
		if err := extractFile("rc.script", RCPath, 0755); err != nil {
			return fmt.Errorf("failed to create rc.d script: %w", err)
		}

		// Enable the service
		if err := exec.Command("service", "dhcp-adguard-sync", "enable").Run(); err != nil {
			return fmt.Errorf("failed to enable service: %w", err)
		}

		fmt.Println("Installation completed successfully!")
		fmt.Printf("Please edit %s with your settings\n", ConfigPath)
		fmt.Println("Then start the service with: service dhcp-adguard-sync start")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(installCmd)
}

func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}
