package issuekey

import (
	"fmt"
	"regexp"
)

func Extract(branch, pattern string) (string, bool, error) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return "", false, fmt.Errorf("compile issue pattern %q: %w", pattern, err)
	}
	m := re.FindString(branch)
	if m == "" {
		return "", false, nil
	}
	return m, true, nil
}
