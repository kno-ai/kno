# kno — User Guide

kno is a knowledge vault for your AI conversations. Every conversation
resets — you re-explain your setup, rediscover prior decisions, and
rebuild context from scratch. kno changes that. It notices when something
worth preserving happens and offers to save it — no commands to
memorize, no habits to build. Your decisions, debugging breakthroughs,
and hard-won context accumulate into living documents that load into
future sessions automatically. The knowledge compounds because kno pays
attention and you decide what matters.

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
kno with detected clients (Claude Desktop, Claude Code). Restart when prompted.

**3. Enter `/kno.start` in a chat**

Start a conversation and type `/kno.start`. kno checks your vault and shows
your pages — if you have any, it offers to load relevant context. Say yes
or just start working. That's the only step you need to remember.

> **Claude Code users:** Claude Code uses a colon separator instead of a
> dot — type `/kno:start` instead of `/kno.start`. The same applies to
> all kno commands: `/kno:capture`, `/kno:curate`, `/kno:load`, etc.

From there, kno stays attentive. When something worth keeping happens — a
decision, a debugging insight, a design that settled — kno notices and
offers:

> "That's a good one — want to add it to your vault?"

You say yes, review the proposed title and tags, and confirm. The insight
that would have been trapped in a single chat is now in your vault,
available to any future session.

When you're working on a topic where your vault has relevant knowledge,
kno recognizes the overlap and offers to load it:

> "kno has notes on this — want to load your AWS Infrastructure page?"

You say yes, and the session starts informed — no re-explaining your setup,
no rediscovering decisions you already made, no cold starts on familiar
problems.

### The knowledge loop

kno runs on a simple loop: **capture → curate → load**. You save what
you learned, periodically weave notes into pages, and load those pages
into future sessions. Each pass compounds the next — better pages make
better loads, which make better sessions, which produce better saves.

