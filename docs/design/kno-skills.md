# kno — Skills Reference

Skills are slash commands available in Claude Desktop. They are the
conversational surface of kno — they interpret what you want, ask when
something is unclear, and translate your intent into precise vault operations.

Every skill is iterative. You don't need to know the right syntax or the
exact command. You describe what you want, the skill proposes what it will
do, and you confirm before anything is written.

---

## Skill Design Principles

These principles govern how all skills behave. Understanding them helps you
trust the system — the skills follow these rules consistently.

**You curate, kno compounds.** Every capture, curate, and load is a moment
where you decide what matters. That curation is the mechanism — the reason
your pages read like documents you'd hand to a colleague, not auto-generated
summaries you'd never reread. The loop takes seconds per session, and the
knowledge it produces is yours: shaped by your judgment, organized by your
priorities, refined every time you run curate.

**Skills are conversational, not transactional.** You don't need to know
command syntax or vault structure. Describe what you want. The skill
interprets, proposes, and waits for confirmation. The structured vault
operation is the last step, not the first.

**Skills surface decisions, not surprises.** Anything that modifies the
vault — writing a note, updating a page, stamping notes as curated
— is shown to you before it happens. The vault is never modified silently.

**Skills are proactive.** They notice when the backlog is large, when a
page hasn't been curated in a long time, when notes cluster around
a theme that has no page. They surface these observations without being
asked. The loop stays healthy without you having to manage it manually.

**Skills narrow before loading.** On load and curate, the skill reads
summaries first to decide what's worth fetching in full. Context usage
stays efficient and predictable — the skill knows what it can afford to
load before it commits to loading it.

**Skills read before they write when state matters.** If a note already
has `curated_into` values, the skill reads the current state before
updating it so no existing references are lost.

**All vault access goes through the CLI.** Skills never touch the vault
directly. Every read and write is a CLI call via MCP. Every change is
traceable, testable, and replaceable.

---

## The Knowledge Loop

The three core skills form a complete loop:

```
  /kno.capture               /kno.curate              /kno.load
  ─────────               ────────────              ─────────
  End of session     →     Periodically        →     Start of session

  Capture what you         Compress notes         Load what's relevant
  learned before           into living page         before you start,
  you close the tab.       documents.                not after.
```

Each pass through the loop compounds the next. Captures feed curate.
Curated pages make load faster and richer. Better load means better
sessions, which produce better notes.

The loop is proactive by design. You choose what to capture, what to focus
on during curate, and what to load. That human judgment at each step is
what produces pages worth reading — documents that reflect how you
actually think, not auto-generated summaries. Your decisions, your
lessons, your open questions, organized by your priorities.

**These three commands are the user habit.**

If you have multiple vaults (e.g. work and personal), the skill commands
are prefixed by the vault name. A vault set up as `kno-personal` exposes
`/kno-personal.capture`, `/kno-personal.curate`, and `/kno-personal.load`.
The prefix is the only thing that changes — every skill works identically
within its vault.
`/kno.page` and `/kno.status`
exist to support the loop, but in practice the skill orchestrates them for
you — surfacing page suggestions during capture, checking vault health
before curate, prompting you when something needs attention. You may go
weeks without calling them directly.

---

## /kno.capture

**When to use it:** At the end of a session, before closing the
conversation. The habit is: finish your work, then run `/kno.capture`.

You can also capture mid-session — if you've reached a natural milestone or
want to capture progress before continuing. The skill captures what's happened
so far without treating it differently. You can capture multiple times in a
long session.

**Why it matters:** Most insight from an LLM session evaporates when you
close the tab. Capture converts the session into a structured, searchable
record that feeds every future session on the same topic. You review the
title, summary, and tags before confirming — that moment of curation is
what makes the knowledge findable and useful later. Ten seconds now
is the entire foundation of the knowledge loop.

**How it works**

The skill reviews the conversation, proposes a title, summary, and tags,
then asks you to confirm before writing anything. Tags are the primary
signal that load and curate use to match sessions to pages and queries
— the skill checks existing tags from recent sessions to suggest
consistent tagging.

```
/kno.capture

Here's what I'll capture from this session:

  Title:    RDS slow query debugging
  Summary:  Query planner regression after minor version upgrade. Fixed by
            pinning parameter group. Key lesson: always test minor upgrades
            in staging.

  Tags:     aws, rds, databases, performance

Save this? [yes / edit / skip]
```

