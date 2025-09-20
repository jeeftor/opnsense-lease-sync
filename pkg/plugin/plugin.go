package plugin

import (
	"embed"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed opnsense-plugin/*
var pluginFiles embed.FS

// InstallPlugin installs the embedded OPNsense plugin files to the system
func InstallPlugin(prefix string, force bool) error {
	if prefix == "" {
		prefix = "/usr/local"
	}
	// Define the mapping of source to destination paths
	fileMappings := map[string]string{
		"opnsense-plugin/mvc/app/views/OPNsense/Dhcpsync":       filepath.Join(prefix, "opnsense/mvc/app/views/OPNsense/Dhcpsync"),
		"opnsense-plugin/mvc/app/models/OPNsense/Dhcpsync":      filepath.Join(prefix, "opnsense/mvc/app/models/OPNsense/Dhcpsync"),
		"opnsense-plugin/mvc/app/controllers/OPNsense/Dhcpsync": filepath.Join(prefix, "opnsense/mvc/app/controllers/OPNsense/Dhcpsync"),
		"opnsense-plugin/service/templates/OPNsense/Dhcpsync":   filepath.Join(prefix, "opnsense/service/templates/OPNsense/Dhcpsync"),
		"opnsense-plugin/service/conf":                          filepath.Join(prefix, "opnsense/service/conf"),
	}

	fmt.Println("Installing OPNsense plugin files...")

	for embeddedPath, destPath := range fileMappings {
		// Check if the embedded path exists
		if _, err := fs.Stat(pluginFiles, embeddedPath); err != nil {
			fmt.Printf("Skipping %s (not found in embedded files)\n", embeddedPath)
			continue
		}

		fmt.Printf("Installing %s -> %s\n", embeddedPath, destPath)

		if err := copyEmbeddedDir(embeddedPath, destPath, force); err != nil {
			return fmt.Errorf("failed to copy %s to %s: %v", embeddedPath, destPath, err)
		}
	}

	fmt.Println("\nPlugin installation completed successfully!")
	fmt.Println("\nRecommended next steps:")
	fmt.Println("  service configd restart")
	fmt.Println("  configctl webgui restart")

	return nil
}

// UninstallPlugin removes the OPNsense plugin files from the system
func UninstallPlugin(prefix string) error {
	if prefix == "" {
		prefix = "/usr/local"
	}
	// Define the destination paths to remove
	pluginPaths := []string{
		filepath.Join(prefix, "opnsense/mvc/app/views/OPNsense/Dhcpsync"),
		filepath.Join(prefix, "opnsense/mvc/app/models/OPNsense/Dhcpsync"),
		filepath.Join(prefix, "opnsense/mvc/app/controllers/OPNsense/Dhcpsync"),
		filepath.Join(prefix, "opnsense/service/templates/OPNsense/Dhcpsync"),
		filepath.Join(prefix, "opnsense/service/conf/actions.d/actions_dhcpsync.conf"),
	}

	fmt.Println("Uninstalling OPNsense plugin files...")

	for _, path := range pluginPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			fmt.Printf("Skipping %s (not found)\n", path)
			continue
		}

		fmt.Printf("Removing %s\n", path)
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %v", path, err)
		}
	}

	fmt.Println("\nPlugin uninstallation completed successfully!")
	fmt.Println("\nRecommended next steps:")
	fmt.Println("  service configd restart")
	fmt.Println("  configctl webgui restart")

	return nil
}

// copyEmbeddedDir copies an entire directory tree from the embedded filesystem to the destination
func copyEmbeddedDir(embeddedPath, destPath string, force bool) error {
	return fs.WalkDir(pluginFiles, embeddedPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate the relative path from the embedded base
		relPath, err := filepath.Rel(embeddedPath, path)
		if err != nil {
			return err
		}

		// Calculate destination path
		dstPath := filepath.Join(destPath, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		return copyEmbeddedFile(path, dstPath, force)
	})
}

// copyEmbeddedFile copies a single file from the embedded filesystem to the destination
func copyEmbeddedFile(embeddedPath, destPath string, force bool) error {
	// Check if file exists and handle accordingly
	if !force {
		if _, err := os.Stat(destPath); err == nil {
			return fmt.Errorf("file %s already exists (use --force to overwrite)", destPath)
		}
	} else {
		// If forcing and file exists, make it writable first
		if _, err := os.Stat(destPath); err == nil {
			os.Chmod(destPath, 0644)
		}
	}
	// Open the embedded file
	srcFile, err := pluginFiles.Open(embeddedPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	// Create destination file
	dstFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// Copy file contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	// Get file info for permissions
	srcInfo, err := fs.Stat(pluginFiles, embeddedPath)
	if err != nil {
		return err
	}

	// Set file permissions
	return os.Chmod(destPath, srcInfo.Mode())
}
