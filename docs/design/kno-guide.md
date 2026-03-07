# kno — User Guide

kno is a knowledge vault for your AI conversations. It gives Claude a memory
that compounds over time — your decisions, lessons, and context accumulate
into living documents that load instantly into any future session.

---

## How it works in one paragraph

At the end of a session you run `/kno.save`. kno saves a structured
summary to your vault. Periodically you run `/kno.distill`, which reads
your saved sessions and synthesizes them into page documents — maintained,
readable files that reflect everything you've learned about a subject. At
the start of your next session you run `/kno.load`, which finds the
relevant page documents and injects them into the conversation. Claude
starts informed instead of cold. That's the loop.

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
prompted. Five slash commands appear automatically: `/kno.save`,
`/kno.load`, `/kno.distill`, `/kno.page`, `/kno.status`.

**3. Start a session and save it**

Open Claude Desktop, have a conversation, then at the end:

```
/kno.save
```

kno reviews the conversation, proposes a title, summary, and tags, and
asks you to confirm. That's it — the session is saved.

**4. Create a page**

After you've saved several sessions on the same subject, kno will suggest
creating a page. You can also do it explicitly:

```
/kno.page new
```

kno will ask what to track and how to maintain it. You describe your
preferences in plain language — what to focus on, what to skip, how to
handle contradictions. This guidance shapes every future update.

When you create a page, kno also offers to distill any existing sessions
that are relevant — so your page starts with real knowledge, not empty.

**5. Distill your sessions into pages**

```
/kno.distill
```

kno scans your undistilled sessions, finds what's relevant to each page,
synthesizes an updated document, shows you what changed, and asks for
confirmation before writing. The result is a maintained page document
that represents everything you've learned — in your own words, organized
the way you think about it.

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

### /kno.save

Run this at the end of a session. kno reviews the conversation and
proposes what to save.

```
/kno.save

Here's what I'll save from this session:

  Title:    RDS slow query debugging
  Summary:  Query planner regression after minor version upgrade. Fixed by
            pinning parameter group.
  Tags:     aws, rds, databases, performance

Save this? [yes / edit / skip]
```

You can steer it conversationally. Hashtags become tags automatically:

```
/kno.save — tag this #aws #rds, the parameter group fix was the key thing
```

### /kno.distill

Run this periodically — weekly or monthly for active users. kno will
remind you when the backlog is significant.

```
/kno.distill

You have 22 sessions waiting to be distilled. 6 pages.

Pages (by time since last distill):
  1. AWS Infrastructure     — 3 weeks ago
  2. Payment Processing     — 2 weeks ago
  3. Kubernetes Migration            — never distilled
  ...

Distill all, or start with one?
```

For each page, kno scans undistilled sessions for relevance, synthesizes
the update, and shows you what changed before writing anything.

### /kno.load

Run this at the start of a session, before your first question.

```
/kno.load

What are you working on today?

> debugging a connection pool issue in our payment service

Found:
  Pages (1):    Payment Processing  — last distilled 2 weeks ago
  Sessions (2):  ACH return handling (3 days ago)
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
of the document and shapes every future distill pass. The knowledge content
grows beneath it as sessions are distilled in.

See the [Architecture](kno-knowledge-architecture) doc for more on pages,
metadata, and the mental model behind the knowledge loop.

---

## Vault management

Your vault automatically manages capacity. When the vault is full and you
save a new session, kno removes the oldest distilled session to make room.
If no distilled sessions exist, it removes the oldest session regardless
and warns you that undistilled knowledge may have been lost. The vault
never blocks — you can always save.

This means you don't need to think about capacity day-to-day. The distill
loop protects your knowledge: once a session's insights are folded into a
page, the raw session can safely be recycled.

---

## Multiple vaults

If you want complete separation between work and personal knowledge — for
sync, encryption, or just peace of mind — run setup a second time with a
different name and path:

```
kno setup --name kno-personal --vault ~/kno-personal
```

This creates a fully independent vault. Its skills appear in Claude Desktop
as `/kno-personal.save`, `/kno-personal.distill`, and `/kno-personal.load`.
The two vaults have no knowledge of each other. You choose which context
you're in by which command you type.

Each vault directory is a plain folder. Sync and encryption are handled
outside kno — point a sync tool or encryption layer at the directory and
kno doesn't need to know about it.

---

## Tips

**Save immediately.** The habit that makes everything else work is running
`/kno.save` before you close the tab. The summary is generated from the
conversation while it's still in context. Waiting until later means
reconstructing it.

**Create pages before you need them.** If you know you're going to work
on something repeatedly — a codebase, a health situation, a hobby — create
a page for it before the first session. Even an empty page gives distill
a home for your sessions.

**Load before you ask.** `/kno.load` at the start of a session is worth
the five seconds. Without it every session starts cold. With it Claude
already knows your setup, your prior decisions, and the approaches you've
already tried.

**Let kno prompt you.** You don't need to remember when to distill or
whether your vault is filling up. kno surfaces these reminders during
save and load. Trust the loop.

---

## Power user reference

For the complete CLI specification, see the [CLI Reference](kno-cli). For
details on how the slash commands work internally, see the
[Skills Reference](kno-skills).

### Browsing your vault from the terminal

```bash
# list sessions
kno note list

# list sessions not yet distilled
kno note list --filter distilled_at=null

# search sessions
kno note search "connection pool"

# show a session in full
kno note show <id>

# list all pages
kno page list

# show a page document
kno page show <id>

# search pages
kno page search "aws infrastructure"
```

### Vault health

```bash
kno vault status
```

Shows session counts (total, distilled, undistilled, remaining capacity),
page list with last distill dates, and current config values.

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
max_count = 200                  # vault capacity
default_list_limit = 50
summary_max_tokens = 100         # hint to skill: target summary length

[pages]
max_content_tokens = 8000        # soft cap; content truncated with warning

[distill]
max_notes_per_run = 50        # max sessions processed per distill run

[search]
default_limit = 5
```

### Admin commands

These are CLI-only — not available via Claude Desktop.

```bash
# remove N oldest sessions (use --dry-run first)
kno admin prune --count 10 --dry-run
kno admin prune --count 10

# delete a page and clean up session references
kno admin page delete <id>

# rebuild search index if it gets out of sync
kno admin index rebuild
```

`admin prune` removes sessions oldest-first regardless of distill status.
Use it for bulk cleanup when you want to reduce vault size beyond what
auto-removal handles. `--dry-run` shows what would be removed without
deleting.

`admin page delete` removes the page document and clears its id from
any session's `distilled_into` field, making those sessions eligible to
be distilled again into a new page.
