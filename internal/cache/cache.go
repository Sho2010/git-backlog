package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Sho2010/git-backlog/internal/backlog"
)

type Cache struct {
	Dir string
	TTL time.Duration
}

func New(ttl time.Duration) (*Cache, error) {
	base := os.Getenv("XDG_CACHE_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("user home: %w", err)
		}
		base = filepath.Join(home, ".cache")
	}
	dir := filepath.Join(base, "git-backlog")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir %s: %w", dir, err)
	}
	return &Cache{Dir: dir, TTL: ttl}, nil
}

func (c *Cache) path(issueKey string) string {
	return filepath.Join(c.Dir, issueKey+".json")
}

func (c *Cache) Get(issueKey string) (*backlog.Issue, bool, error) {
	p := c.path(issueKey)
	info, err := os.Stat(p)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("stat cache: %w", err)
	}
	if time.Since(info.ModTime()) > c.TTL {
		return nil, false, nil
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, false, fmt.Errorf("read cache: %w", err)
	}
	var issue backlog.Issue
	if err := json.Unmarshal(data, &issue); err != nil {
		return nil, false, nil
	}
	return &issue, true, nil
}

func (c *Cache) Put(issue *backlog.Issue) error {
	data, err := json.Marshal(issue)
	if err != nil {
		return fmt.Errorf("marshal issue: %w", err)
	}
	p := c.path(issue.IssueKey)
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write cache: %w", err)
	}
	if err := os.Rename(tmp, p); err != nil {
		return fmt.Errorf("rename cache: %w", err)
	}
	return nil
}
