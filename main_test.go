package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCacheFilePath(t *testing.T) {
	tests := []struct {
		name            string
		claudeConfigDir string
		want            string
	}{
		{
			name:            "no CLAUDE_CONFIG_DIR set",
			claudeConfigDir: "",
			want:            filepath.Join(os.TempDir(), "claudeline-usage.json"),
		},
		{
			name:            "custom config dir claude-personal",
			claudeConfigDir: "/Users/oa/.claude-personal",
			want:            filepath.Join(os.TempDir(), "claudeline-usage-81c94270.json"),
		},
		{
			name:            "custom config dir claude-work",
			claudeConfigDir: "/Users/oa/.claude-work",
			want:            filepath.Join(os.TempDir(), "claudeline-usage-1ef5702c.json"),
		},
		{
			name:            "windows config dir claude-personal",
			claudeConfigDir: `C:\Users\oa\.claude-personal`,
			want:            filepath.Join(os.TempDir(), "claudeline-usage-9b705f7c.json"),
		},
		{
			name:            "windows config dir claude-work",
			claudeConfigDir: `C:\Users\oa\.claude-work`,
			want:            filepath.Join(os.TempDir(), "claudeline-usage-34fd078b.json"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CLAUDE_CONFIG_DIR", tt.claudeConfigDir)
			got := cacheFilePath()
			if got != tt.want {
				t.Errorf("cacheFilePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKeychainServiceName(t *testing.T) {
	tests := []struct {
		name            string
		claudeConfigDir string
		want            string
	}{
		{
			name:            "no CLAUDE_CONFIG_DIR set",
			claudeConfigDir: "",
			want:            "Claude Code-credentials",
		},
		{
			name:            "custom config dir claude-personal",
			claudeConfigDir: "/Users/oa/.claude-personal",
			want:            "Claude Code-credentials-81c94270",
		},
		{
			name:            "custom config dir claude-work",
			claudeConfigDir: "/Users/oa/.claude-work",
			want:            "Claude Code-credentials-1ef5702c",
		},
		{
			name:            "windows config dir claude-personal",
			claudeConfigDir: `C:\Users\oa\.claude-personal`,
			want:            "Claude Code-credentials-9b705f7c",
		},
		{
			name:            "windows config dir claude-work",
			claudeConfigDir: `C:\Users\oa\.claude-work`,
			want:            "Claude Code-credentials-34fd078b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CLAUDE_CONFIG_DIR", tt.claudeConfigDir)
			got := keychainServiceName()
			if got != tt.want {
				t.Errorf("keychainServiceName() = %q, want %q", got, tt.want)
			}
		})
	}
}
