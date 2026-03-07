# kno — User Guide

kno is a knowledge vault for your AI conversations. You decide what's
worth keeping, how it's organized, and when it gets refined — and that
curation is what makes it useful. Your decisions, lessons, and context
accumulate into living documents that load instantly into any future
session. The knowledge compounds because you tend it.

---

## How it works in one paragraph

At the end of a session you run `/kno.capture`. kno reviews the conversation
and proposes what to keep — you confirm, edit, or skip. Periodically you
run `/kno.curate`, which synthesizes your saved sessions into page
documents — maintained, readable files that reflect everything you've
learned about a subject, organized the way you think about it. At the
start of your next session you run `/kno.load`, which finds the relevant
pages and injects them into the conversation. Claude starts informed
instead of cold. That's the loop: capture, curate, load.

The loop takes about 30 seconds per session. That small investment is
what turns a pile of conversations into knowledge you trust — documents
you'd hand to a colleague, not transcripts you'd never reread.

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

This creates your vault at `~/kno`, writes a default config, and registers
kno as an MCP server with Claude Desktop. Restart Claude Desktop when
prompted. Five slash commands appear automatically: `/kno.capture`,
`/kno.load`, `/kno.curate`, `/kno.page`, `/kno.status`.

**3. Start a session and capture it**

Open Claude Desktop, have a conversation, then at the end:

```
/kno.capture
```

kno reviews the conversation, proposes a title, summary, and tags, and
asks you to confirm. Tags matter — they're how load and curate match
sessions to your pages and queries. Use #hashtags in your message to
steer them, or edit the proposal before confirming.

**4. Create a page**

After you've captured several sessions on the same subject, kno will suggest
creating a page. You can also do it explicitly:

```
/kno.page new
```

kno will ask what to track and how to maintain it. You describe your
preferences in plain language — what to focus on, what to skip, how to
handle contradictions. This guidance shapes every future update.

When you create a page, kno also offers to curate any existing sessions
that are relevant — so your page starts with real knowledge, not empty.

**5. Curate your sessions into pages**

```
/kno.curate
```

kno scans your uncurated sessions, matches them to pages by tags and
content, synthesizes an updated document, shows you what changed, and asks
for confirmation before writing. Sessions tagged "aws" or "rds" match
your AWS Infrastructure page; sessions tagged "payments" match Payment
Processing. The result is a maintained page document that represents
everything you've learned — in your own words, organized the way you
think about it.

**6. Load knowledge into your next session**

```
/kno.load
```

kno asks what you're working on, searches your vault, and offers to inject
the relevant page documents and recent sessions into the conversation.
Your first message lands in an already-informed conversation.

---

## The three commands to know

These are the only three you need to build a habit around.

### /kno.capture

Run this at the end of a session — or mid-session when you hit a
milestone. kno reviews the conversation and proposes what to capture.

```
/kno.capture

Here's what I'll capture from this session:

  Title:    RDS slow query debugging
  Summary:  Query planner regression after minor version upgrade. Fixed by
            pinning parameter group.
  Tags:     aws, rds, databases, performance

Save this? [yes / edit / skip]
```

You can steer it conversationally. Hashtags become tags automatically:

```
/kno.capture — tag this #aws #rds, the parameter group fix was the key thing
```

### /kno.curate

Run this periodically — weekly or monthly for active users. kno will
remind you when the backlog is significant.

```
/kno.curate

You have 22 sessions waiting to be curated. 6 pages.

Pages (by time since last curate):
  1. AWS Infrastructure     — 3 weeks ago
  2. Payment Processing     — 2 weeks ago
  3. Kubernetes Migration            — never curated
  ...

Curate all, or start with one?
```

For each page, kno scans uncurated sessions for relevance, synthesizes
the update, and shows you what changed before writing anything.

### /kno.load

Run this at the start of a session, before your first question.

```
/kno.load

What are you working on today?

> debugging a connection pool issue in our payment service

Found:
  Pages (1):    Payment Processing  — last curated 2 weeks ago
                "...connection pool tuning, retry logic, ACH return handling..."

  Sessions (2, matched by tags: payments, mysql, connection-pool):
                 ACH return handling (3 days ago)
                 MySQL connection pool (1 week ago)

Load all 3? [yes / pick / skip]
```

You can also load directly:

```
/kno.load aws infrastructure
```

---

## Pages

Pages are the durable artifact — the point of the whole system. A page
document is a maintained, readable file that reflects everything you've
learned about a subject. You can read it, share it, or load it into any
session.