You confirm, edit, or skip. The skill only writes on confirmation.

**Conversational shortcuts**

You can steer the note before or after the proposal. Hashtags in your
message become tags automatically:

```
/kno.capture — make sure to tag this #aws #rds, the parameter group fix
was the key thing

Here's what I'll note:

  Title:    RDS slow query debugging
  Tags:     aws, rds, databases, performance  ← picked up from your message
  ...
```

You can also correct the proposal in plain language:

```
> the title should be "RDS minor version upgrade regression" and drop the
> databases tag, this was really an RDS-specific issue

Got it. Updated:
  Title:    RDS minor version upgrade regression
  Tags:     aws, rds, performance
```

**Proactive suggestions**

After saving, if the vault is filling with uncurated sessions, the skill
will mention it:

```
Noted. You now have 18 uncurated notes.
Consider running /kno.curate to compress them into your pages.
```

If several notes share tags but no page exists for them, the skill
suggests creating one:

```
You've saved 4 sessions tagged "rds" and "database-performance" with
no matching page. Want to create one?
```

---

## /kno.curate

**When to use it:** Periodically — when uncurated notes have
accumulated, or when you want a page document to be current before a
focused session. A reasonable cadence for an active user is weekly or
monthly. The skill will remind you when the backlog is significant.

**Why it matters:** Notes are raw notes. Curate is where they become
knowledge. The skill reads your uncurated notes, synthesizes what's
new, and updates your page documents following the guidance you've written
into them.

The output is a page document — a maintained, readable document that
reflects everything you've learned about a subject, organized the way you
think about it. Not a dump of session transcripts. Not a generic summary.
A document you own, can read standalone, can share, and that loads
directly into any future session to bring it instantly up to speed.

Without curate, load can only surface raw notes. With it, load
surfaces your conclusions — the refined document that represents months
of accumulated understanding in a form you can actually use.

**How it works**

```
/kno.curate

You have 22 uncurated notes. 6 pages.

Pages (by time since last curate):
  1. AWS Infrastructure     — 3 weeks ago
  2. MySQL Optimization     — 3 weeks ago
  3. React Auth Patterns   — 1 month ago
  4. Payment Processing     — 2 weeks ago
  5. EFT Processing         — 6 days ago
  6. Kubernetes Migration            — never curated

Curate all, or start with one?
```

In both modes, the skill works the same way per page: it scans all
uncurated notes and matches them to pages using tags and content.
Sessions tagged "aws" or "rds" match an AWS Infrastructure page;
sessions tagged "payments" match Payment Processing. Tag overlap is
the primary relevance signal. "Curate all" cycles through every page
in turn. "Start with one" does the same scan for that page only — the
rest of the notes stay uncurated until the next run.

For each page, the skill synthesizes the update from the notes it
found relevant, shows you what changed, and asks for confirmation before
writing:

```
Curating: AWS Infrastructure

Reading notes...
Following your guidance: "Focus on operational lessons learned the hard
way — config decisions, debugging patterns, cost surprises."

Done. Here's what changed:
  — Added: RDS parameter group pinning after minor version upgrade
  — Updated: ECS drain window recommendation (30s → 60s based on recent experience)
  — No change: connection pool section already reflected current thinking

Mark 9 notes as curated? [yes / review first]
```

You can review the updated document before confirming. Nothing is stamped
until you say yes.

**First curate into a new page**

When a page has just been created and has no knowledge content yet, the
first curate populates it from scratch. The skill reads the guidance you
wrote, finds relevant uncurated notes, and synthesizes the initial
document. You can also trigger this immediately after creating a page —
the `/kno.page` skill offers to bootstrap-curate any relevant sessions
right away so the page isn't empty.

**Notes that fit multiple pages**

When the skill finds a note that belongs in more than one page:

```
"ACH return handling — error retry logic" is relevant to both:
  1. Payment Processing
  2. EFT Processing

Curate into both, or just one?
```

**Partial runs**

If your backlog exceeds the configured limit for a single run, the skill
processes what it can and reports the remainder:

```
Curate complete. 8 notes remain.
Run /kno.curate again to continue.
```

**Proactive suggestions**

The skill surfaces curate prompts automatically when the backlog is large,
the vault is nearing capacity, or it has been a long time since a page
was updated. You don't need to remember to check.

---

## /kno.load

**When to use it:** At the start of a session, before you ask your first
question. Especially valuable when returning to ongoing work or a domain
where you have established context.

