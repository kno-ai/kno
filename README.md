# kno

A knowledge vault for your AI conversations.

Every time you close a chat with Claude, the insights from that session
disappear. kno fixes that. You save what you learned, distill it into
living page documents, and load the right context into your next session.
The knowledge compounds because you curate it — 30 seconds of attention
per session is what turns scattered conversations into documents you'd
hand to a colleague.

## The knowledge loop

Three commands. One habit.

- **`/kno.save`** — End of session. Review and confirm a structured summary with tags.
- **`/kno.distill`** — Periodically. Synthesize sessions into living page documents.
- **`/kno.load`** — Start of session. Load relevant knowledge before you begin.

Each pass through the loop makes the next one better. Sessions feed distill.
Distilled pages make load richer. Better load means better sessions.

## Quick start

```bash
# Homebrew
brew tap kno-ai/tap
brew install kno

# Or from source
go install github.com/kno-ai/kno/cmd/kno@latest
```

```bash
kno setup
```

Restart Claude Desktop after setup. Five slash commands appear automatically:
`/kno.save`, `/kno.load`, `/kno.distill`, `/kno.page`, `/kno.status`.

## What it feels like

```
/kno.save — #aws #rds, the parameter group fix was the key thing

Here's what I'll save from this session:

  Title:    RDS slow query debugging
  Summary:  Query planner regression after minor version upgrade.
            Fixed by pinning parameter group.
  Tags:     aws, rds, performance

Save this? [yes / edit / skip]
```

Two weeks later, in a new session:

```
/kno.load

What are you working on?

> debugging a connection pool issue in our payment service

Found:
  Pages (1):     Payment Processing — last distilled 2 weeks ago
  Sessions (2,   matched by tags: payments, mysql, connection-pool):
                 ACH return handling (3 days ago)
                 MySQL connection pool (1 week ago)

Load all 3? [yes / pick / skip]
```

Your vault is just a folder of markdown files. Sync it with git, Dropbox,
iCloud — or browse it in Obsidian alongside your other notes.

## Documentation

- [User Guide](docs/design/kno-guide.md) — full walkthrough and tips
- [Architecture](docs/design/kno-knowledge-architecture.md) — mental model and design principles
- [Skills Reference](docs/design/kno-skills.md) — how the slash commands work
- [CLI Reference](docs/design/kno-cli.md) — complete command specification

## License

[MIT](LICENSE)