This is where it pays off. You start a new session and type `/kno.start`.
If you're in a project with a project vault (see
[Project vaults](#project-vaults)), kno loads your project page
instantly. Otherwise, kno lists your pages and offers to load the
relevant one. Either way, you're working with full context in seconds —
the decisions you made yesterday, the issues you found, the setup details
you figured out. No re-explaining anything.

That context compounds over time. After a few weeks, starting a session
on a familiar topic feels like opening a document a sharp colleague has
been maintaining for you.

---

## Pages — where knowledge lives

Pages are the durable artifact in kno — living documents that reflect
everything you've learned about a subject. A page on "AWS Infrastructure"
might cover RDS parameter tuning, ECS deployment patterns, and cost
gotchas — accumulated over months of sessions, organized the way you
think about it.

Pages are created intentionally. After your first save, kno offers to
create a page — giving your knowledge a home right away. You can also
create one explicitly:

```
/kno.page new
```

kno offers a starting template with guidance that shapes how future
curations maintain the page. You can customize it or start fresh.

When you create a page, kno also offers to fold any existing sessions
into it — so your page starts with real knowledge, not empty.

### Curating notes into pages

Curating is the middle step of the knowledge loop — where your saved
notes become structured knowledge. kno lets you know when notes are
building up and suggests curating — you can also run `/kno.curate`
explicitly any time.

kno scans your uncurated notes, matches them to pages by content
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

Tags are how kno matches your notes to pages and queries. kno proposes
them during saves — you can steer with #hashtags:

```
/kno.capture — #aws #rds, the parameter group fix was the big lesson
```

kno checks existing tags from recent sessions to suggest consistent
tagging. Reuse "aws" rather than introducing "amazon" — consistent tags
make everything downstream work better.

---

## How attentive is kno?

kno pays attention by default. You can adjust how proactive it is in
`config.toml`:

```toml
[skill]
nudge_level = "active"    # "off", "light", or "active"
```

**active** (default) — Broad recognition. kno notices decisions, insights,
and solutions as they happen and offers to save them.

**light** — High-signal moments only. Conservative — stays quiet unless
something genuinely durable has landed.

**off** — No suggestions. Slash commands only. The vault and all commands
still work — you're just driving manually.

Most users won't need to change this. The default balances being helpful
without being noisy.

---

## Project vaults

When you work on a project repeatedly — a codebase, a client engagement,
a research effort — a project vault keeps that knowledge scoped and
shareable. Run `kno init` in the project directory:

```bash
cd ~/code/my-project
kno init
```

This creates a `.kno/` directory at the project root with:
- A config file (`config.toml`) with the project page bound for auto-load
- A default page named after the project, with guidance template
- Directories for notes and pages
- A `.gitignore` that excludes `notes/` and `index/` from git

When you start a session in this directory, kno automatically loads the
project page — decisions, known issues, setup details, all there before
you write a line of code.

### What's shared, what's local

If the project directory is tracked by git, pages are committed and
shared with collaborators. Notes and the search index are local to
each person — they're personal session summaries that feed into the
shared pages through curate.

### Sharing with collaborators

When `.kno/` is committed to a repo or shared directory, everyone who
opens the project gets the curated page and config. New team members
start their first session with the project's accumulated knowledge
loaded automatically.

### Project vaults vs personal vaults

Your personal vault (`~/kno`) is general-purpose — it holds knowledge
across all topics. Project vaults are scoped to a single project. You
can use both: project vaults for project-specific knowledge, your
personal vault for everything else.

When kno detects a project vault, it uses that vault exclusively for the
session. Your personal vault is unaffected.

### In git repos

When the kno MCP server detects a git repository, it enriches the
session automatically — no config needed:

- **Automatic repo tagging** — every save gets the repo name as a tag.
  You don't need to add `#my-project` — it's already there.
- **Note types** — saves can be typed as `decision`, `debt`,
  `runbook`, `bug`, or `dependency`, helping curate organize
  knowledge into the right page sections.
- **Status tracking** — `debt` and `bug` types support `open` /
  `resolved` status. When you fix a known issue, curate updates the
  page accordingly.

Git also makes project vaults natural to share — pages commit to the
repo, notes stay local via `.gitignore`. A well-curated project page
is the onboarding document you wish existed when you joined.

### Disabling project vault prompts

In git repos without a project vault, kno offers to create one. If
you'd rather not be prompted, tell kno to stop asking — it saves the
preference in your vault config so it applies across all repos.

---

## Commands reference

Most of the time, kno handles saves, loads, and curate reminders for
you. These commands are available when you want explicit control.

| Command | What it does |
|---|---|
| `/kno.start` | **Start here.** Shows pages, offers to load. Run this at the start of every chat. |
| `/kno.capture` | Save insights when kno didn't offer, or steer tags explicitly. |
| `/kno.curate` | Synthesize your notes into pages. |
| `/kno.page` | Create or manage pages. |
| `/kno.status` | Check vault health: note counts, page list, capacity. |
| `/kno.load` | Load a specific page or topic. Usually not needed — `/kno.start` handles this. |

---

## Publishing pages

Your curated pages are valuable outside of AI conversations. Publish them
to Obsidian or any markdown viewer that supports YAML frontmatter:

```bash
kno setup --publish ~/obsidian/kno
```

This adds a publish target to your user config (`~/.kno/config.toml`).
Pages from **all your vaults** — personal and project — publish there
automatically whenever a page is updated. You can also publish manually:

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

### Project grouping

When publishing from project vaults to an absolute path (like
`~/obsidian/kno`), pages are automatically grouped into a subdirectory
named after the project. This keeps pages from different projects
organized:

```
~/obsidian/kno/
  my-api/
    my-api.md
  frontend/
    frontend.md
```

Relative paths (like `docs/kno`) are already project-scoped, so no
grouping is applied. You can override this per-target with the `group`
field in config.

### Two-level publish config

Publish targets can be configured at two levels:

- **User config** (`~/.kno/config.toml`) — targets here apply across
  all your vaults. This is where personal destinations like Obsidian go.
  Use `kno setup --publish` to add targets here.

- **Vault config** (`.kno/config.toml` or `~/kno/config.toml`) — targets
  here apply to that vault only. Use this for team-specific destinations
  like a shared docs directory.

Both levels are merged at publish time. Duplicates (same path) are
deduplicated.

```toml
# In ~/.kno/config.toml — personal, applies everywhere
[[publish.targets]]
path = "~/obsidian/kno"
format = "frontmatter"

# In .kno/config.toml — project-specific
[[publish.targets]]
path = "docs/kno"
format = "markdown"
group = "false"       # no project subdirectory for relative paths
```

### Formats

Two formats are available:

- **frontmatter** (default) — YAML frontmatter, wikilinks, and clean
  content. Works with Obsidian and any markdown tool that reads frontmatter.
- **markdown** — raw markdown with guidance stripped. No frontmatter.

Override the format per-publish:

```bash
kno publish --format markdown
```

---

## Vault management

You don't need to think about capacity. When the vault is full, kno
automatically removes the oldest curated note to make room — its
knowledge is already in a page, so nothing is lost. The knowledge loop
is what protects you: once a note's insights are woven
into a page, the raw note can safely be recycled.

If the vault is full and no curated notes exist, kno removes the
oldest note regardless and warns you. Curating regularly prevents
this by folding notes into pages before they age out.

---

## Vault location

Your vault is a directory of plain files — markdown, TOML config,
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
# Restart your client
```

---

## Multiple vaults

For complete separation between work and personal knowledge:

```
kno setup --name kno-personal --vault ~/kno-personal
```

This creates a fully independent vault with its own tools and commands
(`/kno-personal.start`, `/kno-personal.capture`, etc.). The two vaults have no knowledge of
each other.

---

## Tips

**Let kno notice.** The most important saves happen when kno recognizes
a moment worth keeping and you say yes. You don't need to remember to
save — just respond when the offer is right.

**It gets better over time.** Each save feeds curate. Each curated page
makes load richer. After a few weeks, sessions on familiar topics start
with real context — no cold starts, no re-explaining your setup.

**Create pages before you need them.** If you know you're going to work
on something repeatedly, create a page for it. Even an empty page gives
curate a home for your notes.

**Start every chat with `/kno.start`.** That's the one command to remember.
Everything else — saves, loads, curate reminders — kno handles for you.
Slash commands are there if you want explicit control.

---

## Reference

For the complete CLI specification, see the [CLI Reference](kno-cli).
For detailed skill behavior, see the [Skills Reference](kno-skills).
For how the layers connect, see the
[Architecture](https://github.com/kno-ai/kno/blob/main/ARCHITECTURE.md) doc.

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

Your vault config lives at `<vault>/config.toml`. User-level config
(for publish targets that apply everywhere) lives at `~/.kno/config.toml`.

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

[skill]
nudge_level = "active"        # "off", "light", or "active"

# [[publish.targets]]
# path = "~/obsidian/kno"
# format = "frontmatter"            # "frontmatter" or "markdown"
# group = "auto"                    # "auto", "true", or "false"
```

### Managing your vault in conversation

You can ask your client to delete notes, delete pages, or rename pages
directly in conversation — no slash command needed:

```
"delete the session about RDS slow queries"
"rename the AWS Infrastructure page to AWS Cloud Ops"
```

---

Your knowledge shouldn't reset every session. Start every chat with `/kno.start`.
