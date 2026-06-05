package config

import (
	"os/exec"
	"testing"
	"time"
)

func setupRepo(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	t.Setenv("GIT_CONFIG_SYSTEM", "/dev/null")
	for _, e := range []string{BaseURLEnv, APIKeyEnvEnv, IssuePatternEnv} {
		t.Setenv(e, "")
	}
	t.Chdir(t.TempDir())
	if out, err := exec.Command("git", "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
}

func setLocal(t *testing.T, key, val string) {
	t.Helper()
	if out, err := exec.Command("git", "config", "--local", key, val).CombinedOutput(); err != nil {
		t.Fatalf("git config %s=%s: %v: %s", key, val, err, out)
	}
}

func TestLoadAllSet(t *testing.T) {
	setupRepo(t)
	setLocal(t, "backlog.baseUrl", "https://x.backlog.jp")
	setLocal(t, "backlog.apiKeyEnv", "MY_PAT")
	setLocal(t, "backlog.issuePattern", `[A-Z]+-[0-9]+`)
	setLocal(t, "backlog.cacheTTL", "1h")
	t.Setenv("MY_PAT", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BaseURL != "https://x.backlog.jp" {
		t.Errorf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.APIKeyEnv != "MY_PAT" {
		t.Errorf("APIKeyEnv = %q", cfg.APIKeyEnv)
	}
	key, err := cfg.APIKey()
	if err != nil {
		t.Fatalf("APIKey: %v", err)
	}
	if key != "secret" {
		t.Errorf("APIKey = %q", key)
	}
	if cfg.IssuePattern != `[A-Z]+-[0-9]+` {
		t.Errorf("IssuePattern = %q", cfg.IssuePattern)
	}
	if cfg.CacheTTL != time.Hour {
		t.Errorf("CacheTTL = %v", cfg.CacheTTL)
	}
}

func TestLoadDefaults(t *testing.T) {
	setupRepo(t)
	setLocal(t, "backlog.baseUrl", "https://x.backlog.jp")
	t.Setenv("BACKLOG_API_KEY", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.IssuePattern != DefaultIssuePattern {
		t.Errorf("IssuePattern = %q, want default", cfg.IssuePattern)
	}
	if cfg.CacheTTL != DefaultCacheTTL {
		t.Errorf("CacheTTL = %v, want %v", cfg.CacheTTL, DefaultCacheTTL)
	}
}

func TestLoadMissingBaseURL(t *testing.T) {
	setupRepo(t)
	t.Setenv("BACKLOG_API_KEY", "secret")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for missing baseUrl")
	}
}

func TestAPIKeyEmptyErrors(t *testing.T) {
	setupRepo(t)
	setLocal(t, "backlog.baseUrl", "https://x.backlog.jp")
	t.Setenv("BACKLOG_API_KEY", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if _, err := cfg.APIKey(); err == nil {
		t.Fatal("expected error for empty API key env var")
	}
}

func TestLoadBaseURLFromEnv(t *testing.T) {
	setupRepo(t)
	t.Setenv(BaseURLEnv, "https://env.backlog.jp")
	t.Setenv("BACKLOG_API_KEY", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BaseURL != "https://env.backlog.jp" {
		t.Errorf("BaseURL = %q, want env value", cfg.BaseURL)
	}
}

func TestLoadEnvOverridesGitConfig(t *testing.T) {
	setupRepo(t)
	setLocal(t, "backlog.baseUrl", "https://config.backlog.jp")
	t.Setenv(BaseURLEnv, "https://env.backlog.jp")
	t.Setenv("BACKLOG_API_KEY", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.BaseURL != "https://env.backlog.jp" {
		t.Errorf("BaseURL = %q, env should win over git config", cfg.BaseURL)
	}
}

func TestLoadAPIKeyEnvFromEnv(t *testing.T) {
	setupRepo(t)
	setLocal(t, "backlog.baseUrl", "https://x.backlog.jp")
	t.Setenv(APIKeyEnvEnv, "MY_PAT")
	t.Setenv("MY_PAT", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.APIKeyEnv != "MY_PAT" {
		t.Errorf("APIKeyEnv = %q, want env override", cfg.APIKeyEnv)
	}
	key, err := cfg.APIKey()
	if err != nil {
		t.Fatalf("APIKey: %v", err)
	}
	if key != "secret" {
		t.Errorf("APIKey = %q", key)
	}
}

func TestLoadIssuePatternFromEnv(t *testing.T) {
	setupRepo(t)
	setLocal(t, "backlog.baseUrl", "https://x.backlog.jp")
	t.Setenv(IssuePatternEnv, `XX-[0-9]+`)
	t.Setenv("BACKLOG_API_KEY", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.IssuePattern != `XX-[0-9]+` {
		t.Errorf("IssuePattern = %q, want env override", cfg.IssuePattern)
	}
}

func TestLoadInvalidTTL(t *testing.T) {
	setupRepo(t)
	setLocal(t, "backlog.baseUrl", "https://x.backlog.jp")
	setLocal(t, "backlog.cacheTTL", "not-a-duration")
	t.Setenv("BACKLOG_API_KEY", "secret")

	if _, err := Load(); err == nil {
		t.Fatal("expected error for invalid cacheTTL")
	}
}
