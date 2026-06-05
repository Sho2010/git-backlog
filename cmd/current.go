package cmd

import (
	"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/Sho2010/git-backlog/internal/backlog"
	"github.com/Sho2010/git-backlog/internal/cache"
	"github.com/Sho2010/git-backlog/internal/config"
	"github.com/Sho2010/git-backlog/internal/git"
	"github.com/Sho2010/git-backlog/internal/issuekey"
)

var formatFlag string

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Print the Backlog issue title for the current branch",
	RunE:  runCurrent,
}

func runCurrent(cmd *cobra.Command, args []string) error {
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

	c, err := cache.New(cfg.CacheTTL)
	if err != nil {
		return err
	}

	issue, hit, err := c.Get(key)
	if err != nil {
		return err
	}
	if !hit {
		apiKey, err := cfg.APIKey()
		if err != nil {
			return err
		}
		client := backlog.NewClient(cfg.BaseURL, apiKey)
		issue, err = client.GetIssue(key)
		if err != nil {
			return err
		}
		if err := c.Put(issue); err != nil {
			fmt.Fprintln(os.Stderr, "warning: cache write failed:", err)
		}
	}

	return render(issue)
}

func render(issue *backlog.Issue) error {
	if formatFlag == "" {
		fmt.Printf("[%s] %s\n", issue.IssueKey, issue.Summary)
		return nil
	}
	tmpl, err := template.New("out").Parse(formatFlag)
	if err != nil {
		return fmt.Errorf("parse --format: %w", err)
	}
	if err := tmpl.Execute(os.Stdout, issue); err != nil {
		return fmt.Errorf("execute --format: %w", err)
	}
	fmt.Println()
	return nil
}
