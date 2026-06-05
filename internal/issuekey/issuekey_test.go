package issuekey

import "testing"

func TestExtract(t *testing.T) {
	const defaultPattern = `[A-Z][A-Z0-9_]*-[0-9]+`

	tests := []struct {
		name    string
		branch  string
		pattern string
		wantKey string
		wantOK  bool
		wantErr bool
	}{
		{"simple match", "TEST-1-foo", defaultPattern, "TEST-1", true, false},
		{"nested with slash prefix", "feature/PROJ-100-nested", defaultPattern, "PROJ-100", true, false},
		{"first match wins", "MULTI-1-and-OTHER-2-keys", defaultPattern, "MULTI-1", true, false},
		{"underscore allowed in project key", "MY_PROJ-7-x", defaultPattern, "MY_PROJ-7", true, false},
		{"digits allowed in project key", "P2-9-x", defaultPattern, "P2-9", true, false},
		{"no match", "no-key-branch", defaultPattern, "", false, false},
		{"lowercase rejected by default pattern", "lower-99-not-match", defaultPattern, "", false, false},
		{"invalid pattern errors", "TEST-1", `[`, "", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok, err := Extract(tt.branch, tt.pattern)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if ok != tt.wantOK {
				t.Errorf("ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.wantKey {
				t.Errorf("key = %q, want %q", got, tt.wantKey)
			}
		})
	}
}
