package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Sho2010/git-backlog/internal/backlog"
)

func newTestCache(t *testing.T, ttl time.Duration) *Cache {
	t.Helper()
	t.Setenv("XDG_CACHE_HOME", t.TempDir())
	c, err := New(ttl)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	return c
}

func TestPutGetRoundtrip(t *testing.T) {
	c := newTestCache(t, time.Hour)
	issue := &backlog.Issue{ID: 1, IssueKey: "TEST-1", Summary: "hello"}
	if err := c.Put(issue); err != nil {
		t.Fatalf("Put: %v", err)
	}
	got, hit, err := c.Get("TEST-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !hit {
		t.Fatal("expected cache hit")
	}
	if got.IssueKey != issue.IssueKey || got.Summary != issue.Summary {
		t.Errorf("got %+v, want %+v", got, issue)
	}
}

func TestGetMissWhenAbsent(t *testing.T) {
	c := newTestCache(t, time.Hour)
	_, hit, err := c.Get("NOPE-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if hit {
		t.Fatal("expected miss for absent key")
	}
}

func TestGetMissWhenExpired(t *testing.T) {
	c := newTestCache(t, time.Minute)
	issue := &backlog.Issue{IssueKey: "EXP-1", Summary: "old"}
	if err := c.Put(issue); err != nil {
		t.Fatalf("Put: %v", err)
	}
	old := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(c.path("EXP-1"), old, old); err != nil {
		t.Fatalf("Chtimes: %v", err)
	}
	_, hit, err := c.Get("EXP-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if hit {
		t.Fatal("expected miss for expired entry")
	}
}

func TestGetMissOnCorruptJSON(t *testing.T) {
	c := newTestCache(t, time.Hour)
	if err := os.WriteFile(c.path("BAD-1"), []byte("not json"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	_, hit, err := c.Get("BAD-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if hit {
		t.Fatal("expected miss for corrupt JSON")
	}
}

func TestPutIsAtomic(t *testing.T) {
	c := newTestCache(t, time.Hour)
	issue := &backlog.Issue{IssueKey: "ATOM-1", Summary: "x"}
	if err := c.Put(issue); err != nil {
		t.Fatalf("Put: %v", err)
	}
	entries, err := os.ReadDir(c.Dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Errorf("leftover tmp file: %s", e.Name())
		}
	}
}