Pages are created intentionally. kno will suggest creating one when
several sessions cluster around a theme with no existing page. You can
also create one explicitly with `/kno.page new`.

When you create a page, kno asks for your guidance — what to focus on,
what to skip, how to handle contradictions. That guidance lives at the top
of the document and shapes every future curate pass. The knowledge content
grows beneath it as sessions are curated in.

### Editing page guidance

Your guidance isn't locked in at creation time. As your understanding of a
subject evolves, update the guidance to change how future curations maintain
the page. Run `/kno.page` and ask to edit an existing page, or edit the
guidance section directly.

Examples of guidance you might write:

- "Focus on operational lessons and runbooks. Skip theoretical discussion."
- "Track decisions with context on why alternatives were rejected."
- "When new information contradicts existing content, keep both with dates
  so I can see how my understanding changed."
- "Keep code snippets minimal — just the commands I'll actually copy-paste."
- "Organize by service name, not by date."

Good guidance is short and practical. A few sentences is enough — you can
always refine it after seeing how curate uses it.

### What a page looks like

After several curate passes, a page becomes a living document. Here's
what an AWS Infrastructure page might look like after a few months of
sessions:

```markdown
<!-- Guidance -->
Focus on operational lessons learned the hard way — config decisions,
debugging patterns, cost surprises. Skip theoretical explanations of
AWS services. When something contradicts prior experience, keep both
with dates so I can see how my thinking evolved.

## RDS

- Pin parameter groups before minor version upgrades. Learned this after
  a query planner regression in March — the upgrade changed join order
  defaults. Always test minor upgrades in staging first.
- Connection pool: 20 connections per service instance, hard max. Going
  above this caused intermittent timeouts under load (Feb debugging session).
- Read replicas lag 50-200ms under write-heavy loads. Don't use them for
  anything requiring read-after-write consistency.

## ECS

- Drain window: 60 seconds minimum. The default 30s caused dropped
  requests during deploys when health checks were slow to propagate.
  Updated from 30s after March incident.
- Task placement: spread across AZs using the spread strategy, not
  binpack. Binpack saved ~$200/mo but created single-AZ blast radius.

## Cost patterns

- NAT gateway charges dominated our March bill — $1,400 for cross-AZ
  traffic we didn't realize was happening. Fixed by colocating services
  in the same AZ for internal traffic.
```

Notice how it reads like a document you'd hand to a colleague — not a
transcript, not a summary of sessions, but organized knowledge with
specific numbers, dates, and reasoning. Each curate pass adds,
updates, or confirms sections based on new sessions. The guidance at
the top shapes what gets included and how.

See the [Architecture](kno-knowledge-architecture) doc for more on pages,
metadata, and the mental model behind the knowledge loop.

---

## Vault management

You don't need to think about capacity. When the vault is full, kno
automatically removes the oldest curated session to make room — its
knowledge is already in a page, so nothing is lost. The curate loop
is what protects your knowledge: once a session's insights are folded
into a page, the raw session can safely be recycled.

If the vault is full and no curated sessions exist, kno removes the
oldest session regardless and warns you. This is the signal to run
`/kno.curate` — it prevents knowledge loss by folding sessions into
pages before they age out.

---

## Vault location

By default kno creates your vault at `~/kno`. But the vault is just a
directory of plain files — markdown, TOML config, and a search index. You
can put it anywhere.

### Why move your vault?

- **Obsidian / other editors.** Place your vault inside an Obsidian vault
  and your pages and sessions become browsable, searchable, and linkable
  alongside your other notes. kno writes standard markdown — no proprietary
  format.
- **Sync.** Put your vault in a synced folder (iCloud, Dropbox, Syncthing)
  and your knowledge follows you across machines. kno doesn't manage sync —
  it just reads and writes files, so any sync tool works.
- **Backup.** A vault in your existing backup path gets backed up
  automatically.
- **Organization.** Keep your vault next to related projects or notes
  instead of buried in your home directory.

### Moving an existing vault

```bash
# Move the directory
mv ~/kno ~/obsidian-vault/kno

# Re-run setup to update the MCP registration
kno setup --vault ~/obsidian-vault/kno

# Restart Claude Desktop
```

Setup detects the existing vault and preserves your data — it only updates
the MCP registration so Claude Desktop points to the new location.

### Creating a new vault in a custom location

```bash
kno setup --vault ~/obsidian-vault/kno
```

---

## Multiple vaults

