package config

import (
	"fmt"
	"os"
	"time"

	"github.com/Sho2010/git-backlog/internal/git"
)

const (
	DefaultIssuePattern = `[A-Z][A-Z0-9_]*-[0-9]+`
	DefaultAPIKeyEnv    = "BACKLOG_API_KEY"
	DefaultCacheTTL     = 24 * time.Hour
)

type Config struct {
	BaseURL      string
	APIKeyEnv    string
	IssuePattern string
	CacheTTL     time.Duration
}

func Load() (*Config, error) {
	baseURL, ok, err := git.ConfigGet("backlog.baseUrl")
	if err != nil {
		return nil, err
	}
	if !ok || baseURL == "" {
		return nil, fmt.Errorf("backlog.baseUrl is not set (git config --global backlog.baseUrl https://your-space.backlog.jp)")
	}

	apiKeyEnv, _, err := git.ConfigGet("backlog.apiKeyEnv")
	if err != nil {
		return nil, err
	}
	if apiKeyEnv == "" {
		apiKeyEnv = DefaultAPIKeyEnv
	}

	pattern, _, err := git.ConfigGet("backlog.issuePattern")
	if err != nil {
		return nil, err
	}
	if pattern == "" {
		pattern = DefaultIssuePattern
	}

	ttlStr, _, err := git.ConfigGet("backlog.cacheTTL")
	if err != nil {
		return nil, err
	}
	ttl := DefaultCacheTTL
	if ttlStr != "" {
		ttl, err = time.ParseDuration(ttlStr)
		if err != nil {
			return nil, fmt.Errorf("invalid backlog.cacheTTL %q: %w", ttlStr, err)
		}
	}

	return &Config{
		BaseURL:      baseURL,
		APIKeyEnv:    apiKeyEnv,
		IssuePattern: pattern,
		CacheTTL:     ttl,
	}, nil
}

func (c *Config) APIKey() (string, error) {
	key := os.Getenv(c.APIKeyEnv)
	if key == "" {
		return "", fmt.Errorf("API key env var %q is empty", c.APIKeyEnv)
	}
	return key, nil
}
