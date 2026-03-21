package main

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fredrikaverpil/pocket/pk"
	"github.com/fredrikaverpil/pocket/pk/repopath"
	"github.com/fredrikaverpil/pocket/pk/run"
)

// RenderFlags holds flags for the Render task.
type RenderFlags struct {
	JSON string `flag:"json" usage:"path to a specific testdata JSON file"`
}

// Render runs testdata payloads through claudeline for manual inspection.
// With -json, renders a single file. Without it, renders all testdata/stdin_*.json files.
var Render = &pk.Task{
	Name:  "render",
	Usage: "render statusline from testdata stdin payloads",
	Flags: RenderFlags{},
	Do: func(ctx context.Context) error {
		flags := run.GetFlags[RenderFlags](ctx)
		dir := repopath.FromGitRoot("")

		var files []string
		if flags.JSON != "" {
			files = []string{flags.JSON}
		} else {
			var err error
			files, err = filepath.Glob(filepath.Join(dir, "testdata", "stdin_*.json"))
			if err != nil {
				return fmt.Errorf("glob testdata: %w", err)
			}
			if len(files) == 0 {
				return fmt.Errorf("no testdata/stdin_*.json files found")
			}
		}

		for _, f := range files {
			run.Printf(ctx, ":: %s\n", filepath.Base(f))
			cmd := exec.CommandContext(ctx, "sh", "-c", fmt.Sprintf("cat %s | go run .", f))
			cmd.Dir = dir
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return err
			}
			fmt.Println()
		}
		return nil
	},
}

// CaptureFlags holds flags for the Capture task.
type CaptureFlags struct {
	ConfigDir string `flag:"config-dir" usage:"Claude config directory (e.g. ~/.claude-work)"`
}

// Capture extracts the latest stdin payload from a profile's debug log and saves
// it as a testdata file.
//
// The config-dir flag determines which profile to capture from:
//   - Default (~/.claude): debug log at /tmp/claudeline/debug.log
//   - Custom (e.g. ~/.claude-work): debug log with a hash suffix,
//     e.g. /tmp/claudeline/debug-fbb791ba.log
//
// For each profile, the task:
//  1. Extracts the last "raw stdin:" line (the most recent payload from Claude Code)
//  2. Resolves the subscription plan from the profile's credentials (Keychain → file fallback)
//  3. Parses the Claude Code version and model from the payload
//  4. Sanitizes sensitive fields (paths, cost)
//  5. Saves the payload as testdata/stdin_v<version>_<plan>_<model>.json
//
// Prerequisites: claudeline must be running with -debug enabled for the profile
// you want to capture.
var Capture = &pk.Task{
	Name:  "capture",
	Usage: "save latest debug log stdin payload as testdata",
	Flags: CaptureFlags{ConfigDir: "~/.claude"},
	Do: func(ctx context.Context) error {
		flags := run.GetFlags[CaptureFlags](ctx)
		dir := repopath.FromGitRoot("")

		// Expand ~ in the config dir path.
		configDir := expandHome(flags.ConfigDir)

		// Determine the debug log suffix. The default ~/.claude profile
		// has no CLAUDE_CONFIG_DIR set, so it uses no suffix.
		var suffix string
		defaultDir := filepath.Join(mustUserHomeDir(), ".claude")
		if configDir != defaultDir {
			h := sha256.Sum256([]byte(configDir))
			suffix = fmt.Sprintf("-%x", h[:4])
		}

		// Read the debug log.
		logPath := filepath.Join(claudelineCacheDir(), "debug"+suffix+".log")
		data, err := os.ReadFile(logPath)
		if err != nil {
			return fmt.Errorf("read debug log %s: %w — run claudeline with -debug first", logPath, err)
		}

		// Find the last raw stdin line.
		var rawJSON string
		for _, line := range strings.Split(string(data), "\n") {
			if _, after, ok := strings.Cut(line, "raw stdin: "); ok {
				rawJSON = after
			}
		}
		if rawJSON == "" {
			return fmt.Errorf("no 'raw stdin:' line found in %s", logPath)
		}

		// Parse version and model.
		var payload struct {
			Version string `json:"version"`
			Model   struct {
				ID string `json:"id"`
			} `json:"model"`
		}
		if err := json.Unmarshal([]byte(rawJSON), &payload); err != nil {
			return fmt.Errorf("parse stdin JSON: %w", err)
		}
		if payload.Version == "" {
			return fmt.Errorf("version field is empty in payload")
		}

		// Resolve plan from this profile's credentials.
		plan, err := resolvePlanName(ctx, configDir)
		if err != nil {
			return fmt.Errorf("resolve plan: %w", err)
		}

		// Sanitize and pretty-print.
		var m map[string]any
		if err := json.Unmarshal([]byte(rawJSON), &m); err != nil {
			return fmt.Errorf("parse JSON: %w", err)
		}
		sanitizePayload(m)
		formatted, err := json.MarshalIndent(m, "", "  ")
		if err != nil {
			return fmt.Errorf("format JSON: %w", err)
		}

		model := modelShortName(payload.Model.ID)
		filename := fmt.Sprintf("stdin_v%s_%s_%s.json", payload.Version, plan, model)
		outPath := filepath.Join(dir, "testdata", filename)
		if err := os.MkdirAll(filepath.Join(dir, "testdata"), 0o755); err != nil {
			return fmt.Errorf("create testdata dir: %w", err)
		}
		if err := os.WriteFile(outPath, append(formatted, '\n'), 0o644); err != nil {
			return fmt.Errorf("write testdata: %w", err)
		}

		run.Printf(ctx, "saved %s\n", outPath)
		return nil
	},
}

