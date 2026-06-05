package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)

var issues = map[string]struct {
	ID      int64
	Summary string
}{
	"TEST-1":   {1, "Sample test issue"},
	"BUG-42":   {42, "Fake bug ticket"},
	"PROJ-100": {100, "Implement new feature"},
	"MULTI-1":  {1001, "First match wins"},
	"OTHER-2":  {1002, "Second key in a multi-key branch"},
}

func main() {
	addr := flag.String("addr", ":18080", "listen address")
	flag.Parse()

	http.HandleFunc("/api/v2/issues/", handleIssue)
	http.HandleFunc("/api/v2/space", func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"spaceKey": "fake", "name": "fake-backlog"})
	})

	log.Printf("fake-backlog listening on %s", *addr)
	log.Printf("known issues: TEST-1, BUG-42, PROJ-100, MULTI-1, OTHER-2")
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func handleIssue(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimPrefix(r.URL.Path, "/api/v2/issues/")
	apiKey := r.URL.Query().Get("apiKey")
	log.Printf("%s %s key=%q apiKey=%s", r.Method, r.URL.Path, key, redact(apiKey))

	w.Header().Set("Content-Type", "application/json")

	if apiKey == "" {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"errors":[{"message":"Authentication failure.","code":11}]}`)
		return
	}

	issue, ok := issues[key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"errors":[{"message":"No such issue %s","code":6}]}`, key)
		return
	}

	json.NewEncoder(w).Encode(map[string]any{
		"id":       issue.ID,
		"issueKey": key,
		"summary":  issue.Summary,
	})
}

func redact(s string) string {
	if len(s) <= 4 {
		return strings.Repeat("*", len(s))
	}
	return s[:2] + strings.Repeat("*", len(s)-4) + s[len(s)-2:]
}
