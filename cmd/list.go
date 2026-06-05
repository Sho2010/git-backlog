package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/Sho2010/git-backlog/internal/cache"
	"github.com/Sho2010/git-backlog/internal/config"
	"github.com/Sho2010/git-backlog/internal/git"
	"github.com/Sho2010/git-backlog/internal/issuekey"
)

type listRow struct {
	Branch   string
	IssueKey string
	Summary  string
	Cached   bool
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List local branches with cached Backlog issue titles",
	RunE:  runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	branches, err := git.LocalBranches()
	if err != nil {
		return err
	}
	if len(branches) == 0 {
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

	rows := make([]listRow, 0, len(branches))
	for _, b := range branches {
		key, ok, err := issuekey.Extract(b, cfg.IssuePattern)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		row := listRow{Branch: b, IssueKey: key}
		issue, hit, err := c.Get(key)
		if err != nil {
			return err
		}
		if hit {
			row.Summary = issue.Summary
			row.Cached = true
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		return nil
	}

	if formatFlag != "" {
		return renderListWithTemplate(rows)
	}
	return renderListPlain(rows)
}

func renderListPlain(rows []listRow) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, r := range rows {
		summary := r.Summary
		if !r.Cached {
			summary = "(uncached — run `git backlog sync`)"
		}
		fmt.Fprintf(w, "%s\t%s\t%s\n", r.IssueKey, r.Branch, summary)
	}
	return w.Flush()
}

func renderListWithTemplate(rows []listRow) error {
	tmpl, err := template.New("row").Parse(formatFlag)
	if err != nil {
		return fmt.Errorf("parse --format: %w", err)
	}
	for _, r := range rows {
		if err := tmpl.Execute(os.Stdout, r); err != nil {
			return fmt.Errorf("execute --format: %w", err)
		}
		fmt.Println()
	}
	return nil
}
