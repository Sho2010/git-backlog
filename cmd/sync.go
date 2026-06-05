package cmd

import (
	"fmt"
	"os"
	"sort"
	"sync"

	"github.com/spf13/cobra"

	"github.com/Sho2010/git-backlog/internal/backlog"
	"github.com/Sho2010/git-backlog/internal/cache"
	"github.com/Sho2010/git-backlog/internal/config"
	"github.com/Sho2010/git-backlog/internal/git"
	"github.com/Sho2010/git-backlog/internal/issuekey"
)

var (
	syncForce       bool
	syncConcurrency int
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Fetch Backlog issues for all local branches and cache them",
	RunE:  runSync,
}

func init() {
	syncCmd.Flags().BoolVar(&syncForce, "force", false, "Refetch even if cache is fresh")
	syncCmd.Flags().IntVar(&syncConcurrency, "concurrency", 4, "Maximum parallel API requests")
	rootCmd.AddCommand(syncCmd)
}

type syncResult struct {
	Key    string
	Status string // "fetched", "cached", "error"
	Err    error
}

func runSync(cmd *cobra.Command, args []string) error {
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

	seen := make(map[string]struct{})
	var keys []string
	for _, b := range branches {
		key, ok, err := issuekey.Extract(b, cfg.IssuePattern)
		if err != nil {
			return err
		}
		if !ok {
			continue
		}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	if len(keys) == 0 {
		return nil
	}
	sort.Strings(keys)

	c, err := cache.New(cfg.CacheTTL)
	if err != nil {
		return err
	}

	apiKey, err := cfg.APIKey()
	if err != nil {
		return err
	}
	client := backlog.NewClient(cfg.BaseURL, apiKey)

	if syncConcurrency < 1 {
		syncConcurrency = 1
	}
	sem := make(chan struct{}, syncConcurrency)
	results := make(chan syncResult, len(keys))
	var wg sync.WaitGroup

	for _, k := range keys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			results <- fetchOne(c, client, key, syncForce)
		}(k)
	}
	wg.Wait()
	close(results)

	collected := make([]syncResult, 0, len(keys))
	for r := range results {
		collected = append(collected, r)
	}
	sort.Slice(collected, func(i, j int) bool { return collected[i].Key < collected[j].Key })

	var fetched, cached, failed int
	for _, r := range collected {
		switch r.Status {
		case "fetched":
			fetched++
			fmt.Printf("fetched  %s\n", r.Key)
		case "cached":
			cached++
			fmt.Printf("cached   %s\n", r.Key)
		case "error":
			failed++
			fmt.Fprintf(os.Stderr, "error    %s: %v\n", r.Key, r.Err)
		}
	}
	fmt.Fprintf(os.Stderr, "\n%d fetched, %d cached, %d failed\n", fetched, cached, failed)
	if failed > 0 {
		return fmt.Errorf("%d issue(s) failed to sync", failed)
	}
	return nil
}

func fetchOne(c *cache.Cache, client *backlog.Client, key string, force bool) syncResult {
	if !force {
		if _, hit, err := c.Get(key); err == nil && hit {
			return syncResult{Key: key, Status: "cached"}
		}
	}
	issue, err := client.GetIssue(key)
	if err != nil {
		return syncResult{Key: key, Status: "error", Err: err}
	}
	if err := c.Put(issue); err != nil {
		return syncResult{Key: key, Status: "error", Err: err}
	}
	return syncResult{Key: key, Status: "fetched"}
}
