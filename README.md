# kno

A knowledge vault for your AI conversations.

Every time you close a chat with Claude, the insights from that session
disappear. kno fixes that. You capture what you learned, curate it into
living page documents, and load the right context into your next session.

**[Project page](https://kno-ai.github.io/kno/)** — examples, docs, and getting started.

## Quick start

```bash
brew tap kno-ai/tap
brew install kno
kno setup
```

Restart Claude Desktop. Five slash commands appear:
`/kno.capture`, `/kno.load`, `/kno.curate`, `/kno.page`, `/kno.status`.

## The knowledge loop

- **`/kno.capture`** — End of session. Review and confirm a structured summary with tags.
- **`/kno.curate`** — Periodically. Synthesize sessions into living page documents.
- **`/kno.load`** — Start of session. Load relevant knowledge before you begin.

Your vault is just a folder of markdown files. Sync it with git, Dropbox,
iCloud — or browse it in Obsidian alongside your other notes.

```
~/kno/
  config.toml
  notes/
    20260305-rds-slow-query-debugging/
      content.md       # structured session summary
      meta.json        # title, tags, summary, curate status
  pages/
    aws-infrastructure.md          # living knowledge document
    aws-infrastructure.meta.json   # name, last_curated_at
```

## Documentation

- [User Guide](docs/design/kno-guide.md) — full walkthrough and tips
- [Architecture](docs/design/kno-knowledge-architecture.md) — design principles and mental model
- [Skills Reference](docs/design/kno-skills.md) — how the slash commands work
- [CLI Reference](docs/design/kno-cli.md) — complete command specification

## License

[MIT](LICENSE)
