# kno — Architecture and Design Principles

## What kno Is

kno is a local-first knowledge vault for LLM conversations. Its purpose is
to make knowledge compound across sessions — insights, decisions, and context
accumulate over time instead of staying isolated in disconnected conversations.

The vault is plain files on disk — markdown, TOML, and a search index. No
database, no cloud dependency, no proprietary format. You can read every
file kno writes, sync them with any tool, or browse them in a markdown
editor like Obsidian. This is a deliberate trust signal: your knowledge
belongs to you, stored in a format that outlasts any tool.

The vault is client-agnostic. It's backed by a CLI that owns the data and
an MCP server that exposes it. Any MCP-capable client can connect. Active
awareness — kno's ability to notice knowledge checkpoints and offer to
act — is delivered via the MCP `instructions` field at server initialization,
making it available to any client that respects the protocol. Your knowledge
is portable across clients, not locked to any single tool.

---

## The Layers

### CLI — owns the vault

The CLI is the only component that reads and writes the vault. It provides
deterministic, typed CRUD operations against two resources: notes and
pages. It has no opinion about intent, no LLM dependency, and no knowledge
of skills or workflows. Every command has a defined input, a defined output,
and either succeeds or fails with a clear error.

The CLI is fully testable without an LLM anywhere near it.

### MCP Server — exposes the CLI

The MCP server is a thin translation layer. It wraps CLI operations as typed
tools and exposes them to any MCP-capable client. It contains no logic,
makes no decisions, and never touches the vault directly. It calls the vault
layer and returns the result.

The server also delivers **active awareness instructions** at initialization
time via the MCP `instructions` field. These instructions give the connected
client standing awareness of the vault — the ability to recognize knowledge
checkpoints, offer to capture, and suggest loading relevant context. The
instructions are determined by the `nudges.level` config setting.

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

Skills are activated by slash commands. They provide structured, detailed
guidance for complex workflows like curate.

### Active Awareness — standing presence

Active awareness is delivered as MCP server instructions — text sent to
the client at connection time that persists for the entire conversation.
It gives the client the judgment criteria to recognize knowledge checkpoints,
offer to load relevant context, and suggest capture at natural moments.

Awareness and skills are complementary:

- **Awareness** surfaces. It notices the moment and offers an action.
  Nothing happens without user confirmation.
- **Skills** execute. They provide the detailed procedure for capture,
  load, curate, etc.
- **Slash commands** override. They let the user trigger any skill
  explicitly, regardless of whether awareness nudged.

The execution path is identical whether triggered by awareness or slash
command. A capture initiated by a nudge produces the same note as one
initiated by `/kno.capture`.

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
durable artifact is the page document. Notes are created at knowledge
checkpoints — either when the user explicitly captures or when awareness
nudges and the user confirms. The CLI stores them. Curation eventually
consumes them.

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
independent vault — separate notes, pages, search index, config, and
awareness settings.

This is the model for spaces (e.g. work vs. personal). Each vault has its
own MCP registration. The `--name` value becomes the skill prefix: a vault
named `kno-personal` exposes `/kno-personal.capture`, `/kno-personal.load`,
etc. Each vault's awareness operates independently. Separation is enforced
at the filesystem level — vaults have no knowledge of each other. Sync
and encryption are handled outside kno, at the directory level.

---

## The Knowledge Loop

The system supports three primary workflows:

**Capture** — at a knowledge checkpoint (awareness-initiated or
user-initiated), the skill synthesizes a summary and writes it to the vault.
One orient call, one write call.

**Curate** — the skill reads uncurated notes and the current page
document, synthesizes an updated document following the guidance in the page's
content, writes it back, and stamps the consumed notes. Intelligence lives
entirely in the skill. The CLI provides the read and write primitives.

**Load** — at the start of a session (awareness-initiated or user-initiated),
the skill searches the vault for relevant pages and notes, reads them, and
injects the content into the session context. Every session starts informed
rather than cold.

Each loop makes the next load better. Knowledge compounds.

---

## Awareness Configuration

Awareness behavior is controlled by the `nudges.level` config setting:

| Level | Behavior |
|---|---|
| `off` | No awareness instructions sent. Slash commands only. |
| `light` | (default) High-signal checkpoints only. Conservative nudging. |
| `active` | Broader checkpoint recognition. More frequent nudges. |

The setting is read at MCP server startup. Changing it requires restarting
the client to pick up the new instructions.

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
| tags | note | array | skill at capture time |
| summary | note | scalar | skill at capture time |
| curated_at | note | scalar | skill after curate |
| curated_into | note | array | skill after curate |
| (in content) | page | — | user-authored, skill-maintained |
| last_curated_at | page | scalar | skill after curate |
| summary | page | scalar | skill after curate |

