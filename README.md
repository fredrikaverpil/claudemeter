# claudemeter

A minimal, opinionated and non-configurable, Claude Code status line.

```sh
[Opus 4.6 | Team] │ █░░░░░░░░░ 18% │ ████░░░░░░ 40% (2h 40m) │ ██░░░░░░░░ 27% (95h 40m)
```

## Project

A Claude Code statusline plugin written in Go. It displays the current AI model,
subscription plan, context window usage, and 5-hour/7-day quota usage as
ANSI-colored progress bars. Zero external dependencies (stdlib only).

## Build and Run

```bash
go build -o claudemeter .
```

No Makefile, no test suite, no linter config. The binary reads JSON from stdin
(provided by Claude Code) and writes a single ANSI-colored line to stdout.

## Architecture

Single-file (`main.go`), single-package (`main`) design.

**Data flow:** stdin JSON → parse input + read credentials → fetch usage
(cached) → render ANSI output → stdout

Key components:

- **Credential resolution:** macOS Keychain (`security find-generic-password`)
  first, falls back to `~/.claude/.credentials.json`. Failure is non-fatal
  (usage bars are omitted).
- **Usage API:** `GET https://api.anthropic.com/api/oauth/usage` with OAuth
  bearer token. 5-second HTTP timeout.
- **File-based cache:** `/tmp/claudemeter-usage.json` with 60s TTL on
  success, 15s TTL on failure.
- **Progress bars:** 10-char width using `█`/`░` with color thresholds
  (green/yellow/red for context; blue/magenta/red for quota).

## Plugin Metadata

`/.claude-plugin/plugin.json` — plugin identity and version for the Claude Code
marketplace. `/commands/setup.md` — slash command that configures
`~/.claude/settings.json` to use this binary.

## References

- [Create Claude plugins](https://code.claude.com/docs/en/plugins)
- [Customize your status line](https://code.claude.com/docs/en/statusline)
