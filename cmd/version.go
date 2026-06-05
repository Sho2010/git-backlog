package cmd

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

var Version = "dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println(Version)
	},
}

func init() {
	if Version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok {
			if v := info.Main.Version; v != "" && v != "(devel)" {
				Version = v
			}
		}
	}
	rootCmd.AddCommand(versionCmd)
}
