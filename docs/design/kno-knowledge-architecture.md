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

The MCP server is a thin translation layer. It wraps CLI operations as typed
tools and exposes them to Claude Desktop. It contains no logic, makes no
decisions, and never touches the vault directly. It calls the vault layer and
returns the result.

Most MCP tools are used by skills — the skill orchestrates a sequence of
tool calls to implement a workflow. Some tools (delete, rename) are used
directly by the LLM without a skill, for simple operations where
conversational handling is sufficient. Bulk maintenance commands (prune,
index rebuild) are CLI-only.

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
of a session. The CLI stores them. Curation eventually consumes them.

Notes are automatically removed when the vault reaches capacity, oldest
curated first. If no curated notes exist, the oldest note is removed
regardless of status — the vault never deadlocks. The response signals when
uncurated knowledge was lost so the skill can warn the user.

### Pages

Pages are curated, living knowledge documents. They are durable. Page
content is user-owned and skill-maintained — the CLI stores and returns it
without interpretation. By convention the skill structures content to begin
with instructions (what to focus on, what to skip, how to handle
contradictions), followed by the accumulated knowledge document. The skill
reads this content before every curate pass and follows the guidance it
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
skill prefix: a vault named `kno-personal` exposes `/kno-personal.capture`,
`/kno-personal.load`, etc. Separation is enforced at the filesystem level
— vaults have no knowledge of each other. Sync and encryption are handled
outside kno, at the directory level.

---

## The Knowledge Loop

The system supports three primary skill workflows:

**Capture** — at the end of a session, the skill synthesizes a summary and
calls the CLI to store it. One orient call, one write call.

**Curate** — the skill reads uncurated notes and the current page
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
| curated_at | note | scalar | skill after curate |
| curated_into | note | array | skill after curate |
| (in content) | page | — | user-authored, skill-maintained |
| last_curated_at | page | scalar | skill after curate |

The CLI has no knowledge of these keys. The skill decides what metadata
matters and constructs the appropriate commands.

---

## Design Principles

> The CLI owns the vault. The MCP exposes it. The skill interprets intent.
> Intelligence lives only in the skill layer — never in the CLI.

> Notes have no page association until curate runs. curated_into is
> the only field that creates a note-to-page relationship, and only the
> skill writes it — via the CLI.

> Page content is user-owned. By convention it begins with instructions the
> user has written — guidance the skill follows on every curate pass. A page
> without any guidance will prompt the skill to ask before proceeding.

> The page list is a table of contents for your knowledge — intentional,
> finite, and fully under your control. Pages are never created automatically.

> Defaults are designed to be predictably successful. A skill operating within
> default limits always knows the upper bound of what it will receive.

---

## Cross-Layer Mapping

Every operation flows through the same stack: the user (or skill) calls
an MCP tool, which calls the CLI, which reads or writes the vault.
Not every CLI command is exposed as an MCP tool. Not every MCP tool has
a dedicated skill. The table below is the complete mapping.

| CLI Command | MCP Tool | Used by |
|---|---|---|
| `kno note create` | `kno_note_create` | /kno.capture |
| `kno note list` | `kno_note_list` | /kno.capture, /kno.curate |
| `kno note show` | `kno_note_show` | /kno.curate, /kno.load |
| `kno note update` | `kno_note_update` | /kno.curate |
| `kno note delete` | `kno_note_delete` | conversational |
| `kno note search` | `kno_note_search` | /kno.load |
| `kno note prune` | — | CLI only |
| `kno page create` | `kno_page_create` | /kno.page |
| `kno page list` | `kno_page_list` | /kno.page |
| `kno page show` | `kno_page_show` | /kno.curate, /kno.load |
| `kno page update` | `kno_page_update` | /kno.curate, /kno.page |
| `kno page rename` | `kno_page_rename` | conversational |
| `kno page delete` | `kno_page_delete` | conversational |
| `kno page search` | `kno_page_search` | /kno.load |
| `kno vault status` | `kno_vault_status` | all skills |
| `kno vault rebuild-index` | — | CLI only |

**"Conversational"** means the LLM uses the tool directly when the user
asks — no skill needed. Delete and rename are simple enough that a
slash command would add ceremony without value.

**"CLI only"** means the operation is not exposed via MCP. Bulk
maintenance (prune) and index repair are terminal operations.

### Skill workflows

Each skill is a sequence of tool calls. The skill decides which tools
to call and in what order based on the conversation.

**Capture:** orient (`vault status`) → write (`note create`)

**Curate:** orient (`vault status`) → list uncurated (`note list`) →
read relevant notes (`note show`) → read page (`page show`) →
synthesize → write page (`page update`) → stamp notes (`note update`)

**Load:** orient (`vault status`) → search (`page search`, `note search`) →
read matches (`page show`, `note show`) → inject into conversation

**Page:** orient (`vault status`) → create or update (`page create`,
`page update`, `page rename`)

**Status:** orient (`vault status`)
