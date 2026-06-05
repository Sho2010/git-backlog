package git

import (
	"os/exec"
	"sort"
	"testing"
)

func setupRepo(t *testing.T) {
	t.Helper()
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	t.Setenv("GIT_CONFIG_SYSTEM", "/dev/null")
	t.Chdir(t.TempDir())
	for _, args := range [][]string{
		{"init", "-q", "-b", "main"},
		{"config", "user.email", "test@example.com"},
		{"config", "user.name", "test"},
		{"commit", "-q", "--allow-empty", "-m", "init"},
	} {
		if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v: %s", args, err, out)
		}
	}
}

func TestCurrentBranch(t *testing.T) {
	setupRepo(t)
	got, err := CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch: %v", err)
	}
	if got != "main" {
		t.Errorf("CurrentBranch = %q, want main", got)
	}
}

func TestLocalBranches(t *testing.T) {
	setupRepo(t)
	for _, b := range []string{"TEST-1-foo", "BUG-42-bar"} {
		if out, err := exec.Command("git", "branch", b).CombinedOutput(); err != nil {
			t.Fatalf("git branch %s: %v: %s", b, err, out)
		}
	}
	got, err := LocalBranches()
	if err != nil {
		t.Fatalf("LocalBranches: %v", err)
	}
	sort.Strings(got)
	want := []string{"BUG-42-bar", "TEST-1-foo", "main"}
	if len(got) != len(want) {
		t.Fatalf("LocalBranches = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("LocalBranches[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestLocalBranchesEmpty(t *testing.T) {
	t.Setenv("GIT_CONFIG_GLOBAL", "/dev/null")
	t.Setenv("GIT_CONFIG_SYSTEM", "/dev/null")
	t.Chdir(t.TempDir())
	if out, err := exec.Command("git", "init", "-q", "-b", "main").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v: %s", err, out)
	}
	got, err := LocalBranches()
	if err != nil {
		t.Fatalf("LocalBranches: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("LocalBranches before any commit = %v, want []", got)
	}
}

func TestConfigGet(t *testing.T) {
	setupRepo(t)
	if out, err := exec.Command("git", "config", "--local", "backlog.baseUrl", "https://x").CombinedOutput(); err != nil {
		t.Fatalf("git config: %v: %s", err, out)
	}

	v, ok, err := ConfigGet("backlog.baseUrl")
	if err != nil {
		t.Fatalf("ConfigGet: %v", err)
	}
	if !ok {
		t.Fatal("ok = false, want true")
	}
	if v != "https://x" {
		t.Errorf("value = %q", v)
	}
}

func TestConfigGetMissing(t *testing.T) {
	setupRepo(t)
	v, ok, err := ConfigGet("backlog.missing")
	if err != nil {
		t.Fatalf("ConfigGet: %v", err)
	}
	if ok {
		t.Errorf("ok = true for missing key (value=%q)", v)
	}
	if v != "" {
		t.Errorf("value = %q, want empty", v)
	}
}
