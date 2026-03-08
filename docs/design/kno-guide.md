# kno — User Guide

kno is a knowledge vault for your AI conversations. Every conversation
you have is disconnected from the last — you re-explain your setup,
rediscover prior decisions, and rebuild context from scratch. kno changes
that. It watches for moments worth preserving and offers to capture them
— no commands to memorize, no habits to build. Your decisions, debugging
breakthroughs, and hard-won context accumulate into living documents that
load into future sessions automatically. The knowledge compounds because
kno pays attention and you decide what matters.

---

## Getting started

**1. Install kno**

```sh
# Homebrew
brew tap kno-ai/tap
brew install kno

# Or from source
go install github.com/kno-ai/kno/cmd/kno@latest
```

**2. Run setup**

```
kno setup
```

This creates your vault at `~/kno`, writes a default config, and connects
kno to Claude Desktop. Restart when prompted.

**3. Enter `/kno` in a chat**

Start a conversation and type `/kno`. kno checks your vault and shows
your pages — if you have any, it offers to load relevant context. Say yes
or just start working. That's the only step you need to remember.

From there, kno stays aware. When something worth keeping happens — a
decision, a debugging insight, a design that settled — kno notices and
offers:

> "That's a good one — want to add it to your vault?"

You say yes, review the proposed title and tags, and confirm. The insight
that would have been trapped in a single chat is now in your vault,
available to any future session.

When you're working on a topic where your vault has relevant knowledge,
kno recognizes the overlap and suggests loading it:

> "kno has notes on this — want to load your AWS Infrastructure page?"

You say yes, and the session starts informed — no re-explaining your setup,
no rediscovering decisions you already made, no cold starts on familiar
problems.

---

## Pages — where knowledge lives

Pages are the durable artifact in kno — living documents that reflect
everything you've learned about a subject. A page on "AWS Infrastructure"
might cover RDS parameter tuning, ECS deployment patterns, and cost
gotchas — accumulated over months of sessions, organized the way you
think about it.

Pages are created intentionally. After you've captured several sessions
on the same topic, kno suggests creating a page to give them a home. You
can also create one explicitly:

```
/kno.page new
```

kno asks what to track and how to maintain it. You describe your
preferences in plain language — what to focus on, what to skip, how to
handle contradictions. This guidance shapes every future update.

When you create a page, kno also offers to curate any existing sessions
that are relevant — so your page starts with real knowledge, not empty.

### Curating sessions into pages

Curate is where captured sessions become structured knowledge. kno
lets you know when notes are building up and suggests curating — you
can also run `/kno.curate` explicitly any time.

kno scans your uncurated sessions, matches them to pages by content
and tags, synthesizes an updated document, shows you what changed, and
asks for confirmation before writing. A reasonable cadence is weekly or
monthly, but you don't need to track when it's time — kno will tell you.

Each curate pass also updates the page's summary — a short description of
what the page currently covers. This summary is what lets kno recognize
when future conversations overlap with your existing knowledge.

### What a page looks like

After several curate passes, a page becomes a living document:

```markdown
<!-- Guidance -->
Focus on operational lessons learned the hard way — config decisions,
debugging patterns, cost surprises. Skip theoretical explanations of
AWS services.

## RDS

- Pin parameter groups before minor version upgrades. Learned this after
  a query planner regression in March — the upgrade changed join order
  defaults. Always test minor upgrades in staging first.
- Connection pool: 20 connections per service instance, hard max.

## ECS

- Drain window: 60 seconds minimum. The default 30s caused dropped
  requests during deploys.
- Task placement: spread across AZs, not binpack.

## Cost patterns

- NAT gateway charges dominated our March bill — $1,400 for cross-AZ
  traffic. Fixed by colocating services in the same AZ.
```

This reads like a document you'd hand to a colleague — not a transcript,
not a summary of sessions, but organized knowledge with specific numbers,
dates, and reasoning.

### Editing page guidance

Your guidance isn't locked in at creation time. As your understanding
evolves, update it to change how future curations maintain the page:

- "Focus on operational lessons and runbooks. Skip theoretical discussion."
- "Track decisions with context on why alternatives were rejected."
- "When new information contradicts existing content, keep both with dates."
- "Organize by service name, not by date."

### Tags

Tags are how kno matches sessions to your pages and queries. kno proposes
them during capture — you can steer with #hashtags:

```
/kno.capture — #aws #rds, the parameter group fix was the big lesson
```

kno checks existing tags from recent sessions to suggest consistent
tagging. Reuse "aws" rather than introducing "amazon" — consistent tags
make everything downstream work better.

---

## How proactive is kno?

kno pays attention by default. You can adjust how proactive it is in
`config.toml`:

```toml
[nudges]
level = "light"    # "off", "light", or "active"
```

**light** (default) — Nudges only for high-signal knowledge checkpoints.
Conservative, stays quiet unless something genuinely durable has landed.

**active** — Broader checkpoint recognition. Good for users building a
vault quickly or who want more capture opportunities surfaced.

**off** — No suggestions. Slash commands only. The vault and all
commands still work — you're just driving manually.

Most users won't need to change this. The default balances being helpful
without being noisy.

---

## Commands reference

Most of the time, kno handles captures, loads, and curate reminders
for you. These commands are available when you want explicit control.

