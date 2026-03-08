# kno

A knowledge vault for your AI conversations.

Your AI conversations reset every session. The insight you reached, the
decision you worked through, the solution that finally clicked — none of
it carries forward. kno pays attention so you don't have to. It notices
when something worth preserving happens and offers to save it. When
you're revisiting familiar territory, kno offers to bring back what
you already know. Your knowledge compounds across every
conversation.

**[Getting started, examples, and docs →](https://kno-ai.github.io/kno/)**

---

## A conversation with kno

You've just tracked down a tricky issue. Before you move on:

> **kno:** That root cause was non-obvious — want to add it to your vault?
>
> **You:** yes
>
> **kno:** Here's what I'll save:
>
>   Title:    RDS slow query debugging
>   Summary:  Query planner regression after minor version upgrade.
>             Fixed by pinning the parameter group.
>   Tags:     aws, rds, performance
>
> Save this? [yes / edit / skip]

Two weeks later, you're back in familiar territory:

> **kno:** You have notes on this — want to load your AWS Infrastructure page?
>
> **You:** yes
>
> *Session starts with full context from prior work. No re-explaining your setup.*

You decide what matters. kno makes sure you don't lose it.

---

## Quick start

```bash
brew tap kno-ai/tap
brew install kno
kno setup
```

Restart your client (Claude Desktop or Claude Code). Start a chat and
enter `/kno` to connect. It's the only command you need — kno will
notice what matters from there.

---

## How it works

**Start every chat with `/kno`.** kno checks your vault, shows your
pages, and offers to bring in relevant context. Say yes or just start
working — kno stays attentive from there.

**kno notices when something's worth keeping** — a decision reached, a
root cause found, a design that settled — and kno offers to save it. You
confirm, and kno saves the insight with a title, summary, and tags. Ten
seconds. The cost of preserving what you learned drops from "remember
to do it later" to "say yes."

**kno recognizes familiar territory** — when your conversation touches
something already in your vault, kno offers to load that context.
Sessions start informed instead of cold.

**kno turns notes into knowledge** — periodically, kno offers to weave
your saved notes into a living page document. Each page reflects
everything you've learned about a subject, in your own words, organized
the way you think about it. This is the one intentional step — and the
one that makes your vault worth returning to.

---

## Connects to Obsidian and the tools you already use

kno works alongside your existing knowledge tools — it doesn't replace
them. Connect kno to Obsidian and your pages flow there automatically
after every update, with tags, links, and metadata already in place.
Your AI conversations become part of your knowledge base, browsable and
searchable alongside everything else you've written.

```bash
kno setup --publish ~/obsidian/kno
```

Works with any markdown tool that supports frontmatter. No extra steps
once it's set up.

---

## For developers

In Claude Code, kno detects git repositories automatically. Everything
you save gets tagged with the project name. kno tracks the knowledge
that actually matters in a codebase: decisions with dates and rationale,
known issues with open/resolved status, non-obvious setup, hard problems
solved.

Project settings travel with the repo in a `.kno` file — commit it to
share with your team. A well-curated project page is the onboarding
document you wish existed when you joined.

See the [Developer Guide](docs/guide/kno-dev-guide.md) for the full story.

---

## Your knowledge is yours

No database, no cloud service, no lock-in. Everything kno builds for
you lives in plain markdown files — readable, portable, and fully under
your control.

```
~/kno/
  config.toml
  notes/
    20260305-rds-slow-query-debugging/
      content.md
      meta.json
  pages/
    aws-infrastructure.md
    aws-infrastructure.meta.json
```

Sync it with git, Dropbox, or iCloud. Works with Claude Desktop,
Claude Code, and any AI client that supports MCP.

---

## Status

kno is in early development but ready to try. We're continuing to refine
the experience — feedback and ideas welcome via
[issues](https://github.com/kno-ai/kno/issues).

---

Your knowledge shouldn't reset every session. Give kno a try.

[User Guide](docs/guide/kno-guide.md) · [Developer Guide](docs/guide/kno-dev-guide.md) · [CLI Reference](docs/guide/kno-cli.md) · [MIT License](LICENSE)
