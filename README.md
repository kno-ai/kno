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

Restart Claude Desktop. Enter `/kno` in a chat to connect.

To publish curated pages to Obsidian or another markdown viewer:

```bash
kno setup --publish ~/obsidian/kno
```

## How it works

Start every chat with `/kno`. kno checks your vault and offers to load
relevant context. From there, it stays aware — noticing knowledge
checkpoints and offering to capture them, recognizing familiar topics
and suggesting loads.

Over time, kno prompts you to curate your captures into **page
documents** — living, readable files that reflect everything you've
learned about a subject. Pages can be published to Obsidian or any
markdown tool that supports frontmatter — your knowledge becomes
browsable outside of AI conversations.

The loop is capture, curate, load — and kno drives it. Slash commands
are there when you want explicit control.

## Your vault

Your vault is just a folder of markdown files — no database, no cloud
service, no lock-in. Sync it with git, Dropbox, iCloud — or publish
curated pages to Obsidian with frontmatter, tags, and wikilinks. kno
works with Claude Desktop today, with more AI clients coming soon.

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

- [User Guide](docs/design/kno-guide.md) — getting started and vault management
- [Architecture](docs/design/kno-knowledge-architecture.md) — design principles, layers, and knowledge model
- [Skills Reference](docs/design/kno-skills.md) — skill behavior and slash command details
- [CLI Reference](docs/design/kno-cli.md) — complete command specification

## License

[MIT](LICENSE)
