# kno — Developer Guide

kno works the same way in software development as anywhere else —
save, curate, load. But when the kno MCP server detects a git repo,
it adapts. Everything you save gets tagged with the repo name. kno
uses a vocabulary that fits development work. Pages get guidance
tailored to how engineering knowledge actually looks — decisions with
dates, known issues with status, setup that somebody had to figure
out the hard way.

---

## Why kno for development

Knowledge walks out the door when engineers leave. Design decisions,
known issues, non-obvious setup steps, workarounds for upstream bugs —
this context lives in people's heads or buried in old PRs that nobody
searches. The usual fix is to write documentation, but documentation is
a separate step from the work, so it doesn't happen.

kno saves this knowledge as a byproduct of working. When you debug a
race condition and find the fix, kno notices and offers to save it.
When you decide to use Postgres over DynamoDB and talk through the
tradeoffs, that reasoning gets preserved. You're not writing docs —
you're saying yes to a prompt.

Over time, a project page accumulates the knowledge that makes a
codebase legible: why things are the way they are, what's broken, how
to set it up. New team members load that page at the start of a session
and skip the cold start.

---

## How it works

The kno MCP server detects `.git` in the working directory at session
start. No new commands. The same save/curate/load loop runs, enriched
with project context — repo name, dev-specific types, status tracking.

When kno detects a git repo, it confirms:

> kno active — detected: cloud-infra

Everything you save from this session gets tagged with the repo name
automatically.

> **Slash command naming:** Claude Code uses a colon separator for
> commands — `/kno:start`, `/kno:capture`, etc. — instead of the dot
> used by Claude Desktop (`/kno.start`). The commands work the same way.

---

## What changes in dev context

### Automatic repo tagging

Every save in a dev session gets the repo name as a tag automatically.
You don't need to add `#cloud-infra` — it's already there.
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

### Developer page guidance

When you create a page in a git context, kno applies a developer-focused
guidance template. It tells curate to track decisions with dates,
maintain issue status, and keep setup instructions current:

```markdown
<!-- Guidance -->
Track architecture decisions with dates and reasoning. Maintain known
issues with status (open/resolved). Keep setup instructions current.
When something gets fixed, move it to solved with the resolution date.

## Decisions

- **2026-02-14** — Chose RDS Postgres over DynamoDB for config storage.
  Query patterns are relational, access volume fits a single instance,
  and the team already knows Postgres.

## Known issues

- **open** — Flaky timeout on cross-AZ health check after 3 retries.
  Workaround: increase timeout to 30s in config. See #412.
- **resolved (2026-03-01)** — Race condition in ECS task placement.
  Fixed by switching to spread placement strategy.

## Setup

- Requires Docker for local Postgres: `docker compose up -d`
- Seed data: `make seed` — takes ~2 minutes

## Solved problems

- Connection pool exhaustion under load — fixed by setting max pool
  size to 20 per instance and adding connection timeout of 5s.
```

This guidance evolves — update it as you learn what matters for your
project.

---

## Auto-load with `.kno`

When you create a project page in a git repo, kno offers to bind it
to a `.kno` file. Once bound, every future `/kno:start` in that repo
loads your project page instantly — decisions, known issues, setup,
all there before you write a line of code.

This is the general `.kno` feature (see the
[User Guide](kno-guide#auto-loading-pages-with-kno)) applied to
development. Git gives kno a stable project identifier, so it can
suggest a page name and offer to bind it at the right moment.

If you'd rather not be prompted about `.kno` setup, tell kno to stop
asking. The preference is saved in your vault config and applies across
all repos.

---

## Team use

When `.kno` is committed to the repo, every developer who clones it
gets auto-load for the bound page name. Each developer needs the
matching page in their own vault — when they do, `/kno.start` loads
it instantly. New team members start their first session with decisions,
known issues, and setup already loaded.

Add `.kno` to `.gitignore` if you prefer to keep it personal.

---

## Reference

For the general user guide, see the [User Guide](kno-guide). For the
complete CLI specification, see the [CLI Reference](kno-cli). For skill
behavior, see the [Skills Reference](kno-skills).

---

Start every session with `/kno.start`. Your project knowledge loads automatically.
