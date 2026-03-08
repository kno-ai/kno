# kno

**A knowledge vault for your AI conversations.** &nbsp; [View on GitHub](https://github.com/kno-ai/kno)

Your AI conversations are disconnected. Each one starts from scratch —
you re-explain your setup, rediscover prior decisions, and lose the
context you built last time. kno pays attention so you don't have to.
It notices when something worth preserving happens and offers to capture
it. It recognizes when you're working on a familiar topic and offers to
load what you already know. Your knowledge compounds across sessions —
not because you remember to do anything, but because kno listens in
every conversation.

---

## How it works

Start every chat with `/kno`. kno checks your vault, shows your pages,
and offers to load relevant context. Say yes or just start working —
kno stays aware from there.

**kno notices knowledge checkpoints** — a decision reached, a root cause
found, a design that settled — and offers to capture them:

> "That's a good one — want to add it to your vault?"

You confirm, and it's saved with a title, summary, and tags. Ten seconds.
The cost of preserving an insight drops from "remember to save, context-
switch to a notes tool, decide what to write, format it" to "say yes."

**kno suggests relevant context** — when your conversation overlaps with
existing vault knowledge, it offers to load it:

> "kno has notes on this — want to load your AWS Infrastructure page?"

Your session starts informed instead of cold. No re-explaining your
setup, no rediscovering decisions you already made.

**kno prompts you to curate** — when captures accumulate, kno suggests
folding them into living page documents. Each page reflects everything
you've learned about a subject, organized the way you think about it.
If you've configured a publish target, pages are automatically published
with frontmatter and tags — your knowledge becomes browsable in Obsidian
or any markdown viewer. Curate is the one step that stays intentional,
where you decide what matters.

**The loop compounds.** Each capture feeds curate. Each curated page
makes load faster and richer. Better loads mean better sessions, which
produce better captures. The more you use it, the more valuable it
becomes.

---

## What it feels like

Mid-conversation, after you've debugged a tricky issue:

> **kno:** "That root cause was non-obvious — want to add it to your
> vault?"
>
> **You:** "yes"
>
> **kno:** Here's what kno will capture:
>
>   Title:    RDS slow query debugging
>   Summary:  Query planner regression after minor version upgrade.
>             Fixed by pinning parameter group.
>   Tags:     aws, rds, performance
>
> Save this? [yes / edit / skip]

Two weeks later, you type `/kno` at the start of a new session:

> **kno:** Your vault has **CNC Machine Maintenance** and **AWS
> Infrastructure**. Want me to load any of these?
>
> **You:** "load the CNC page"
>
> *Session starts with full context from prior work.*

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

Restart Claude Desktop after setup. Enter `/kno` in a chat to connect —
kno shows your pages and offers to load relevant context. From there,
it stays aware: noticing knowledge checkpoints and suggesting loads as
topics come up.

See the [User Guide](design/kno-guide) for the full walkthrough.

---

## Your vault is just files

No database, no cloud service, no lock-in. Your vault is a folder of
markdown files and a TOML config. Sync it with git or Dropbox, back it
up however you back up everything else. You can read every file kno writes.

Publish curated pages to Obsidian or any markdown viewer that supports
frontmatter — your pages get YAML metadata (title, tags, summary, dates)
and cross-references become wikilinks:

```sh
kno setup --publish ~/obsidian/kno
```

Pages are published automatically after every curate. You can also run
`kno publish` manually at any time.

## Works with Claude Desktop

kno works with Claude Desktop today. Support for more AI clients is
coming soon — your knowledge won't be locked to one tool.

---

## Project status

kno is in active development. The knowledge loop — capture, curate, load
— is functional end-to-end.

## Documentation

- [User Guide](design/kno-guide) — getting started and vault management
- [Architecture](design/kno-knowledge-architecture) — design principles, layers, and knowledge model
- [Skills Reference](design/kno-skills) — skill behavior and slash command details
- [CLI Reference](design/kno-cli) — complete command specification

<p style="margin-top: 3rem; color: #666; font-size: 0.85rem;">
  MIT License | <a href="https://github.com/kno-ai/kno">GitHub</a>
</p>
