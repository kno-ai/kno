# kno — Skills Reference

kno operates through two complementary interfaces: **active awareness** and
**slash commands**. Awareness is the default — kno watches for knowledge
checkpoints and offers to act at the right moments without being asked.
Slash commands provide explicit control for users who want to drive the
loop manually or override what awareness doesn't catch.

Both interfaces execute the same procedures. A capture initiated by an
awareness nudge is identical to one initiated by `/kno.capture` — same
structure, same metadata, same confirmation step. The trigger changes,
the execution path does not.

---

## Design Principles

**You curate, kno compounds.** Every capture and curate is a moment
where you decide what matters. That curation is the mechanism — the reason
your pages read like documents you'd hand to a colleague, not auto-generated
summaries you'd never reread. The knowledge compounds because your judgment
shapes every step.

**Awareness first, commands second.** kno pays attention to the
conversation and surfaces the right action at the right moment. Slash
commands are always available for explicit control, but the primary
experience is a conversation that manages its own knowledge.

**Skills surface decisions, not surprises.** Anything that modifies the
vault — writing a note, updating a page, stamping notes as curated
— is shown to you before it happens. The vault is never modified silently.

**Skills narrow before loading.** On load and curate, the skill reads
summaries first to decide what's worth fetching in full. Context usage
stays efficient and predictable.

**Skills read before they write when state matters.** If a note already
has `curated_into` values, the skill reads the current state before
updating it so no existing references are lost.

**All vault access goes through the CLI.** Skills never touch the vault
directly. Every read and write is a CLI call via MCP. Every change is
traceable, testable, and replaceable.

---

## Active Awareness

Active awareness is kno's standing presence in every conversation. It
requires no activation — it's active from the moment a session begins
and remains present for its duration.

The LLM already understands the full context of the conversation — the
tradeoffs weighed, the root causes found, the reasoning behind decisions.
Awareness leverages that understanding. The LLM is the natural entity to
recognize when something durable has crystallized and when prior knowledge
would help.

### Knowledge checkpoints

The core awareness behavior: recognizing moments where durable knowledge
has crystallized and offering to capture it.

A checkpoint is:

- A decision reached after weighing tradeoffs
- A non-obvious root cause identified
- A design that settled after iteration
- A working solution after failed attempts
- Something explicitly flagged as important

When kno recognizes a checkpoint:

> "That's a good one — want to add it to your vault?"

The user confirms or declines. On confirmation, kno runs the full capture
procedure: proposes title, summary, tags, and content, confirms with the
user, then writes to the vault.

### Topic awareness

Early in a conversation, if the topic overlaps with existing vault
knowledge, kno mentions it once:

> "kno has notes on this — want to load your AWS Infrastructure page?"

On confirmation, kno searches and loads relevant pages and notes. If
declined, it drops the subject.

### Session wrap-up

If the conversation produced valuable knowledge and capture hasn't
happened, kno offers when the user signals they're wrapping up:

> "Want me to add this to your vault before we wrap up?"

### Nudge discipline

- High threshold. Over-nudging is worse than under-nudging.
- At most one nudge between user-initiated captures.
- Never re-nudge after a decline.
- Frame as offering value, not reminding: "That's a good one for your
  vault" not "Don't forget to save."
- One sentence. No explanations unless asked.

### Nudge levels

Configured in `config.toml`:

```toml
[nudges]
level = "light"
```

- **off** — No awareness nudges. Slash commands only.
- **light** — (default) High-signal checkpoints only. Conservative.
- **active** — Broader checkpoint recognition. More nudges.

---

## The Knowledge Loop

The three core operations form a complete loop:

```
  capture                    curate                   load
  ───────                    ──────                   ────
  Knowledge checkpoint  →    Periodically        →    New session on
  (awareness offers)         /kno.curate              familiar topic
                             (user-initiated)         (awareness offers)

  Capture what you           Compress notes           Load what's relevant
  learned.                   into living page         before you start.
                             documents.
```

Each pass through the loop compounds the next. Captures feed curate.
Curated pages make load faster and richer. Better loads mean better
sessions, which produce better captures. The more you use it, the more
valuable the vault becomes.

Awareness initiates capture and load automatically — it notices the right
moment and offers. The user's only cost is saying yes. Curate is where
the user decides what matters: a periodic synthesis pass where they shape
their pages with full attention. It's the intentional step that makes
pages worth reading — and the right division of labor. Low-judgment
moments (noticing a checkpoint, recognizing a familiar topic) are
automated. The high-judgment moment (shaping knowledge into a lasting
document) stays with the user.

---

## /kno

**The entry point.** Start every chat with `/kno` to connect to your vault.

The skill calls `kno_vault_status`, lists your pages by name, and offers
to load relevant context. The user says yes or just starts working —
awareness takes over from there.

