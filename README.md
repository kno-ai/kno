# kno

A knowledge vault for your AI conversations.

Your AI conversations are disconnected — each one starts from scratch.
The debugging breakthrough, the design decision, the config that finally
worked — none of it carries forward. kno pays attention so you don't
have to. It notices when something worth preserving happens, offers to
capture it, and loads relevant context into future sessions automatically.
Your knowledge compounds across every conversation.

**[Project page](https://kno-ai.github.io/kno/)** — examples, docs, and getting started.

## Quick start

```bash
brew tap kno-ai/tap
brew install kno
kno setup
```

Restart Claude Desktop. kno is immediately active.

## How it works

kno listens as you work. It watches for **knowledge checkpoints** —
decisions, debugging insights, designs that settled — and offers to
capture them. When you start a session on a familiar topic, it recognizes
the overlap and offers to load your existing context.

Periodically, you run `/kno.curate` to synthesize your captures into
**page documents** — living, readable files that reflect everything you've
learned about a subject.

The loop is capture, curate, load — and kno initiates most of it for you.
Slash commands (`/kno.capture`, `/kno.load`, `/kno.curate`, `/kno.page`,
`/kno.status`) are available for explicit control.

## Your vault

Your vault is just a folder of markdown files — no database, no cloud
service, no lock-in. Sync it with git, Dropbox, iCloud — or browse it
in Obsidian alongside your other notes. kno works with Claude Desktop
today, with more AI clients coming soon.

```
~/kno/
  config.toml
  notes/
    20260305-rds-slow-query-debugging/
      content.md       # structured session summary
      meta.json        # title, tags, summary, curate status
    20260301-onboarding-handoff-failures/
      content.md
      meta.json
  pages/
    aws-infrastructure.md          # living knowledge document
    aws-infrastructure.meta.json   # name, summary, last_curated_at
    customer-onboarding.md
    customer-onboarding.meta.json
```

## Documentation

- [User Guide](docs/design/kno-guide.md) — getting started, awareness, and vault management
- [Architecture](docs/design/kno-knowledge-architecture.md) — design principles, layers, and knowledge model
- [Skills Reference](docs/design/kno-skills.md) — awareness behavior and slash command details
- [CLI Reference](docs/design/kno-cli.md) — complete command specification

## License

[MIT](LICENSE)
