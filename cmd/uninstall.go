// cmd/uninstall.go
package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

var (
	removeConfig bool // Flag to remove config directory
	forceful     bool // Flag to ignore errors
)

// isProcessRunning checks if a process with given PID exists and is our service
func isProcessRunning(pid int) bool {
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix systems, FindProcess always succeeds, so we need to check if the process actually exists
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// isServiceRunning checks if the service is currently running
func isServiceRunning() (bool, int) {
	pidfile := "/var/run/dhcp_adguard_sync.pid"

	// Check if pidfile exists
	content, err := os.ReadFile(pidfile)
	if err != nil {
		return false, 0
	}

	// Read PID from file
	var pid int
	_, err = fmt.Sscanf(string(content), "%d", &pid)
	if err != nil {
		return false, 0
	}

	// Check if process is running
	return isProcessRunning(pid), pid
}

// killProcess attempts to kill a process by PID
func killProcess(pid int) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return process.Kill()
}

// stopService attempts to stop the service gracefully, then forcefully if needed
func stopService() error {
	// First check if service is running
	running, pid := isServiceRunning()
	if !running {
		fmt.Println("Service is not running")
		return nil
	}

	fmt.Printf("Service is running with PID %d\n", pid)
	fmt.Println("Attempting graceful service stop...")

	stopCmd := exec.Command("service", "dhcp-adguard-sync", "stop")
	stopCmd.Start()

	// Give it some time to stop gracefully
	done := make(chan error, 1)
	go func() {
		done <- stopCmd.Wait()
	}()

	// Wait up to 10 seconds for graceful shutdown
	select {
	case err := <-done:
		if err == nil {
			fmt.Println("Service stopped gracefully")
			return nil
		}
		fmt.Printf("Graceful stop failed: %v\n", err)
	case <-time.After(10 * time.Second):
		fmt.Println("Service stop timed out")
	}

	// If we're here, graceful shutdown failed or timed out
	if !forceful {
		return &cleanError{message: "service failed to stop gracefully. Use --force to force termination"}
	}

	// Force kill the process
	fmt.Println("Attempting forceful termination...")
	if err := killProcess(pid); err != nil {
		fmt.Printf("Failed to kill process: %v\n", err)
	} else {
		fmt.Printf("Process %d terminated\n", pid)
	}

	// Clean up pidfile
	pidfile := "/var/run/dhcp_adguard_sync.pid"
	if err := os.Remove(pidfile); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: Failed to remove pidfile: %v\n", err)
	}

	return nil
}

// cleanError is a custom error type that prevents help from being shown
type cleanError struct {
	message string
}

func (e *cleanError) Error() string {
	return e.message
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall the service and optionally remove configuration",
	Long: `Uninstall dhcp-adguard-sync service and files.
This will:
1. Stop and disable the service
2. Remove the rc.d service script
3. Remove the binary from /usr/local/bin
4. Optionally remove configuration files (with --remove-config flag)`,
	SilenceUsage: true, // Don't show usage on error
	RunE: func(cmd *cobra.Command, args []string) error {
		// First stop and disable the service
		fmt.Println("Stopping and disabling service...")
		if err := stopService(); err != nil && !forceful {
			return err
		}

		if err := exec.Command("service", "dhcp-adguard-sync", "disable").Run(); err != nil && !forceful {
			return &cleanError{message: fmt.Sprintf("failed to disable service: %v", err)}
		}

		// Remove rc.d script
		fmt.Println("Removing rc.d script...")
		if err := os.Remove(RCPath); err != nil && !os.IsNotExist(err) && !forceful {
			return &cleanError{message: fmt.Sprintf("failed to remove rc.d script: %v", err)}
		}

		// Remove binary
		fmt.Println("Removing binary...")
		if err := os.Remove(InstallPath); err != nil && !os.IsNotExist(err) && !forceful {
			return &cleanError{message: fmt.Sprintf("failed to remove binary: %v", err)}
		}

		// Optionally remove config directory
		if removeConfig {
			fmt.Println("Removing configuration directory...")
			configDir := filepath.Dir(ConfigPath)
			if err := os.RemoveAll(configDir); err != nil && !os.IsNotExist(err) && !forceful {
				return &cleanError{message: fmt.Sprintf("failed to remove config directory: %v", err)}
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
	uninstallCmd.Flags().BoolVar(&forceful, "force", false, "Continue even if errors occur and force process termination")
}
