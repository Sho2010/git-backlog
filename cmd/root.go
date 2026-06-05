package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:           "git-backlog",
	Short:         "Backlog ticket helpers for git topic-branch workflows",
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE:          runCurrent,
}

func init() {
	rootCmd.AddCommand(currentCmd)
	rootCmd.PersistentFlags().StringVar(&formatFlag, "format", "", "Go template for output (default: plain text title)")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
