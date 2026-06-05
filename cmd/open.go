package cmd

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/Sho2010/git-backlog/internal/config"
	"github.com/Sho2010/git-backlog/internal/git"
	"github.com/Sho2010/git-backlog/internal/issuekey"
)

var openCmd = &cobra.Command{
	Use:   "open",
	Short: "Open the current branch's Backlog issue in your browser",
	RunE:  runOpen,
}

func init() {
	rootCmd.AddCommand(openCmd)
}

func runOpen(cmd *cobra.Command, args []string) error {
	branch, err := git.CurrentBranch()
	if err != nil {
		return err
	}

	pattern, err := config.IssuePattern()
	if err != nil {
		return err
	}
	key, ok, err := issuekey.Extract(branch, pattern)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/view/%s", strings.TrimRight(cfg.BaseURL, "/"), key)
	return openURL(url)
}

func openURL(url string) error {
	var c *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		c = exec.Command("open", url)
	case "linux":
		c = exec.Command("xdg-open", url)
	case "windows":
		c = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	if err := c.Start(); err != nil {
		return fmt.Errorf("open browser: %w", err)
	}
	return nil
}
