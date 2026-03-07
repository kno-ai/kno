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

**Skills are conversational, not transactional.** You don't need to know
command syntax or vault structure. Describe what you want. The skill
interprets, proposes, and waits for confirmation. The structured vault
operation is the last step, not the first.

**Skills surface decisions, not surprises.** Anything that modifies the
vault — writing a note, updating a page, stamping notes as distilled
— is shown to you before it happens. The vault is never modified silently.

**Skills are proactive.** They notice when the backlog is large, when a
page hasn't been distilled in a long time, when notes cluster around
a theme that has no page. They surface these observations without being
asked. The loop stays healthy without you having to manage it manually.

**Skills narrow before loading.** On load and distill, the skill reads
summaries first to decide what's worth fetching in full. Context usage
stays efficient and predictable — the skill knows what it can afford to
load before it commits to loading it.

**Skills read before they write when state matters.** If a note already
has `distilled_into` values, the skill reads the current state before
updating it so no existing references are lost.

**All vault access goes through the CLI.** Skills never touch the vault
directly. Every read and write is a CLI call via MCP. Every change is
traceable, testable, and replaceable.

---

## The Knowledge Loop

The three core skills form a complete loop:

```
  /kno.save               /kno.distill              /kno.load
  ─────────               ────────────              ─────────
  End of session     →     Periodically        →     Start of session

  Save what you            Compress notes         Load what's relevant
  learned before           into living page         before you start,
  you close the tab.       documents.                not after.
```

Each pass through the loop compounds the next. Notes feed distill.
Distilled pages make load faster and richer. Better load means better
sessions, which produce better notes.

Over time, the vault accumulates a body of working knowledge that reflects
how you actually think — not a generic summary, but your decisions, your
lessons, your open questions.

**These three commands are the user habit.**

If you have multiple vaults (e.g. work and personal), the skill commands
are prefixed by the vault name. A vault set up as `kno-personal` exposes
`/kno-personal.save`, `/kno-personal.distill`, and `/kno-personal.load`.
The prefix is the only thing that changes — every skill works identically
within its vault.
 `/kno.page` and `/kno.status`
exist to support the loop, but in practice the skill orchestrates them for
you — surfacing page suggestions during note, checking vault health
before distill, prompting you when something needs attention. You may go
weeks without calling them directly.

---

## /kno.save

**When to use it:** At the end of a session, before closing the
conversation. The habit is: finish your work, then run `/kno.save`.

You can also save mid-session — if you've reached a natural milestone or
want to capture progress before continuing. The skill saves what's happened
so far without treating it differently. You can save multiple times in a
long session.

**Why it matters:** Most insight from an LLM session evaporates when you
close the tab. Note converts the session into a structured, searchable
record that feeds every future session on the same page. The ten seconds
it takes is the entire foundation of the knowledge loop.

**How it works**

The skill reviews the conversation, proposes a title, summary, and tags,
then asks you to confirm before writing anything.

```
/kno.save

Here's what I'll save from this session:

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
/kno.save — make sure to tag this #aws #rds, the parameter group fix
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

After note, if the vault is filling with undistilled notes, the skill
will note it:

```
Noted. You now have 18 undistilled notes.
Consider running /kno.distill to compress them into your pages.
```

If several notes share a theme but no page exists for it, the skill
may suggest creating one:

```
You've noted 4 sessions related to RDS and database performance with
no matching page. Would you like to create one before we distill?
```

---

## /kno.distill

**When to use it:** Periodically — when undistilled notes have
accumulated, or when you want a page document to be current before a
focused session. A reasonable cadence for an active user is weekly or
monthly. The skill will remind you when the backlog is significant.

**Why it matters:** Notes are raw notes. Distill is where they become
knowledge. The skill reads your undistilled notes, synthesizes what's
new, and updates your page documents following the guidance you've written
into them.

The output is a page document — a maintained, readable document that
reflects everything you've learned about a subject, organized the way you
think about it. Not a dump of session transcripts. Not a generic summary.
A document you own, can read standalone, can share, and that loads
directly into any future session to bring it instantly up to speed.

Without distill, load can only surface raw notes. With it, load
surfaces your conclusions — the refined document that represents months
of accumulated understanding in a form you can actually use.

**How it works**

```
/kno.distill

You have 22 undistilled notes. 6 pages.

