# kno CLI Reference

kno is a local-first knowledge vault for LLM conversations. The CLI provides
deterministic, testable CRUD operations against the vault. No LLM calls are
made by the CLI — intelligence lives in the skill layer above it.

All commands support `--json` for machine-readable output. Human-readable
output is the default. The MCP server exposes all commands except those in
the ADMIN namespace.

---

## ERROR CONTRACT

All commands follow these conventions.

**Exit codes**

- `0` — success
- `1` — error (not found, validation failure, vault error)

**Output streams**

- Normal output → stdout
- Error messages → stderr

**Error format (default)**

```
✗  <message>
```

**Error format (--json)**

```json
{"error": "<message>"}
```

JSON errors are written to stderr with exit code 1.

**Common errors**

```
✗  Not found: note x9f200
✗  Not found: page b81e44
✗  --title is required
✗  --name is required
✗  --count is required
```

Commands that accept IDs fail immediately on the first unrecognized ID and
make no changes to the vault. Partial writes do not occur.

---

## SETUP

Setup is a one-time installation step. It is not part of the MCP surface and
not part of the admin namespace — it is a top-level command because it is the
first thing every user runs and friction here costs adoption.

---

### kno setup

```
kno setup  [--name <name>]  [--vault <path>]  [--no-claude-desktop]
```

Initialize a kno vault and register it as an MCP server with Claude Desktop.
Running `kno setup` once is all a first-time user needs. Running it again
with a different `--name` and `--vault` creates a second independent vault
without affecting the first — this is how multiple spaces (e.g. work vs.
personal) are managed.

Running setup a second time with a different `--name` and `--vault` creates
an independent vault — a separate space with its own pages, notes, search
index, and MCP registration. Existing vaults are not affected.

**What it does**

1. Creates the vault directory (default: `~/kno`)
2. Writes a default `config.toml` to the vault
3. Detects Claude Desktop and registers the vault as an MCP server using
   the provided name (default: `kno`)
4. Prints a confirmation summary and next steps

If Claude Desktop is not installed, step 3 is skipped silently. The vault
is still created and fully functional via the CLI.

If `kno setup` has already been run, it is safe to run again — it will not
overwrite an existing vault or config, but will re-register the MCP server
if Claude Desktop is detected.

**Options**

    --name <name>           Name for this vault's MCP server registration.
                            Defaults to `kno`. Use a different name for
                            each additional vault (e.g. `kno-personal`).
                            The name determines the skill prefix in Claude
                            Desktop — a vault named `kno-personal` exposes
                            /kno-personal.save, /kno-personal.load, etc.

    --vault <path>          Vault directory path. Default: ~/kno for
                            the first vault. Use a distinct path for each
                            additional vault (e.g. ~/kno-personal).

    --name <name>           MCP server name registered with Claude Desktop.
                            Default: kno. Use a distinct name for each
                            additional vault (e.g. kno-personal). The name
                            becomes the skill prefix in Claude Desktop:
                            /kno-personal.save, /kno-personal.load, etc.

    --no-claude-desktop     Skip Claude Desktop detection and MCP
                            registration. Useful in headless or CI
                            environments.

**Output**

```
✓  Vault created at ~/kno
✓  Config written to ~/kno/config.toml
✓  MCP server "kno" registered with Claude Desktop  (name: kno)

Restart Claude Desktop to activate kno skills.

Quick start:
  /kno.save    — save a session summary to your vault
  /kno.load       — load knowledge into a new session
  kno note list  — browse your vault from the terminal
```

**Output (Claude Desktop not found)**

```
✓  Vault created at ~/kno
✓  Config written to ~/kno/config.toml
—  Claude Desktop not found — skipping MCP registration

To register manually, add the following to your Claude Desktop config:

  {
    "mcpServers": {
      "<name>": {
        "command": "kno",
        "args": ["--vault", "<vault-path>", "mcp"]
      }
    }
  }
```

