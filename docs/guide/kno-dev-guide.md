# kno — Developer Guide

kno works the same way in software development as anywhere else —
save, curate, load. But when you're working inside a git repo with
Claude Code, kno detects the project context and adapts. Everything
you save gets tagged with the repo name. kno uses a vocabulary that
fits development work. Pages get guidance tailored to how engineering
knowledge actually looks — decisions with dates, known issues with
status, setup that somebody had to figure out the hard way.

---

## Why kno for development

Knowledge walks out the door when engineers leave. Design decisions,
known issues, non-obvious setup steps, workarounds for upstream bugs —
this context lives in people's heads or buried in old PRs that nobody
searches. The usual fix is to write documentation, but documentation is
a separate step from the work, so it doesn't happen.

kno saves this knowledge as a byproduct of working. When you debug a
race condition in Claude Code and find the fix, kno notices and offers
to save it. When you decide to use Postgres over DynamoDB and talk
through the tradeoffs, that reasoning gets preserved. You're not
writing docs — you're saying yes to a prompt.

Over time, a project page accumulates the knowledge that makes a
codebase legible: why things are the way they are, what's broken, how
to set it up. New team members load that page at the start of a session
and skip the cold start.

---

## How it works

When you start a session in a directory with a `.git` folder, kno
detects the project automatically. No new commands. The same
save/curate/load loop runs, enriched with project awareness — repo
name, dev-specific types, status tracking.

Dev features require Claude Code. Claude Code has direct filesystem
access, which is how kno detects `.git` and reads the `.kno` config
file. Claude Desktop gets the standard kno experience without project
detection.

---

## What changes in dev context

### Session confirmation

When kno detects a git repo, it confirms at session start:

> kno active — detected: payments-service

This tells you the project context is live. Everything you save from
this session will be tagged with the repo name.

### Automatic repo tagging

Every save in a dev session gets the repo name as a tag automatically.
You don't need to add `#payments-service` — it's already there.
Additional tags work the same as always.

### Type vocabulary

Dev context adds types that fit engineering work:

| Type | Use for |
|---|---|
| `decision` | Architecture choices, library selections, tradeoff reasoning |
| `debt` | Known shortcuts, things to fix later |
| `runbook` | Operational procedures, deploy steps, incident response |
| `bug` | Bugs found, root causes, workarounds |
| `dependency` | Upstream constraints, version pins, API quirks |

These aren't exclusive — standard types still work. Types help
curate organize knowledge into the right sections of a project page.

### Status tracking

`debt` and `bug` types support status: **open** or **resolved**. When
you fix a known issue, mark it resolved and curate will update the page
accordingly — moving it to a solved section or noting the resolution
date.

---

## The `.kno` file

A `.kno` file at the repo root holds project-specific settings. kno
creates it when you accept the auto-load preference during a session.

```toml
[skill]
auto_load_on_confirm = true
```

Commit it to share settings with your team. Or add it to `.gitignore`
to keep it personal — either works.

---

## Project pages

A good project page reads like the document you wish existed when you
joined the team:

```markdown
<!-- Guidance -->
Track architecture decisions with dates and reasoning. Maintain known
issues with status (open/resolved). Keep setup instructions current.
When something gets fixed, move it to solved with the resolution date.

## Decisions

- **2026-02-14** — Chose Postgres over DynamoDB for order storage.
  Query patterns are relational, access volume fits a single RDS
  instance, and the team already knows Postgres.
- **2026-01-20** — Event bus: SNS+SQS over Kafka. Lower operational
  overhead, sufficient throughput for current scale.

## Known issues

- **open** — Flaky timeout on webhook retry after 3 attempts.
  Workaround: increase timeout to 30s in config. See #412.
- **resolved (2026-03-01)** — Race condition in order lock acquisition.
  Fixed by switching to advisory locks.

## Setup

- Requires Docker for local Postgres: `docker compose up -d`
- Seed data: `make seed` — takes ~2 minutes
- `.env.example` has all required vars. Copy to `.env` and fill in
  Stripe test keys from 1Password.

## Solved problems

- Connection pool exhaustion under load — fixed by setting max pool
  size to 20 per instance and adding connection timeout of 5s.
```

The guidance comment at the top shapes how curate maintains the page.
Developer-specific guidance tells curate to track decisions with dates,
maintain issue status, and keep setup instructions current. This
guidance evolves — update it as you learn what matters for your project.

---

## Team use

When `.kno` is committed to the repo, every developer who clones it
gets the same project settings. Combined with a shared vault page for
the project, this means onboarding context that loads automatically —
new team members start their first Claude Code session with decisions,
known issues, and setup already loaded.

---

## Settings

The `.kno` file supports these settings under `[skill]`:

```toml
# .kno — project-specific kno settings
# Commit to share with your team, or add to .gitignore to keep personal.

[skill]
auto_load_on_confirm = true    # true | false (omit to leave unset)
nudge_level = "active"         # "off" | "light" | "active" (omit to use vault default)
```

### `auto_load_on_confirm`

Controls whether kno loads the matching project page automatically after
confirming the repo detection.

- **`true`** — Load immediately on confirm, no extra prompt.
- **`false`** — Standard flow. kno won't offer auto-load.
- **omitted** — Standard flow. After the first time you confirm a load
  suggestion, kno offers to save your preference for this project.

### `nudge_level`

Override the vault-wide nudge level for a specific repo. Useful when you
want kno to be more attentive in a project you're actively building,
while keeping your global setting at `light`. Values: `off`, `light`,
`active`. Omit to inherit from your vault config.

---

## Client support

Dev features — `.git` detection, `.kno` file reading, repo-aware
tagging — require Claude Code. Claude Code has direct filesystem
access, which is what makes project detection possible.

Claude Desktop connects to kno through MCP and gets the full
save/curate/load experience, but without project context. If you
use both clients, your vault is shared — pages curated from Claude Code
sessions are available in Claude Desktop and vice versa.

---

## Reference

For the general user guide, see the [User Guide](kno-guide). For the
complete CLI specification, see the [CLI Reference](kno-cli). For skill
behavior, see the [Skills Reference](kno-skills).