| Command | What it does |
|---|---|
| `/kno` | **Start here.** Shows pages, offers to load. Run this at the start of every chat. |
| `/kno.capture` | Capture insights when kno didn't offer, or steer tags explicitly. |
| `/kno.curate` | Synthesize uncurated sessions into pages. |
| `/kno.page` | Create or manage pages. |
| `/kno.status` | Check vault health: session counts, page list, capacity. |
| `/kno.load` | Load a specific page or topic. Usually not needed — `/kno` handles this. |

---

## Publishing pages

Your curated pages are valuable outside of AI conversations. Publish them
to Obsidian or any markdown viewer that supports YAML frontmatter:

```bash
kno setup --publish ~/obsidian/kno
```

This adds a publish target to your config. From then on, pages are
**automatically published after every curate** — no extra step needed.
You can also publish manually at any time:

```bash
kno publish
```

### What gets published

Published pages include:

- **YAML frontmatter** — title, aliases, tags, summary, created and
  updated dates
- **Wikilinks** — references to other page names become `[[wikilinks]]`
  for easy cross-navigation in Obsidian
- **Clean content** — guidance comments at the top of vault pages are
  stripped from the published output

The vault remains the source of truth. Published files are derived
artifacts that can be regenerated at any time with `kno publish`.

### Formats

Two formats are available:

- **frontmatter** (default) — YAML frontmatter, wikilinks, and clean
  content. Works with Obsidian and any markdown tool that reads frontmatter.
- **markdown** — raw markdown with guidance stripped. No frontmatter.

Override the format per-publish:

```bash
kno publish --format markdown
```

### Multiple targets

Add multiple targets in `config.toml`:

```toml
[[publish.targets]]
path = "~/obsidian/kno"
format = "frontmatter"

[[publish.targets]]
path = "~/docs/kno"
format = "markdown"
```

### Time to value

Publishing is the fastest way to see the value of your vault. After one
curate pass, you have a living document in Obsidian — browsable, searchable,
and linked to your other notes. Just one good page makes the loop click.

---

## Vault management

You don't need to think about capacity. When the vault is full, kno
automatically removes the oldest curated session to make room — its
knowledge is already in a page, so nothing is lost. The curate loop
is what protects your knowledge: once a session's insights are folded
into a page, the raw session can safely be recycled.

If the vault is full and no curated sessions exist, kno removes the
oldest session regardless and warns you. Curating regularly prevents
this by folding sessions into pages before they age out.

---

## Vault location

Your vault is just a directory of plain files — markdown, TOML config,
and a search index. You can put it anywhere.

- **Obsidian / other editors.** Publish curated pages to an Obsidian vault
  with `kno setup --publish ~/obsidian/kno`. Pages get frontmatter, tags,
  and wikilinks — browsable alongside your other notes.
- **Sync.** Put your vault in a synced folder (iCloud, Dropbox, Syncthing)
  and your knowledge follows you across machines.
- **Backup.** A vault in your existing backup path gets backed up
  automatically.

### Moving an existing vault

```bash
mv ~/kno ~/obsidian-vault/kno
kno setup --vault ~/obsidian-vault/kno
# Restart Claude Desktop
```

---

## Multiple vaults

For complete separation between work and personal knowledge:

```
kno setup --name kno-personal --vault ~/kno-personal
```

This creates a fully independent vault with its own tools and commands
(`/kno-personal.capture`, etc.). The two vaults have no knowledge of
each other.

---

## Tips

**Let kno notice.** The most important captures happen when kno recognizes
a checkpoint and you say yes. You don't need to remember to capture — just
respond when the offer is right.

**It gets better over time.** Each capture feeds curate. Each curated page
makes load richer. After a few weeks, sessions on familiar topics start
with real context — no cold starts, no re-explaining your setup.

**Create pages before you need them.** If you know you're going to work
on something repeatedly, create a page for it. Even an empty page gives
curate a home for your sessions.

**Start every chat with `/kno`.** That's the one command to remember.
Everything else — captures, loads, curate reminders — kno handles
for you. Slash commands are there if you want explicit control.

---

## Reference

For the complete CLI specification, see the [CLI Reference](kno-cli). For
detailed skill behavior, see the [Skills Reference](kno-skills). For
how the layers connect, see the
[Architecture](kno-knowledge-architecture) doc.

### Browsing your vault from the terminal

Your vault is plain files. You can browse it directly:

```bash
kno note list
kno note list --filter curated_at=null
kno note search "connection pool"
kno note show <id>
kno page list
kno page show <id>
kno page search "aws infrastructure"
```

### Vault health

```bash
kno vault status
```

### Config

`~/kno/config.toml`:

```toml
[notes]
max_count = 500
default_list_limit = 50
summary_max_tokens = 100
max_content_tokens = 3000

[pages]
max_content_tokens = 12000

[curate]
max_notes_per_run = 50

[search]
default_limit = 10

[nudges]
level = "light"              # "off", "light", or "active"

# [[publish.targets]]
# path = "~/obsidian/kno"
# format = "frontmatter"            # "frontmatter" or "markdown"
```

### Managing your vault in conversation

You can ask your client to delete sessions, delete pages, or rename pages
directly in conversation — no slash command needed:

```
"delete the session about RDS slow queries"
"rename the AWS Infrastructure page to AWS Cloud Ops"
```