**Notes**

- The MCP server is registered by writing to Claude Desktop's
  `claude_desktop_config.json`. On macOS this is:
  `~/Library/Application Support/Claude/claude_desktop_config.json`
- Claude Desktop must be restarted after setup for the MCP server to activate
- Running `kno setup` a second time with a different `--name` and `--vault`
  creates an independent vault without affecting existing ones. Each vault
  gets its own MCP registration, config, and search index. Sync and
  encryption are handled at the directory level — kno does not manage them.
- The vault path can be changed later by editing `~/kno/config.toml` and
  re-running `kno setup`
- Each named vault is fully independent — separate notes, pages, search
  index, and config. Use this for work/personal separation, different sync
  or encryption policies, or any context where data should not cross over.

---

## NOTES

Notes are structured summaries of LLM sessions. They are ephemeral by
design — the durable knowledge artifact is the page document.

**Note lifecycle**

1. Created by the skill at the end of a session
2. Available for load and distill while undistilled
3. Marked as distilled after a distill pass (`distilled_at`, `distilled_into`)
4. Automatically removed when the vault reaches capacity (`notes.max_count`)

Auto-removal algorithm on `note create` when at capacity:
1. Remove the oldest note where `distilled_at` is set (preferred — this
   knowledge is already preserved in a page document)
2. If no distilled notes exist, remove the oldest note regardless of
   distill status. The response includes `auto_removed_undistilled: true` to
   signal that undistilled knowledge was lost.

A new note can always be written. The vault never deadlocks.

---

### kno note list

```
kno note list  [--filter <key>=<value>]...  [--limit <n>]  [--json]
```

List notes in the vault, newest first.

**Options**

    --filter <key>=<value>    Filter by metadata value. Repeatable.
                              Exact match for scalar values.
                              Contains check for array values.
                              Use the value `null` to match records where
                              the key is absent or set to JSON null — this
                              mirrors the null value shown in JSON output.
                              Example: --filter distilled_at=null
    --limit <n>               Maximum results to return.
                              Default: notes.default_list_limit (config)
    --json                    Machine-readable output

**Output (default)**

```
ID        TITLE                            CREATED       STATUS
x9f200    RDS slow query debugging          2026-03-05    not distilled
a3b100    SQS dead letter queue             2026-02-20    distilled
c7d400    EFT file processing               2026-02-10    distilled
```

**Output (--json)**

```json
[
  {
    "id": "x9f200",
    "title": "RDS slow query debugging",
    "metadata": {
      "tags": ["aws", "databases", "performance"],
      "summary": "Query planner regression after minor version upgrade. Fixed by pinning parameter group. Key lesson: always test minor upgrades in staging.",
      "distilled_at": null,
      "distilled_into": null
    },
    "created_at": "2026-03-05T14:22:00Z"
  }
]
```

Note: `summary` lives in `metadata` and is always included in JSON output
when present — it is the primary field the skill uses for routing and load
decisions without requiring a `note show` call. Multi-value metadata
fields are returned as JSON arrays. Single-value fields are returned as
scalars. Absent or unset fields appear as JSON null.

---

### kno note show

```
kno note show <id> [<id>...]  [--json]
```

Show full content of one or more notes. Accepts multiple ids for bulk
read in a single call.

**Options**

    --json    Machine-readable output

**Output (default)**

```
━━━ RDS slow query debugging  [x9f200]  2026-03-05 ━━━━━━━━━━━━━━━━━━━━━━━━

Query planner regression after RDS 14.3 → 14.4 minor version upgrade...
[full note content]

tags: aws, databases, performance
distilled_at: —
```

**Output (--json)**

```json
[
  {
    "id": "x9f200",
    "title": "RDS slow query debugging",
    "content": "...",
    "metadata": {
      "tags": ["aws", "databases", "performance"],
      "summary": "Query planner regression after minor version upgrade...",
      "distilled_at": null,
      "distilled_into": null
    },
    "created_at": "2026-03-05T14:22:00Z"
  }
]
```