// expandHome replaces a leading ~ with the user's home directory.
func expandHome(path string) string {
	if strings.HasPrefix(path, "~/") || path == "~" {
		return filepath.Join(mustUserHomeDir(), path[1:])
	}
	return path
}

// mustUserHomeDir returns the user's home directory or panics.
func mustUserHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("get home dir: %v", err))
	}
	return home
}

// sanitizePayload removes or replaces sensitive fields in a stdin payload map
// before writing to testdata. Mutates the map in place.
func sanitizePayload(m map[string]any) {
	// Identity.
	setNestedString(m, "sanitized-session-id", "session_id")

	// Paths.
	setNestedString(m, "/sanitized/transcript.jsonl", "transcript_path")
	setNestedString(m, "/sanitized/project", "cwd")
	if ws, ok := m["workspace"].(map[string]any); ok {
		setNestedString(ws, "/sanitized/project", "current_dir")
		setNestedString(ws, "/sanitized/project", "project_dir")
	}

	// Cost — cumulative spending and session timing.
	if cost, ok := m["cost"].(map[string]any); ok {
		cost["total_cost_usd"] = 0
		cost["total_duration_ms"] = 0
		cost["total_api_duration_ms"] = 0
		cost["total_lines_added"] = 0
		cost["total_lines_removed"] = 0
	}
}

// setNestedString sets a key in a map to the given string value.
func setNestedString(m map[string]any, value, key string) {
	if _, ok := m[key]; ok {
		m[key] = value
	}
}

// claudelineCacheDir mirrors claudeline's cacheDir() — /tmp/claudeline on non-Windows.
func claudelineCacheDir() string {
	base := "/tmp"
	if runtime.GOOS == "windows" {
		base = os.TempDir()
	}
	return filepath.Join(base, "claudeline")
}

// resolvePlanName reads credentials for a specific config dir and returns the
// lowercase plan name. It tries the macOS Keychain first, falling back to the
// credentials file on disk.
func resolvePlanName(ctx context.Context, configDir string) (string, error) {
	type creds struct {
		ClaudeAiOauth struct {
			SubscriptionType string `json:"subscriptionType"`
		} `json:"claudeAiOauth"`
	}

	// Compute the keychain suffix for this config dir.
	var keychainSuffix string
	defaultDir := filepath.Join(mustUserHomeDir(), ".claude")
	if configDir != defaultDir {
		h := sha256.Sum256([]byte(configDir))
		keychainSuffix = fmt.Sprintf("-%x", h[:4])
	}

	// Try macOS keychain first.
	if runtime.GOOS == "darwin" {
		serviceName := "Claude Code-credentials" + keychainSuffix
		ctx, cancel := context.WithTimeout(ctx, 5_000_000_000) // 5s
		defer cancel()
		out, err := exec.CommandContext(ctx,
			"/usr/bin/security", "find-generic-password",
			"-s", serviceName, "-w",
		).Output()
		if err == nil {
			var c creds
			if err := json.Unmarshal(out, &c); err == nil {
				if plan := planName(c.ClaudeAiOauth.SubscriptionType); plan != "" {
					return plan, nil
				}
			}
		}
	}

	// File fallback.
	data, err := os.ReadFile(filepath.Join(configDir, ".credentials.json"))
	if err != nil {
		return "", fmt.Errorf("read credentials: %w", err)
	}
	var c creds
	if err := json.Unmarshal(data, &c); err != nil {
		return "", fmt.Errorf("parse credentials: %w", err)
	}
	plan := planName(c.ClaudeAiOauth.SubscriptionType)
	if plan == "" {
		return "", fmt.Errorf("unknown subscription type: %q", c.ClaudeAiOauth.SubscriptionType)
	}
	return plan, nil
}

// planName mirrors claudeline's planName() but returns lowercase for filenames.
func planName(subType string) string {
	lower := strings.ToLower(subType)
	switch {
	case strings.Contains(lower, "max"):
		return "max"
	case strings.Contains(lower, "pro"):
		return "pro"
	case strings.Contains(lower, "team"):
		return "team"
	case strings.Contains(lower, "enterprise"):
		return "enterprise"
	default:
		return ""
	}
}

// modelShortName extracts a short model name from a Claude model ID.
// e.g. "claude-opus-4-6[1m]" → "opus", "claude-sonnet-4-6" → "sonnet".
func modelShortName(id string) string {
	for _, name := range []string{"opus", "sonnet", "haiku"} {
		if strings.Contains(id, name) {
			return name
		}
	}
	s := strings.ReplaceAll(id, "[", "")
	s = strings.ReplaceAll(s, "]", "")
	return s
}