Page summaries power topic awareness — they let kno recognize when a new
conversation overlaps with existing vault knowledge without reading full
page content. The curate skill updates summaries on every pass.

The CLI has no knowledge of these keys. The skill decides what metadata
matters and constructs the appropriate commands.

---

## Design Principles

> The CLI owns the vault. The MCP exposes it. The skill interprets intent.
> Awareness initiates at the right moment. Intelligence lives only in the
> skill and awareness layers — never in the CLI.

> Notes have no page association until curate runs. curated_into is
> the only field that creates a note-to-page relationship, and only the
> skill writes it — via the CLI.

> Page content is user-owned. By convention it begins with instructions the
> user has written — guidance the skill follows on every curate pass. A page
> without any guidance will prompt the skill to ask before proceeding.

> The page list is a table of contents for your knowledge — intentional,
> finite, and fully under your control. Pages are never created automatically.

> Active awareness is additive. It offers; it never acts without
> confirmation. The vault is never modified silently.

> The vault is client-agnostic. Any MCP-capable client can connect. Awareness
> instructions are delivered via the MCP protocol, not hardcoded to any
> specific client.

---

## Skill Voice

Skills shape how kno feels to the user. The personality is consistent
across all skills — individual skills specialize (capture is brief,
curate walks through changes) but the underlying character is the same.

**The character:** A knowledgeable colleague who sees value in what you've
done and offers to help you keep it. Not a task manager, not a reminder
system, not a productivity coach. Think of the colleague who says "that
was a good insight" — not the one who says "don't forget to document that."

**Core principles:**

- **Offer, don't assign.** "Want me to add this to your vault?" not
  "You should capture this." The user is never told what to do — they're
  offered something useful. Every suggestion frames the benefit to the
  user's knowledge base, not an obligation to fulfill.

- **Acknowledge growth, don't pressure.** "You're building up good context
  on this" not "you have a backlog." Notes accumulating is a positive sign
  — it means the vault is working. Curate is an opportunity to strengthen
  pages, not a chore that's overdue.

- **Be brief, be warm, get out of the way.** One sentence for a nudge.
  Two or three lines after a capture. The user is here to work — kno
  makes the work compound, it doesn't compete for attention.

- **Teach by doing.** Mention slash commands naturally in context — "that'll
  feed into your page next time you run `/kno.curate`" — so the user
  learns the vocabulary without a tutorial. Never explain what a command
  does unless asked.

- **Respect silence.** If the user declines, drop it. If there's nothing
  useful to offer, say nothing. Quiet confidence beats eager helpfulness.

**Anti-patterns to avoid:**

- Urgency language: "before we move on," "waiting," "backlog," "overdue"
- Productivity framing: "you need to," "don't forget," "you should"
- Over-explaining: justifying why something is worth capturing
- Cheerleading: "Great work!" "Awesome session!" — the user knows what
  they did

Each skill's Voice section builds on this foundation. When writing or
modifying a skill, read this section first, then specialize for the
skill's context.

---

## Cross-Layer Mapping

Every operation flows through the same stack: the user (or skill, or
awareness) calls an MCP tool, which calls the CLI, which reads or writes
the vault.

| CLI Command | MCP Tool | Used by |
|---|---|---|
| `kno note create` | `kno_note_create` | capture (awareness or /kno.capture) |
| `kno note list` | `kno_note_list` | capture, curate |
| `kno note show` | `kno_note_show` | curate, load |
| `kno note update` | `kno_note_update` | curate |
| `kno note delete` | `kno_note_delete` | conversational |
| `kno note search` | `kno_note_search` | load (awareness or /kno.load) |
| `kno note prune` | — | CLI only |
| `kno page create` | `kno_page_create` | /kno.page |
| `kno page list` | `kno_page_list` | /kno.page |
| `kno page show` | `kno_page_show` | curate, load |
| `kno page update` | `kno_page_update` | curate, /kno.page |
| `kno page rename` | `kno_page_rename` | conversational |
| `kno page delete` | `kno_page_delete` | conversational |
| `kno page search` | `kno_page_search` | load (awareness or /kno.load) |
| `kno vault status` | `kno_vault_status` | all skills, awareness |
| `kno vault rebuild-index` | — | CLI only |

**"Conversational"** means the LLM uses the tool directly when the user
asks — no skill needed. Delete and rename are simple enough that a slash
command would add ceremony without value.

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