Note: Always returns an array, even for a single id. This ensures consistent
handling by the skill regardless of how many notes are requested.

---

### kno note create

```
kno note create --title <title>  [--meta <key>=<value>]...  < <content>
```

Create a new note. Content is read from stdin and is required. Title is
required. Metadata is optional. The vault index is updated automatically.

If the vault is at capacity (`notes.max_count`), the oldest distilled
note is automatically removed before writing the new one. If no distilled
notes exist, the oldest note (regardless of status) is removed instead.
The response indicates what was removed and whether it was undistilled.

**Options**

    --title <title>           Required. Display title for the note.
    --meta <key>=<value>      Optional metadata. Repeatable. Duplicate keys
                              produce an array — single key produces a scalar.
                              Common keys:
                                tags       one --meta per tag value
                                summary    short summary (auto-generated
                                           by skill if not provided)

**Output (default)**

```
Created: RDS slow query debugging  [x9f200]
```

When a distilled note was auto-removed to make room:

```
Created: RDS slow query debugging  [x9f200]
Removed: EFT file processing       [c7d400]  (oldest distilled — distill backlog reminder)
```

When an undistilled note was auto-removed (no distilled notes available):

```
Created: RDS slow query debugging  [x9f200]
Removed: SQS visibility timeout    [f9e300]  (oldest — UNDISTILLED, knowledge may be lost. Run /kno.distill)
```

**Output (--json)**

```json
{
  "id": "x9f200",
  "title": "RDS slow query debugging",
  "created_at": "2026-03-05T14:22:00Z",
  "auto_removed": null
}
```

When a distilled note was auto-removed:

```json
{
  "id": "x9f200",
  "title": "RDS slow query debugging",
  "created_at": "2026-03-05T14:22:00Z",
  "auto_removed": "c7d400"
}
```

When an undistilled note was auto-removed:

```json
{
  "id": "x9f200",
  "title": "RDS slow query debugging",
  "created_at": "2026-03-05T14:22:00Z",
  "auto_removed": "f9e300",
  "auto_removed_undistilled": true
}
```

---

### kno note update

```
kno note update <id>  [--meta <key>=<value>]...  [< <content>]
```

Update an existing note. Content and metadata are both optional —
provide either or both. Piping content replaces the full note content.
Unspecified metadata keys are unchanged. Specified keys are replaced.

**Options**

    --meta <key>=<value>    Update metadata. Repeatable. Always replaces
                            the existing value for that key. Duplicate keys
                            produce an array. Single key produces a scalar.
                            Skill must read before writing when appending
                            to an existing array (e.g. distilled_into).

**Examples**

```bash
# stamp distillation into one page
kno note update x9f200 \
  --meta distilled_at=2026-03-05T14:22:00Z \
  --meta distilled_into=b81e44

# stamp distillation into two pages — skill reads first, then writes all values
kno note update x9f200 \
  --meta distilled_at=2026-03-05T14:22:00Z \
  --meta distilled_into=b81e44 \
  --meta distilled_into=c90d12

# update content only
echo "<revised content>" | kno note update x9f200

# update content and tags
echo "<revised content>" | kno note update x9f200 \
  --meta tags=aws \
  --meta tags=rds
```

**Output (default)**

```
Updated: RDS slow query debugging  [x9f200]
```

**Output (--json)**

```json
{
  "id": "x9f200",
  "updated_at": "2026-03-05T15:00:00Z"
}
```

---

### kno note search

```
kno note search <query>  [--filter <key>=<value>]...  [--limit <n>]  [--json]
```

Full-text search across note content and titles. Returns ranked results
with summaries. Metadata filters are applied on top of search results.

