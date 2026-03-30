// Package git provides lightweight git helpers that read from the filesystem directly.
package git

import (
	"os"
	"strings"
)

// Branch returns the current git branch name, or "" if not in a git repo.
// It reads .git/HEAD directly (no subprocess).
func Branch() string {
	data, err := os.ReadFile(".git/HEAD")
	if err != nil {
		return ""
	}
	s := strings.TrimSpace(string(data))
	if after, ok := strings.CutPrefix(s, "ref: refs/heads/"); ok {
		return after
	}
	return "" // detached HEAD or bare repo
}
