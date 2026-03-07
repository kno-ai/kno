# kno

**A knowledge vault for your AI conversations.** &nbsp; [View on GitHub](https://github.com/kno-ai/kno)

Every time you close a chat with Claude, the insights from that session
disappear. kno fixes that. You capture what you learned, curate it into
living page documents, and load the right context into your next session.
The knowledge compounds because you curate it — 30 seconds of attention
per session turns scattered conversations into documents you trust.

---

## The knowledge loop

Three commands. One habit.

**`/kno.capture`** — At the end of a session, kno reviews the conversation and
proposes a structured summary with title, tags, and key points. You confirm,
edit, or skip — that moment of curation is what makes the knowledge findable
later. Use #hashtags to steer tags directly.

**`/kno.curate`** — Periodically, kno reads your saved sessions and folds
them into page documents. Each page reflects everything you've learned about
a subject — organized the way you think about it, following guidance you've
written.

**`/kno.load`** — At the start of a session, kno finds relevant pages and
recent sessions and injects them into the conversation. You don't re-explain
your setup. You don't rediscover dead ends. You pick up where your knowledge
left off.

Each pass through the loop makes the next one better.

---

## What it feels like

```
/kno.capture — #aws #rds, the parameter group fix was the key thing

Here's what I'll capture from this session:

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
  Pages (1):     Payment Processing — last curated 2 weeks ago
  Sessions (2,   matched by tags: payments, mysql, connection-pool):
                 ACH return handling (3 days ago)
                 MySQL connection pool (1 week ago)

Load all 3? [yes / pick / skip]
```

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
kno setup
```

Restart Claude Desktop after setup. Five slash commands appear automatically:
`/kno.capture`, `/kno.load`, `/kno.curate`, `/kno.page`, `/kno.status`.

See the [User Guide](design/kno-guide) for the full walkthrough.

---

## Documentation

- [User Guide](design/kno-guide) — full walkthrough, tips, and CLI reference
- [Architecture](design/kno-knowledge-architecture) — design principles and mental model
- [Skills Reference](design/kno-skills) — how the slash commands work
- [CLI Reference](design/kno-cli) — complete command specification

---

## Project status

kno is in active development. The knowledge loop — capture, curate, load —
is functional end-to-end with Claude Desktop integration.

Your vault is just a folder of markdown files. Sync it with git, Dropbox,
iCloud — or browse it in Obsidian alongside your other notes.

<p style="margin-top: 3rem; color: #666; font-size: 0.85rem;">
  MIT License | <a href="https://github.com/kno-ai/kno">GitHub</a>
</p>
