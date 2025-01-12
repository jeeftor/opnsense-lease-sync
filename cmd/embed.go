// cmd/embedded.go
package cmd

import (
	"embed"
	"os"
	"path/filepath"
)

//go:embed templates/*
var templates embed.FS

// extractFile extracts an embedded file to the specified path
func extractFile(templatePath, destPath string, mode os.FileMode) error {
	content, err := templates.ReadFile(filepath.Join("templates", templatePath))
	if err != nil {
		return err
	}
	return os.WriteFile(destPath, content, mode)
}
