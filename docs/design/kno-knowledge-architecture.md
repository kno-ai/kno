# kno — Architecture and Design Principles

## What kno Is

kno is a local-first knowledge vault for LLM conversations. Its purpose is
to make knowledge compound across sessions — insights, decisions, and context
accumulate over time instead of vanishing with each chat.

The system has three layers, each with a single responsibility.

---

## The Three Layers

### CLI — owns the vault

The CLI is the only component that reads and writes the vault. It provides
deterministic, typed CRUD operations against two resources: notes and
pages. It has no opinion about intent, no LLM dependency, and no knowledge
of skills or workflows. Every command has a defined input, a defined output,
and either succeeds or fails with a clear error.

The CLI is fully testable without an LLM anywhere near it.

### MCP Server — exposes the CLI

The MCP server is a thin translation layer. It wraps CLI commands as typed
tools and exposes them to Claude Desktop. It contains no logic, makes no
decisions, and never touches the vault directly. It calls the CLI and returns
the result.

Admin commands are not exposed via MCP — they are CLI-only.

### Skill — interprets intent

The skill is where intelligence lives. It receives slash commands and
conversational input from the user, interprets intent, and translates that
intent into precise CLI calls via MCP. The skill asks questions when needed,
handles ambiguity, and synthesizes user input into the right sequence of
commands at the right moment.

The skill's output is always a CLI command. That is the unit of work.

---

## The Contract

> The CLI is 100% predictable, deterministic, and testable. The skill is
> conversational. The boundary between them is absolute — intelligence lives
> in the skill, determinism lives in the CLI, and neither crosses into the
> other's domain.

This contract has three consequences:

**Testability.** The CLI can be fully unit tested with no mocks and no LLM.
The skill can be evaluated independently — does it produce the right CLI
commands given a conversation? These are separate test surfaces with no
overlap.

**Transparency.** Every change to the vault is traceable to a specific CLI
command. Nothing is ever modified by anything opaque.

**Replaceability.** The skill is replaceable. A better model, a different
interface, a new slash command syntax — none of that touches the CLI. The
vault stays stable underneath whatever skill layer evolves on top of it.

---

## The Two Resources

### Notes

Notes are structured summaries of LLM sessions. They are ephemeral — the
durable artifact is the page document. The skill creates notes at the end
of a session. The CLI stores them. Distillation eventually consumes them.

Notes are automatically removed when the vault reaches capacity, oldest
distilled first. If no distilled notes exist, the oldest note is removed
regardless of status — the vault never deadlocks. The response signals when
undistilled knowledge was lost so the skill can warn the user.

### Pages

Pages are curated, living knowledge documents. They are durable. Page
content is user-owned and skill-maintained — the CLI stores and returns it
without interpretation. By convention the skill structures content to begin
with instructions (what to focus on, what to skip, how to handle
contradictions), followed by the accumulated knowledge document. The skill
reads this content before every distill pass and follows the guidance it
finds there.

Pages are created intentionally by the user. They are never created
automatically. This keeps the page list finite and owned.

---

## Multiple Vaults

A vault is a self-contained directory. Running `kno setup` a second time
with a different `--name` and `--vault` path creates a completely
independent vault — separate notes, pages, search index, and config.

This is the model for spaces (e.g. work vs. personal). Each vault has its
own MCP registration in Claude Desktop. The `--name` value becomes the
skill prefix: a vault named `kno-personal` exposes `/kno-personal.save`,
`/kno-personal.load`, etc. Separation is enforced at the filesystem level
— vaults have no knowledge of each other. Sync and encryption are handled
outside kno, at the directory level.

---

## The Knowledge Loop

The system supports three primary skill workflows:

**Save** — at the end of a session, the skill synthesizes a summary and
calls the CLI to store it. One orient call, one write call.

**Distill** — the skill reads undistilled notes and the current page
document, synthesizes an updated document following the guidance in the page's content,
writes it back, and stamps the consumed notes. Intelligence lives entirely
in the skill. The CLI provides the read and write primitives.

**Load** — at the start of a session, the skill searches the vault for
relevant pages and notes, reads them, and injects the content into the
new session context. Every session starts informed rather than cold.

Each loop makes the next load better. Knowledge compounds.

---

## Metadata Model

Both notes and pages support arbitrary key-value metadata. The CLI stores
and returns metadata without interpretation — it has no knowledge of what any
key means.

Multi-value fields use repeated --meta flags. A single flag produces a
scalar value. Duplicate flags for the same key produce an array. Filter
operations perform exact match on scalars and contains check on arrays.

Key fields used by the skill:

| Key | Resource | Type | Set by |
|---|---|---|---|
| tags | note | array | skill at note time |
| summary | note | scalar | skill at note time |
| distilled_at | note | scalar | skill after distill |
| distilled_into | note | array | skill after distill |
| (in content) | page | — | user-authored, skill-maintained |
| last_distilled_at | page | scalar | skill after distill |

The CLI has no knowledge of these keys. The skill decides what metadata
matters and constructs the appropriate commands.

---

## Named Vaults

A vault is a self-contained directory — notes, pages, search index, and
config all live inside it. Running `kno setup` a second time with a different
`--name` and `--vault` path creates a completely independent vault with its
own MCP registration.

This is the spaces model. Work and personal knowledge live in separate vaults,
each with its own MCP server name. In Claude Desktop the skill commands are
prefixed by the vault name:

```
kno setup                                  → /kno.save, /kno.load, /kno.distill
kno setup --name kno-personal --vault ~/kno-personal  → /kno-personal.save, etc.
```

No data crosses between vaults. Sync, encryption, and backup policies are
handled at the directory level — each vault path is just a directory that the
user controls. kno has no opinion about what happens to it.

---

## Design Principles

> The CLI owns the vault. The MCP exposes it. The skill interprets intent.
> Intelligence lives only in the skill layer — never in the CLI.

> Notes have no page association until distill runs. distilled_into is
> the only field that creates a note-to-page relationship, and only the
> skill writes it — via the CLI.

> Page content is user-owned. By convention it begins with instructions the
> user has written — guidance the skill follows on every distill pass. A page
> without any guidance will prompt the skill to ask before proceeding.

> The page list is a table of contents for your knowledge — intentional,
> finite, and fully under your control. Pages are never created automatically.

> Defaults are designed to be predictably successful. A skill operating within
> default limits always knows the upper bound of what it will receive.