This is the only step users need to remember. Everything else — capture
nudges, load suggestions, curate reminders — is handled by awareness or
surfaced at the right moment.

---

## /kno.capture

**When to use it:** When you want to explicitly capture, or when awareness
didn't nudge for something you want to save. Also useful mid-session
at natural milestones — you can capture multiple times in a long session.

**How it works**

The skill reviews the conversation, proposes a title, summary, and tags,
then asks you to confirm before writing anything. Tags are the primary
signal that load and curate use to match sessions to pages and queries.

```
/kno.capture

Here's what kno will capture from this session:

  Title:    RDS slow query debugging
  Summary:  Query planner regression after minor version upgrade. Fixed by
            pinning parameter group. Key lesson: always test minor upgrades
            in staging.

  Tags:     aws, rds, databases, performance

Save this? [yes / edit / skip]
```

You can steer it conversationally. Hashtags in your message become tags:

```
/kno.capture — make sure to tag this #aws #rds, the parameter group fix
was the key thing
```

**After capturing**, kno briefly connects the note to your vault —
mentioning the page it'll strengthen, or suggesting curate when notes
have accumulated. If sessions cluster around an untracked theme, kno
suggests creating a page.

---

## /kno.curate

**When to use it:** Periodically — when notes have accumulated, or when
you want a page to be current. A reasonable cadence for an active user
is weekly or monthly. kno will let you know when your pages could benefit
from new notes.

**Why it matters:** Notes are raw captures. Curate is where they become
knowledge. The skill reads your uncurated notes, synthesizes what's
new, and updates your page documents following the guidance you've written
into them.

**How it works**

```
/kno.curate

You have 22 uncurated notes. 6 pages.

Pages (by time since last curate):
  1. AWS Infrastructure        — 3 weeks ago
  2. CNC Machine Maintenance   — 3 weeks ago
  3. Customer Onboarding       — 1 month ago
  4. Kubernetes Migration      — never curated

Curate all, or start with one?
```

For each page, the skill synthesizes the update and shows what changed:

```
Curating: AWS Infrastructure

Reading notes...
Following your guidance: "Focus on operational lessons learned the hard
way — config decisions, debugging patterns, cost surprises."

Done. Here's what changed:
  — Added: RDS parameter group pinning after minor version upgrade
  — Updated: ECS drain window recommendation (30s -> 60s)
  — No change: connection pool section already reflected current thinking

Mark 9 notes as curated? [yes / review first]
```

**Notes that fit multiple pages** — the skill asks which pages to include
the note in. **Partial runs** — if uncurated notes exceed the configured
limit, the skill processes what it can and reports the remainder.

**Publishing** — if publish targets are configured, pages are automatically
published after each curate update. The skill confirms this when it happens:
"Pages updated and published." No extra step needed.

---

## /kno.load

**When to use it:** When you want to load a specific page or topic that
awareness hasn't surfaced. Usually you don't need this — `/kno` and
awareness handle loading automatically. Use it when you know exactly
what you want.

```
/kno.load

What are you working on today?

> troubleshooting vibration on CNC mill #3 after the spindle replacement

Found:

  Pages (1):
    CNC Machine Maintenance  — last curated 2 weeks ago

  Recent sessions (2, matched by tags: cnc, spindle, vibration):
    Spindle bearing diagnosis — root cause was worn bearing    3 days ago
    Mill #3 alignment after motor swap                        1 week ago

Load all 3? [yes / pick / skip]
```

You can also load directly: `/kno.load aws infrastructure`

The skill demonstrates understanding of what it loaded — not just "loaded
3 items" but a brief reflection showing it absorbed the content.

---

## /kno.page

**When to use it:** When you want to create, review, or edit a page.
kno will also suggest creating pages when sessions cluster around an
untracked theme.

```
/kno.page new

What area of knowledge do you want to track?

> our customer onboarding process

Good. What should I focus on when updating this page? What can I skip?

> Focus on edge cases, handoff failures, and what we learned from
> difficult onboardings. Skip the standard checklist stuff.

Creating page: Customer Onboarding
Confirm? [yes / edit]
```

After creating a page, the skill checks for relevant uncurated sessions
and offers to curate them immediately — so the page starts with real
knowledge.

---

## /kno.status

**When to use it:** When you want the raw picture of vault health. kno
checks vault status internally during other workflows and surfaces what
matters — you rarely need to call this directly.

```
/kno.status

Vault: ~/kno

Notes: 143 / 500  (357 remaining)
  Curated:      121
  Uncurated:      22

Pages:
  AWS Infrastructure        last curated 3 days ago
  CNC Machine Maintenance   last curated 2 weeks ago
  Customer Onboarding        never curated
  ...
```