**Options**

    --filter <key>=<value>    Filter results by metadata value. Repeatable.
                              Exact match for scalars, contains check for
                              arrays. Use `null` to match absent or JSON null
                              values. Applied after full-text ranking.
    --limit <n>               Maximum results. Default: search.default_limit
    --json                    Machine-readable output

**Output (default)**

```
ID        TITLE                          SCORE   STATUS
x9f200    RDS slow query debugging       0.92    not distilled
f3a100    Aurora connection pool tuning  0.81    distilled
```

**Output (--json)**

```json
[
  {
    "id": "x9f200",
    "title": "RDS slow query debugging",
    "score": 0.92,
    "metadata": {
      "tags": ["aws", "databases", "performance"],
      "summary": "Query planner regression after minor version upgrade...",
      "distilled_at": null,
      "distilled_into": null
    },
    "created_at": "2026-03-05T14:22:00Z"
  }
]
```

Note: `summary` in `metadata` is always included in JSON output when present
so the skill can make load decisions without follow-up `show` calls.

---

## PAGES

Pages are curated, living knowledge documents. They are durable — notes
are ephemeral, pages are not. Page content is user-owned and skill-maintained.
The skill typically structures content to include instructions alongside the
accumulated knowledge document, but the CLI stores and returns content without
interpretation.

---

### kno page list

```
kno page list  [--filter <key>=<value>]...  [--json]
```

List all pages. Pages are finite and curated — no limit is applied.

**Options**

    --filter <key>=<value>    Filter by metadata value. Repeatable.
                              Exact match for scalars, contains check for
                              arrays. Use `null` to match absent or JSON null
                              values.
    --json                    Machine-readable output

**Output (default)**

```
ID        NAME                    LAST DISTILLED
b81e44    AWS Infrastructure      2026-03-01
a3f9c2    Payment Processing      2026-02-15
c90d12    Homebrewing             —
```

**Output (--json)**

```json
[
  {
    "id": "b81e44",
    "name": "AWS Infrastructure",
    "metadata": {
      "last_distilled_at": "2026-03-01T10:00:00Z"
    },
    "created_at": "2026-01-15T09:00:00Z"
  }
]
```

---

### kno page show

```
kno page show <id>  [--json]
```

Show full page document and metadata.

**Options**

    --json    Machine-readable output

**Output (default)**

```
━━━ AWS Infrastructure  [b81e44]  last distilled 2026-03-01 ━━━━━━━━━━━━

## AWS Infrastructure — Current Understanding
...
[full page content]
```

**Output (--json)**

```json
{
  "id": "b81e44",
  "name": "AWS Infrastructure",
  "content": "## AWS Infrastructure — Current Understanding\n...",
  "metadata": {
    "last_distilled_at": "2026-03-01T10:00:00Z"
  },
  "created_at": "2026-01-15T09:00:00Z"
}
```

Note: Returns a single object, not an array. Unlike `note show` which
supports multiple IDs and always returns an array, `page show` takes one
ID and returns one object.

---

### kno page create

```
kno page create --name <name>  [--meta <key>=<value>]...  [< <content>]
```

Create a new page. Name is required. Initial content is optional — if not
provided the page starts empty and is populated on the first distill pass.

Page content is user-owned and skill-maintained. The skill typically
structures content to include instructions (what to focus on, what to skip,
how to handle contradictions) alongside the accumulated knowledge document.
The CLI stores and returns content without interpretation.

**Options**

    --name <name>           Required. Display name for the page.
    --meta <key>=<value>    Optional metadata. Repeatable.

**Output (default)**

```
Created: AWS Infrastructure  [b81e44]
```

**Output (--json)**

```json
{
  "id": "b81e44",
  "name": "AWS Infrastructure",
  "created_at": "2026-01-15T09:00:00Z"
}
```

---

### kno page update

```
kno page update <id>  [--meta <key>=<value>]...  [< <content>]
```

Update an existing page. Content and metadata are both optional — provide
either or both. Piping content replaces the full page content.

