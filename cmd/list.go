package cmd

import (
	"fmt"
	"os"
	"sync"
	"text/tabwriter"
	"text/template"

	"github.com/spf13/cobra"

	"github.com/Sho2010/git-backlog/internal/backlog"
	"github.com/Sho2010/git-backlog/internal/cache"
	"github.com/Sho2010/git-backlog/internal/config"
	"github.com/Sho2010/git-backlog/internal/git"
	"github.com/Sho2010/git-backlog/internal/issuekey"
)

type listRow struct {
	Branch   string
	IssueKey string
	Summary  string
	Resolved bool
}

var (
	listNoFetch     bool
	listConcurrency int
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List local branches with their Backlog issue titles",
	RunE:  runList,
}

func init() {
	listCmd.Flags().BoolVar(&listNoFetch, "no-fetch", false, "Don't fetch missing entries from the API (cache only)")
	listCmd.Flags().IntVar(&listConcurrency, "concurrency", 4, "Maximum parallel API requests when fetching missing entries")
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
	var missing []string
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
			row.Resolved = true
		} else {
			missing = append(missing, key)
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		return nil
	}

	if !listNoFetch && len(missing) > 0 {
		apiKey, err := cfg.APIKey()
		if err != nil {
			return err
		}
		client := backlog.NewClient(cfg.BaseURL, apiKey)
		fetched := fetchKeys(c, client, missing, listConcurrency)
		for i := range rows {
			if rows[i].Resolved {
				continue
			}
			if issue, ok := fetched[rows[i].IssueKey]; ok {
				rows[i].Summary = issue.Summary
				rows[i].Resolved = true
			}
		}
	}

	if formatFlag != "" {
		return renderListWithTemplate(rows)
	}
	return renderListPlain(rows)
}

func fetchKeys(c *cache.Cache, client *backlog.Client, keys []string, concurrency int) map[string]*backlog.Issue {
	if concurrency < 1 {
		concurrency = 1
	}
	type res struct {
		key   string
		issue *backlog.Issue
	}
	results := make(chan res, len(keys))
	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	for _, k := range keys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			issue, err := client.GetIssue(key)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warning: fetch %s failed: %v\n", key, err)
				results <- res{key: key}
				return
			}
			if err := c.Put(issue); err != nil {
				fmt.Fprintf(os.Stderr, "warning: cache write %s failed: %v\n", key, err)
			}
			results <- res{key: key, issue: issue}
		}(k)
	}
	wg.Wait()
	close(results)

	out := make(map[string]*backlog.Issue, len(keys))
	for r := range results {
		if r.issue != nil {
			out[r.key] = r.issue
		}
	}
	return out
}

func renderListPlain(rows []listRow) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, r := range rows {
		summary := r.Summary
		if !r.Resolved {
			if listNoFetch {
				summary = "(uncached — drop --no-fetch or run `git backlog sync`)"
			} else {
				summary = "(unresolved)"
			}
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
