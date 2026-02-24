# Setup claudeline

Install and configure the claudeline plugin as your Claude Code statusline.

Re-running this command will update claudeline to the latest version.

## Steps

1. **Download the latest pre-built binary from GitHub releases.**

Detect OS and architecture:

```bash
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
esac
```

Download and install:

```bash
mkdir -p ~/.local/bin
curl -fsSL "https://github.com/fredrikaverpil/claudeline/releases/latest/download/claudeline_${OS}_${ARCH}.tar.gz" | tar -xz -C ~/.local/bin claudeline
chmod +x ~/.local/bin/claudeline
```

If the download fails (e.g. `curl` is not available, no internet, or the
platform is unsupported), fall back to `go install`:

```bash
go install github.com/fredrikaverpil/claudeline@latest
```

This requires Go to be installed with `$GOPATH/bin` in `$PATH`.

2. **Verify the binary is available.**

Check if the binary works by running it directly:

```bash
~/.local/bin/claudeline -version
```

If the binary was installed via `go install`, check with `which claudeline`
instead.

3. **Update `~/.claude/settings.json`** — set the `statusLine` field.

If `~/.local/bin` is in the user's `$PATH`, use:

```json
{
  "statusLine": {
    "type": "command",
    "command": "claudeline"
  }
}
```

Otherwise, use the full path:

```json
{
  "statusLine": {
    "type": "command",
    "command": "~/.local/bin/claudeline"
  }
}
```

Preserve all other fields in the file.

4. **Confirm** the change was made and tell the user to restart their Claude
   Code session for the statusline to take effect.