**Options**

    --meta <key>=<value>    Update metadata. Repeatable. Always replaces
                            the existing value for that key. Duplicate
                            keys produce an array.

**Examples**

```bash
# update content — primary distill write-back
echo "<updated content>" | kno page update b81e44

# update metadata only
kno page update b81e44 --meta last_distilled_at=2026-03-05T14:22:00Z

# update both — typical distill write-back
echo "<updated content>" | kno page update b81e44 \
  --meta last_distilled_at=2026-03-05T14:22:00Z
```

**Output (default)**

```
Updated: AWS Infrastructure  [b81e44]
```

**Output (--json)**

```json
{
  "id": "b81e44",
  "updated_at": "2026-03-05T14:22:00Z"
}
```

---

### kno page search

```
kno page search <query>  [--filter <key>=<value>]...  [--limit <n>]  [--json]
```

Full-text search across page content and names.
Returns ranked results.

**Options**

    --filter <key>=<value>    Filter results by metadata value. Repeatable.
                              Exact match for scalars, contains check for
                              arrays. Use `null` to match absent or JSON null.
    --limit <n>               Maximum results. Default: search.default_limit
    --json                    Machine-readable output

**Output (default)**

```
ID        NAME                  SCORE
b81e44    AWS Infrastructure    0.95
a3f9c2    Payment Processing    0.42
```

**Output (--json)**

```json
[
  {
    "id": "b81e44",
    "name": "AWS Infrastructure",
    "score": 0.95,
    "excerpt": "...RDS performance and connection pool tuning..."
  }
]
```

---

## VAULT

---

### kno vault status

```
kno vault status  [--json]
```

Return a snapshot of vault health, capacity, and configuration. Exposed via
MCP — the skill calls this at the start of distill and load operations to
orient before acting. Also useful as a power user overview.

**Output (default)**

```
Vault: ~/kno

Notes: 143 / 200  (57 remaining)
  Distilled:    121
  Undistilled:   22

Pages: 6
  b81e44    AWS Infrastructure       last distilled 3 days ago
  a3f9c2    Payment Processing       last distilled 2 weeks ago
  c90d12    Homebrewing              last distilled 1 month ago
  d11e33    Vasculitis Treatment     never distilled
  e22f44    EFT Processing           last distilled 6 days ago
  f33a55    MySQL Optimization       last distilled 3 weeks ago

Config:
  notes.max_count           200
  notes.default_list_limit   50
  notes.summary_max_tokens  100
  pages.max_content_tokens   8000
  distill.max_notes_per_run  50
  search.default_limit           5
```

**Output (--json)**

```json
{
  "vault_path": "/Users/kevin/kno",
  "notes": {
    "total": 143,
    "max_count": 200,
    "remaining": 57,
    "distilled": 121,
    "undistilled": 22
  },
  "pages": [
    {
      "id": "b81e44",
      "name": "AWS Infrastructure",
      "metadata": {
        "last_distilled_at": "2026-03-02T10:00:00Z"
      }
    },
    {
      "id": "a3f9c2",
      "name": "Payment Processing",
      "metadata": {
        "last_distilled_at": "2026-02-19T10:00:00Z"
      }
    }
  ],
  "config": {
    "notes": {
      "max_count": 200,
      "default_list_limit": 50,
      "summary_max_tokens": 100
    },
    "pages": {
      "max_content_tokens": 8000
    },
    "distill": {
      "max_notes_per_run": 50
    },
    "search": {
      "default_limit": 5
    }
  }
}
```

Note: Config values are included in JSON output so the skill can plan
context usage and enforce size budgets without reading config files directly.
`vault_path` is always the fully resolved absolute path — never a
tilde-expanded shorthand.

---

## ADMIN

Admin commands support vault maintenance and are not exposed via the MCP
server. They are available to power users and operators via the CLI only.
The skill layer has no access to these commands.

---

### kno admin prune

