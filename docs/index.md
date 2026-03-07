---
layout: default
---

# kno

**A knowledge vault for your AI conversations.**

Every time you close a chat with Claude, the insights from that session
disappear. kno fixes that. It saves what you learned, synthesizes it into
living page documents over time, and loads the right context into your next
session automatically. Knowledge compounds instead of evaporating.

---

## The knowledge loop

Three commands. One habit.

```
  /kno.save                /kno.distill              /kno.load
  ─────────                ────────────              ─────────
  End of session      →     Periodically        →     Start of session

  Save what you             Synthesize sessions       Load what's relevant
  learned before            into living page         before you start,
  you close the tab.        documents.                not after.
```

**`/kno.save`** — At the end of a session, kno reviews the conversation,
proposes a structured summary with title, tags, and key points, and saves it
to your vault. Ten seconds, and the session is preserved.

**`/kno.distill`** — Periodically, kno reads your saved sessions and folds
them into page documents. Each page reflects everything you've learned about
a subject — organized the way you think about it, following guidance you've
written.

**`/kno.load`** — At the start of a session, kno finds relevant pages and
recent sessions and injects them into the conversation. You don't re-explain
your setup. You don't rediscover dead ends. You pick up where your knowledge
left off.

Each pass through the loop makes the next one better.

---

## Getting started

```sh
# Install via Homebrew
brew tap kno-ai/tap
brew install kno

# Or from source
go install github.com/kno-ai/kno/cmd/kno@latest
```

```sh
# Initialize vault and register with Claude Desktop
kno setup
```

Restart Claude Desktop after setup. Five slash commands appear automatically:
`/kno.save`, `/kno.load`, `/kno.distill`, `/kno.page`, `/kno.status`.

See the [User Guide](design/kno-guide) for the full walkthrough.

---

## How it works

kno has three layers:

- **The CLI** owns your vault — a plain directory of markdown and JSON files
  on your filesystem. Deterministic, testable, no AI involved.
- **The MCP server** exposes the CLI to Claude Desktop as typed tools.
- **The skills** are the conversational layer — they interpret your intent,
  propose what to save, and guide you through the knowledge loop.

Your vault is just a folder. Sync it with git, Dropbox, iCloud, or don't.

```
~/kno/
  config.toml
  notes/
    20260305T142200Z-a3b1c2/
      content.md       # structured session summary
      meta.json        # title, tags, summary, distill status
  pages/
    b81e44/
      content.md       # living knowledge document
      meta.json        # name, last_distilled_at
```

---

## Design

- [User Guide](design/kno-guide) — full walkthrough, tips, and CLI reference
- [Architecture](design/kno-knowledge-architecture) — mental model, design principles, metadata
- [Skills Reference](design/kno-skills) — how the slash commands work and what they do
- [CLI Reference](design/kno-cli) — complete command specification

---

## Project status

kno is in active development. The knowledge loop — save, distill, load —
is functional end-to-end with Claude Desktop integration.

<p style="margin-top: 3rem; color: #666; font-size: 0.85rem;">
  MIT License | <a href="https://github.com/kno-ai/kno">GitHub</a>
</p>
