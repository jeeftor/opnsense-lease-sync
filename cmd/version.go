// cmd/version.go
package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version information set by build flags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  `Display the version, build date, and git commit of dhcpsync`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("dhcpsync version %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built: %s\n", date)
		fmt.Printf("  runtime: %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	// No flags for version command
}
