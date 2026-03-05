---
layout: default
---

# kno

**Local-first knowledge capture for AI conversations.**

kno captures the insights from your AI conversations before they disappear into chat history. It stores structured knowledge artifacts on your local filesystem — plain files, fully under your control.

---

## Using kno with Claude

kno integrates directly with Claude Desktop via MCP (Model Context Protocol). Once installed, two slash commands become available in your Claude conversations:

### `/kno.capture`

At the end of a productive conversation, invoke `/kno.capture`. Claude will:

- Review the conversation for key decisions, insights, and outcomes
- Generate a structured summary with TL;DR, decisions, key points, and next steps
- Save it as a capture in your local vault

### `/kno.load`

Start a new conversation with context from a previous one. `/kno.load` lists your recent captures and lets you pick one to load into the current session.

### Setup

```sh
# Install
go install github.com/kno-ai/kno/cmd/kno@latest

# Initialize vault and register with Claude Desktop
kno setup
```

Restart Claude Desktop after setup. The `/kno.capture` and `/kno.load` commands will appear automatically.

---

## CLI Reference

kno also works as a standalone command-line tool for scripting, automation, and direct use.

### `kno setup`

Initialize a vault and optionally register with Claude Desktop.

```sh
kno setup                        # defaults: ~/kno vault, registers with Claude
kno setup --vault ~/notes/kno    # custom vault path
kno setup --no-mcp               # skip MCP registration
```

### `kno capture`

Capture text from stdin, clipboard, or inline.

```sh
echo "meeting notes..." | kno capture --stdin --title "Standup notes"
kno capture --clipboard --title "Design thoughts" --meta topic=architecture
kno capture notes.md --title "Session notes" --meta project=kno
```

Sources are mutually exclusive: use `--stdin`, `--clipboard`, or pass a file path. The `--meta key=value` flag is repeatable and stores arbitrary metadata alongside the capture.

### `kno list`

List recent captures.

```sh
kno list          # 10 most recent
kno list -n 5     # last 5
```

### `kno show`

Display a capture's metadata and content.

```sh
kno show 20250305-143022-sqs-throughput-tuning
```

### `kno version`

Print the installed version.

```sh
kno version
```

### `kno mcp`

Start the MCP server (used internally by Claude Desktop).

```sh
kno mcp
```

---

## How it works

Each capture is stored as a directory containing two files:

```
~/kno/captures/
  20250305-143022-sqs-throughput-tuning/
    capture.md     # human-readable markdown
    meta.json      # structured metadata
```

- **capture.md** is clean markdown with no frontmatter — open it in any editor or viewer
- **meta.json** holds id, timestamp, source, status, and custom metadata

The vault is just a directory. Sync it however you like — git, Dropbox, iCloud, or don't.

---

## Project status

kno is in early development. The capture pipeline and Claude Desktop integration are functional. Knowledge distillation and synthesis features are planned.

<p style="margin-top: 3rem; color: #666; font-size: 0.85rem;">
  MIT License | <a href="https://github.com/kno-ai/kno">GitHub</a>
</p>
