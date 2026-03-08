# kno

**A knowledge vault for your AI conversations.** &nbsp; [View on GitHub](https://github.com/kno-ai/kno)

Your AI conversations reset every session. The insight you reached, the
decision you worked through, the solution that finally clicked — none of
it carries forward. kno pays attention so you don't have to. It notices
when something worth preserving happens and offers to save it. When
you're revisiting familiar territory, kno offers to bring back what you
already know. Your knowledge compounds across every conversation.

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

Two weeks later, you type `/kno.start` at the start of a new session:

> **kno:** Your vault has **CNC Machine Maintenance** and **AWS
> Infrastructure**. Want me to load any of these?
>
> **You:** load the CNC page
>
> *Session starts with full context from prior work.*

You decide what matters. kno makes sure you don't lose it.

---

## How it works

**Start every chat with `/kno.start`.** kno checks your vault, shows your
pages, and offers to bring in relevant context. Say yes or just start
working — kno stays attentive from there.

**kno notices when something's worth keeping** — a decision reached, a
root cause found, a design that settled — and offers to save it. You
confirm, and it's saved with a title, summary, and tags. Ten seconds.
The cost of preserving an insight drops from "remember to save,
context-switch to a notes tool, decide what to write" to "say yes."

**kno recognizes familiar territory** — when your conversation touches
something already in your vault, kno offers to load that context.
Sessions start informed instead of cold. No re-explaining your setup,
no rediscovering decisions you already made.

**kno turns notes into knowledge** — when your saved notes build up,
kno offers to weave them into a living page document. Each page reflects
everything you've learned about a subject, organized the way you think
about it. This is the one intentional step — and the one that makes
your vault worth returning to.

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

Restart your client (Claude Desktop or Claude Code). Enter `/kno.start` in a
chat to connect (in Claude Code, use `/kno:start`) — kno shows your pages
and offers to bring in relevant context. From there, it stays attentive:
noticing when something's worth keeping and offering to load what you
already know.

See the [User Guide](guide/kno-guide) for the full walkthrough.

---

## Connects to Obsidian and the tools you already use

kno works alongside your existing knowledge tools — it doesn't replace
them. Connect kno to Obsidian and your pages flow there automatically
after every update, with tags, links, and metadata already in place.
Your AI conversations become part of your knowledge base, browsable and
searchable alongside everything else you've written.

```sh
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
solved. Project settings travel with the repo in a `.kno` file — commit
it to share with your team, or keep it personal.

See the [Developer Guide](guide/kno-dev-guide) for the full story.

---

## Your knowledge is yours

No database, no cloud service, no lock-in. Everything kno builds for
you lives in plain markdown files — readable, portable, and fully under
your control. Sync it with git, Dropbox, or iCloud. Works with Claude
Desktop, Claude Code, and any AI client that supports MCP.

---

## Status

kno is in early development but ready to try. We're continuing to refine
the experience — feedback and ideas welcome via
[issues](https://github.com/kno-ai/kno/issues).

---

## Documentation

- [User Guide](guide/kno-guide) — getting started and vault management
- [Developer Guide](guide/kno-dev-guide) — git detection, project pages, team use
- [Architecture](https://github.com/kno-ai/kno/blob/main/ARCHITECTURE.md) — design principles, layers, and knowledge model
- [Skills Reference](guide/kno-skills) — skill behavior and slash command details
- [CLI Reference](guide/kno-cli) — complete command specification

<p style="margin-top: 3rem; color: #666; font-size: 0.85rem;">
  MIT License | <a href="https://github.com/kno-ai/kno">GitHub</a>
</p>