Pages (by time since last distill):
  1. AWS Infrastructure     — 3 weeks ago
  2. MySQL Optimization     — 3 weeks ago
  3. React Auth Patterns   — 1 month ago
  4. Payment Processing     — 2 weeks ago
  5. EFT Processing         — 6 days ago
  6. Kubernetes Migration            — never distilled

Distill all, or start with one?
```

In both modes, the skill works the same way per page: it scans all
undistilled notes and decides which are relevant to that page.
"Distill all" cycles through every page in turn. "Start with one"
does the same scan for that page only — the rest of the notes stay
undistilled until the next run.

For each page, the skill synthesizes the update from the notes it
found relevant, shows you what changed, and asks for confirmation before
writing:

```
Distilling: AWS Infrastructure

Reading notes...
Following your guidance: "Focus on operational lessons learned the hard
way — config decisions, debugging patterns, cost surprises."

Done. Here's what changed:
  — Added: RDS parameter group pinning after minor version upgrade
  — Updated: ECS drain window recommendation (30s → 60s based on recent experience)
  — No change: connection pool section already reflected current thinking

Mark 9 notes as distilled? [yes / review first]
```

You can review the updated document before confirming. Nothing is stamped
until you say yes.

**First distill into a new page**

When a page has just been created and has no knowledge content yet, the
first distill populates it from scratch. The skill reads the guidance you
wrote, finds relevant undistilled notes, and synthesizes the initial
document. You can also trigger this immediately after creating a page —
the `/kno.page` skill offers to bootstrap-distill any relevant sessions
right away so the page isn't empty.

**Notes that fit multiple pages**

When the skill finds a note that belongs in more than one page:

```
"ACH return handling — error retry logic" is relevant to both:
  1. Payment Processing
  2. EFT Processing

Distill into both, or just one?
```

**Partial runs**

If your backlog exceeds the configured limit for a single run, the skill
processes what it can and reports the remainder:

```
Distill complete. 8 notes remain.
Run /kno.distill again to continue.
```

**Proactive suggestions**

The skill surfaces distill prompts automatically when the backlog is large,
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
    Payment Processing  — last distilled 2 weeks ago
    "...connection pool tuning, retry logic, ACH return handling..."

  Recent notes (2, undistilled):
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
first (distilled, durable knowledge), then recent undistilled notes
(fresh, not yet integrated). It reads summaries to decide what's worth
loading in full before fetching the content.

---

## /kno.page

**When to use it:** Most of the time, you won't invoke this directly —
the skill will suggest creating or reviewing a page at the right moment
during a note or distill session. But you can also call it explicitly
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

> kubernetes-migration

Good. I'll follow your guidance each time I update this page. What should
I focus on? What can I skip? How should I handle contradictions between
sessions?

> Focus on process improvements and lessons from batches that went wrong.
> Skip equipment specs — I know those. Flag anything that contradicts
> a previous decision.

Creating page: Kubernetes Migration

  Your guidance goes at the top of the document and shapes every future
  update. The knowledge content grows as notes are distilled into it.

Confirm? [yes / edit]
```

**Listing pages**

```
/kno.page list

Your pages:

  AWS Infrastructure     last distilled 3 days ago
  Payment Processing     last distilled 2 weeks ago
  Kubernetes Migration            never distilled
  React Auth Patterns   last distilled 1 month ago
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
internally at the start of every distill and load workflow, and surfaces
what matters when it matters. If the vault is nearly full, you'll hear
about it during note. If a page has gone months without a distill,
you'll hear about it when you next run distill.

Call it directly when you want the raw picture: how full is the vault,
when were pages last updated, how big is the backlog.

**Why it matters:** The skill uses this data to make every workflow
decision — context budgets, distill ordering, capacity warnings. You're
looking at the same orient call the skill makes before it acts.

**How it works**

```
/kno.status

Vault: ~/kno

Notes: 143 / 200  (57 remaining)
  Distilled:    121
  Undistilled:   22

Pages:
  AWS Infrastructure     last distilled 3 days ago
  Payment Processing     last distilled 2 weeks ago
  Kubernetes Migration            never distilled
  React Auth Patterns   last distilled 1 month ago
  EFT Processing         last distilled 6 days ago
  MySQL Optimization     last distilled 3 weeks ago

22 undistilled notes.
Run /kno.distill to compress them into your pages.
```