You can also load mid-session — if the conversation shifts to a topic
where you have existing knowledge, `/kno.load` will search and inject
that context without disrupting your flow.

**Why it matters:** Without load, every session starts cold — Claude has
no memory of your prior decisions, the approaches you've already tried, or
the lessons you've already learned. With load, the session starts informed.
You don't re-explain your setup. You don't rediscover the same dead ends.
You pick up where your knowledge left off.

**How it works**

```
/kno.load

What are you working on today?

> debugging a connection pool exhaustion issue in our payment service

Found:

  Pages (1):
    Payment Processing  — last curated 2 weeks ago
    "...connection pool tuning, retry logic, ACH return handling..."

  Recent sessions (2, matched by tags: payments, mysql, connection-pool):
    ACH return handling — error retry logic         3 days ago
    MySQL connection pool — max_connections         1 week ago

Load all 3? [yes / pick / skip]
```

On confirmation, the skill reads the full content of the selected pages
and notes and injects them into the session. Your first real message
lands in an already-informed conversation.

**Direct load**

If you already know what you need, skip the question:

```
/kno.load aws infrastructure
```

The skill searches immediately and presents what it finds.

**What gets loaded**

The skill balances relevance against context budget. It searches pages
first (curated, durable knowledge), then recent uncurated notes
(fresh, not yet integrated). It uses tag overlap as a primary relevance
signal — sessions tagged with terms matching your query or a page's
theme rank higher than those with only generic text matches. It reads
summaries and tags to decide what's worth loading in full before
fetching the content.

---

## /kno.page

**When to use it:** Most of the time, you won't invoke this directly —
the skill will suggest creating or reviewing a page at the right moment
during a note or curate session. But you can also call it explicitly
when you want to set up a new knowledge area, review your page list, or
update the guidance that shapes how a page is maintained.

**Why it matters:** Pages are the table of contents for your vault.
A note without a relevant page will never become structured knowledge —
it will sit in the backlog and eventually age out. The skill watches for
this pattern and surfaces it: if several notes share a theme and no
page exists for it, it will suggest creating one before you lose the thread.

When you do create a page, it's an intentional act. The skill guides you
through it conversationally, asking what to focus on and what to skip, so
the document reflects how *you* think about the subject — not a generic summary.

**Creating a page**

```
/kno.page new

What area of knowledge do you want to track?

> our kubernetes migration

Good. I'll follow your guidance each time I update this page. What should
I focus on? What can I skip? How should I handle contradictions between
sessions?

> Focus on migration blockers, service-by-service lessons, and networking
> gotchas. Skip basic k8s concepts — I know those. Flag anything that
> contradicts a previous decision.

Creating page: Kubernetes Migration

  Your guidance goes at the top of the document and shapes every future
  update. The knowledge content grows as notes are curated into it.

Confirm? [yes / edit]
```

**Listing pages**

```
/kno.page list

Your pages:

  AWS Infrastructure     last curated 3 days ago
  Payment Processing     last curated 2 weeks ago
  Kubernetes Migration            never curated
  React Auth Patterns   last curated 1 month ago
```

**Editing a page**

You can update the guidance, the knowledge content, or both:

```
/kno.page edit kubernetes-migration

Current content: [shows full document]

What would you like to change? You can revise the guidance at the top,
edit the knowledge document directly, or both.
```

---

## /kno.status

**When to use it:** Rarely directly — the skill checks vault status
internally at the start of every curate and load workflow, and surfaces
what matters when it matters. If the vault is nearly full, you'll hear
about it during note. If a page has gone months without a curate,
you'll hear about it when you next run curate.

Call it directly when you want the raw picture: how full is the vault,
when were pages last updated, how big is the backlog.

**Why it matters:** The skill uses this data to make every workflow
decision — context budgets, curate ordering, capacity warnings. You're
looking at the same orient call the skill makes before it acts.

**How it works**

```
/kno.status

Vault: ~/kno

Notes: 143 / 500  (357 remaining)
  Curated:      121
  Uncurated:      22

Pages:
  AWS Infrastructure     last curated 3 days ago
  Payment Processing     last curated 2 weeks ago
  Kubernetes Migration            never curated
  React Auth Patterns   last curated 1 month ago
  EFT Processing         last curated 6 days ago
  MySQL Optimization     last curated 3 weeks ago

22 uncurated notes.
Run /kno.curate to compress them into your pages.
```