If you want complete separation between work and personal knowledge — for
sync, encryption, or just peace of mind — run setup a second time with a
different name and path:

```
kno setup --name kno-personal --vault ~/kno-personal
```

This creates a fully independent vault. Its skills appear in Claude Desktop
as `/kno-personal.capture`, `/kno-personal.curate`, and `/kno-personal.load`.
The two vaults have no knowledge of each other. You choose which context
you're in by which command you type.

Each vault directory is a plain folder. Sync and encryption are handled
outside kno — point a sync tool or encryption layer at the directory and
kno doesn't need to know about it.

---

## Tags

Tags are how load and curate match sessions to your pages and queries.
kno proposes them during capture — you can steer with #hashtags:

```
/kno.capture — #aws #rds, the parameter group fix was the big lesson
```

kno shows existing tags from recent sessions so you can stay consistent.
Reuse "aws" rather than introducing "amazon" — consistent, specific tags
make everything downstream work better. When several sessions share tags
with no matching page, kno suggests creating one.

---

## Tips

**Capture immediately.** The habit that makes everything else work is running
`/kno.capture` before you close the tab. The summary is generated from the
conversation while it's still in context — you review it, confirm the
tags, and move on. Thirty seconds of curation now means Claude starts
your next session already knowing what happened.

**Create pages before you need them.** If you know you're going to work
on something repeatedly — a codebase, a health situation, a hobby — create
a page for it before the first session. Even an empty page gives curate
a home for your sessions.

**Load before you ask.** `/kno.load` at the start of a session is worth
the five seconds. Without it every session starts cold. With it Claude
already knows your setup, your prior decisions, and the approaches you've
already tried.

**Capture mid-session too.** `/kno.capture` doesn't have to wait until the end.
If you've hit a milestone in a long session — a decision made, a bug
found, a design settled — capture it now and keep going. You can capture
multiple times in one session.

**Let kno prompt you.** You don't need to remember when to curate or
whether your vault is filling up. kno surfaces these reminders during
capture and load. Trust the loop.

---

## Power user reference

For the complete CLI specification, see the [CLI Reference](kno-cli). For
how the slash commands work, see the [Skills Reference](kno-skills). For
how the layers connect (CLI → MCP → skills), see the
[Architecture](kno-knowledge-architecture) doc.

### Browsing your vault from the terminal

```bash
# list sessions
kno note list

# list sessions not yet curated
kno note list --filter curated_at=null

# search sessions
kno note search "connection pool"

# show a session in full
kno note show <id>

# list all pages
kno page list

# show a page document
kno page show <id>

# rename a page (updates files and note references)
kno page rename <id> --name "New Name"

# search pages
kno page search "aws infrastructure"
```

### Vault health

```bash
kno vault status
```

Shows session counts (total, curated, uncurated, remaining capacity),
page list with last curate dates, and current config values.

### Multiple vaults at the terminal

Use `--vault <path>` to target a specific vault on any command:

```bash
kno --vault ~/kno-personal note list
kno --vault ~/kno-work vault status
```

### Config

`~/kno/config.toml` (or the vault's own `config.toml` for additional vaults):

```toml
[notes]
max_count = 500                  # vault capacity
default_list_limit = 50
summary_max_tokens = 100         # hint to skill: target summary length
max_content_tokens = 3000        # hard limit on note content size

[pages]
max_content_tokens = 12000        # hard limit on page content size

[curate]
max_notes_per_run = 50        # max sessions processed per curate run

[search]
default_limit = 10
```

### Managing your vault in Claude

You can ask Claude to delete sessions, delete pages, or rename pages
directly in conversation — no slash command needed. These are simple
operations where the LLM handles them conversationally using the vault
tools:

```
"delete the session about RDS slow queries"
"rename the AWS Infrastructure page to AWS Cloud Ops"
"delete all pages about health"
```

Claude will confirm before making changes. Deleting a page cleans up
session references automatically — sessions that were curated into it
become eligible for curation again.

### More CLI commands

```bash
# delete a session
kno note delete <id>

# delete a page (cleans up session references)
kno page delete <id>

# remove N oldest sessions (use --dry-run first)
kno note prune --count 10 --dry-run
kno note prune --count 10

# rebuild search index if it gets out of sync
kno vault rebuild-index
```

`note prune` removes sessions oldest-first regardless of curate status.
Use it for bulk cleanup when you want to reduce vault size beyond what
auto-removal handles. `--dry-run` shows what would be removed without
deleting.
