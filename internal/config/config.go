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

	BaseURLEnv      = "BACKLOG_BASE_URL"
	APIKeyEnvEnv    = "BACKLOG_API_KEY_ENV"
	IssuePatternEnv = "BACKLOG_ISSUE_PATTERN"
)

type Config struct {
	BaseURL      string
	APIKeyEnv    string
	IssuePattern string
	CacheTTL     time.Duration
}

func resolve(envName, configKey, fallback string) (string, error) {
	if v := os.Getenv(envName); v != "" {
		return v, nil
	}
	v, _, err := git.ConfigGet(configKey)
	if err != nil {
		return "", err
	}
	if v != "" {
		return v, nil
	}
	return fallback, nil
}

func Load() (*Config, error) {
	baseURL, err := resolve(BaseURLEnv, "backlog.baseUrl", "")
	if err != nil {
		return nil, err
	}
	if baseURL == "" {
		return nil, fmt.Errorf("backlog.baseUrl is not set (git config --global backlog.baseUrl https://your-space.backlog.jp, or export %s=...)", BaseURLEnv)
	}

	apiKeyEnv, err := resolve(APIKeyEnvEnv, "backlog.apiKeyEnv", DefaultAPIKeyEnv)
	if err != nil {
		return nil, err
	}

	pattern, err := resolve(IssuePatternEnv, "backlog.issuePattern", DefaultIssuePattern)
	if err != nil {
		return nil, err
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

func IssuePattern() (string, error) {
	return resolve(IssuePatternEnv, "backlog.issuePattern", DefaultIssuePattern)
}

func (c *Config) APIKey() (string, error) {
	key := os.Getenv(c.APIKeyEnv)
	if key == "" {
		return "", fmt.Errorf("API key env var %q is empty", c.APIKeyEnv)
	}
	return key, nil
}
