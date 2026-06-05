package backlog

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetIssueHappyPath(t *testing.T) {
	var gotPath, gotKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.URL.Query().Get("apiKey")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":42,"issueKey":"TEST-1","summary":"hello","description":"d"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "PAT-123")
	issue, err := c.GetIssue("TEST-1")
	if err != nil {
		t.Fatalf("GetIssue: %v", err)
	}
	if issue.ID != 42 || issue.IssueKey != "TEST-1" || issue.Summary != "hello" || issue.Description != "d" {
		t.Errorf("unexpected issue: %+v", issue)
	}
	if gotPath != "/api/v2/issues/TEST-1" {
		t.Errorf("path = %q", gotPath)
	}
	if gotKey != "PAT-123" {
		t.Errorf("apiKey query = %q", gotKey)
	}
}

func TestGetIssue401SurfacesBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errors":[{"message":"Authentication failure.","code":11}]}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "bad")
	_, err := c.GetIssue("TEST-1")
	if err == nil {
		t.Fatal("expected error on 401")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("err missing status: %v", err)
	}
	if !strings.Contains(err.Error(), "Authentication failure") {
		t.Errorf("err missing body: %v", err)
	}
}

func TestGetIssue404ReturnsErrNotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"errors":[{"message":"No such issue.","code":6}]}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k")
	_, err := c.GetIssue("MISS-1")
	if err == nil {
		t.Fatal("expected error on 404")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("err = %v, want ErrNotFound", err)
	}
}

func TestGetIssueMalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "k")
	_, err := c.GetIssue("TEST-1")
	if err == nil {
		t.Fatal("expected error on bad JSON")
	}
	if !strings.Contains(err.Error(), "decode") {
		t.Errorf("err = %v", err)
	}
}