```
kno admin prune --count <n>  [--dry-run]  [--json]
```

Remove the N oldest notes from the vault regardless of distill status.
This is a last-resort maintenance command for when the vault is at capacity
and no distilled notes exist to auto-remove. Prefer letting the vault
self-regulate via auto-removal of distilled notes on `note create`.

The algorithm removes notes strictly by age — oldest first, no other
criteria. No judgment about value or content. `--dry-run` shows what would
be removed without deleting anything.

**Options**

    --count <n>    Required. Number of notes to remove.
    --dry-run      Preview removals without deleting. Recommended before
                   running without it.
    --json         Machine-readable output

**Output (--dry-run, default)**

```
Would remove 5 notes (oldest first):

  c7d400    EFT file processing           2026-01-10    distilled
  b2a100    MySQL index tuning            2026-01-15    distilled
  f9e300    SQS visibility timeout        2026-01-20    not distilled
  a4c200    ACH return handling           2026-01-25    not distilled
  d1b800    ECS task scaling              2026-02-01    distilled

Run without --dry-run to proceed.
```

**Output (default)**

```
Removed 5 notes (oldest first):

  c7d400    EFT file processing           2026-01-10    distilled
  b2a100    MySQL index tuning            2026-01-15    distilled
  f9e300    SQS visibility timeout        2026-01-20    not distilled
  a4c200    ACH return handling           2026-01-25    not distilled
  d1b800    ECS task scaling              2026-02-01    distilled
```

**Output (--json)**

```json
{
  "removed": 5,
  "ids": ["c7d400", "b2a100", "f9e300", "a4c200", "d1b800"]
}
```

**Note:** `--dry-run` is strongly recommended before running this command.
Removed notes are not recoverable. Undistilled notes in the removal
set represent knowledge that has not been preserved in any page document.

---

### kno admin page delete

```
kno admin page delete <id>
```

Permanently delete a page document. This operation is irreversible and
not exposed via MCP.

Associated notes are not deleted. Their `distilled_into` arrays are
updated to remove the deleted page id. Notes that referenced only this
page have `distilled_into` set back to null, making them eligible to be
distilled again on a future pass.

**Output (default)**

```
Deleted: AWS Infrastructure  [b81e44]
```

**Output (--json)**

```json
{
  "id": "b81e44",
  "deleted": true
}
```

---

### kno admin index rebuild

```
kno admin index rebuild
```

Rebuild the full-text search index from scratch by walking the vault
directory. Use when the index is suspected to be out of sync with vault
contents. Under normal operation the index is maintained automatically
on every write and this command should not be needed.

**Output (default)**

```
Rebuilding index...
Indexed 42 notes, 4 pages.
Done.
```

---

## CONFIGURATION

```toml
# ~/kno/config.toml

[notes]
max_count = 200                  # vault capacity; oldest distilled removed first
default_list_limit = 50          # default for kno note list
summary_max_tokens = 100         # hint to skill: target length for note summaries

[pages]
max_content_tokens = 8000        # soft cap on page document size; content
                                 # exceeding this limit is truncated with a
                                 # warning — never a hard failure

[distill]
max_notes_per_run = 50        # max notes processed in a single distill
                                 # pass. Set high enough to cover typical
                                 # backlogs in one run. If backlog exceeds
                                 # this limit, distill reports how many remain
                                 # and the skill can prompt a follow-up run.

[search]
default_limit = 5                # default result count for all search commands
```

Defaults are designed to be predictably successful. A skill operating within
default limits always knows the upper bound of what it will receive.

Key behaviors:
- Exceeding `notes.max_count` at create time removes the oldest distilled
  note. If none exist, the oldest note is removed regardless of status.
- Exceeding `pages.max_content_tokens` truncates with a warning — never
  a hard failure.
- Exceeding `distill.max_notes_per_run` processes up to the limit and
  reports how many notes remain for a follow-up run.
