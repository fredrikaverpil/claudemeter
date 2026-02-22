# Setup claudeline

Configure the claudeline plugin as your Claude Code statusline.

## Prerequisites

- Go must be installed
- `$GOPATH/bin` must be in your `$PATH`

## Steps

1. Install the binary:

```bash
go install github.com/fredrikaverpil/claudeline@latest
```

2. Verify the binary is available:

```bash
which claudeline
```

If this fails, tell the user to add `$GOPATH/bin` to their `$PATH` and retry.

3. Read the current `~/.claude/settings.json` and update the `statusLine` field to:

```json
{
  "statusLine": {
    "type": "command",
    "command": "claudeline"
  }
}
```

Preserve all other fields in the file.

4. Confirm the change was made and tell the user to restart their Claude Code session for the statusline to take effect.
