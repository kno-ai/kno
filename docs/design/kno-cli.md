# kno CLI Reference

kno is a local-first knowledge vault for LLM conversations. The CLI provides
deterministic, testable CRUD operations against the vault. The CLI owns the
data — it has no knowledge of awareness, nudges, or skills. Those concerns
live in the MCP and skill layers above.

All commands support `--json` for machine-readable output. Human-readable
output is the default.

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
✗  Not found: note 20260305-rds-slow-query-debugging
✗  Not found: page aws-infrastructure
✗  --title is required
✗  --name is required
✗  --count is required
```

Commands that accept IDs fail immediately on the first unrecognized ID and
make no changes to the vault. Partial writes do not occur.

---

## SETUP

Setup is a one-time installation step and the first thing every user runs.

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

    --name <name>           MCP server name registered with Claude Desktop.
                            Default: kno. Use a distinct name for each
                            additional vault (e.g. kno-personal). The name
                            becomes the skill prefix in Claude Desktop:
                            /kno-personal.capture, /kno-personal.load, etc.

    --vault <path>          Vault directory path. Default: ~/kno for
                            the first vault. Use a distinct path for each
                            additional vault (e.g. ~/kno-personal).

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
  /kno.capture    — capture a session summary to your vault
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

1. Created at the end of a session
2. Available for load and curate while uncurated
3. Marked as curated after a curate pass (`curated_at`, `curated_into`)
4. Automatically removed when the vault reaches capacity (`notes.max_count`)

Auto-removal algorithm on `note create` when at capacity:
1. Remove the oldest note where `curated_at` is set (preferred — this
   knowledge is already preserved in a page document)
2. If no curated notes exist, remove the oldest note regardless of
   curate status. The response includes `auto_removed_uncurated: true` to
   signal that uncurated knowledge was lost.

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
                              Example: --filter curated_at=null
    --limit <n>               Maximum results to return.
                              Default: notes.default_list_limit (config)
    --json                    Machine-readable output

**Output (default)**

```
ID                                    TITLE                          CREATED       STATUS
20260305-rds-slow-query-debugging     RDS slow query debugging       2026-03-05    not curated
20260220-sqs-dead-letter-queue        SQS dead letter queue          2026-02-20    curated
20260110-spindle-bearing-diagnosis    Spindle bearing diagnosis      2026-01-10    curated
```

**Output (--json)**

```json
[
  {
    "id": "20260305-rds-slow-query-debugging",
    "title": "RDS slow query debugging",
    "metadata": {
      "tags": ["aws", "databases", "performance"],
      "summary": "Query planner regression after minor version upgrade. Fixed by pinning parameter group. Key lesson: always test minor upgrades in staging.",
      "curated_at": null,
      "curated_into": null
    },
    "created_at": "2026-03-05T14:22:00Z"
  }
]
```

Note: `summary` lives in `metadata` and is always included in JSON output
when present — it enables routing and load decisions without requiring
a `note show` call. Multi-value metadata
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
━━━ RDS slow query debugging  [20260305-rds-slow-query-debugging]  2026-03-05 ━━━━━━━━━━━━━━━━━━━━━━━━

Query planner regression after RDS 14.3 → 14.4 minor version upgrade...
[full note content]

tags: aws, databases, performance
curated_at: —
```

**Output (--json)**

```json
[
  {
    "id": "20260305-rds-slow-query-debugging",
    "title": "RDS slow query debugging",
    "content": "...",
    "metadata": {
      "tags": ["aws", "databases", "performance"],
      "summary": "Query planner regression after minor version upgrade...",
      "curated_at": null,
      "curated_into": null
    },
    "created_at": "2026-03-05T14:22:00Z"
  }
]
```

Note: Always returns an array, even for a single id.

---

### kno note create

```
kno note create --title <title>  [--meta <key>=<value>]...  < <content>
```

Create a new note. Content is read from stdin and is required. Title is
required. Metadata is optional. The vault index is updated automatically.

If the vault is at capacity (`notes.max_count`), the oldest curated
note is automatically removed before writing the new one. If no curated
notes exist, the oldest note (regardless of status) is removed instead.
The response indicates what was removed and whether it was uncurated.

**Options**

    --title <title>           Required. Display title for the note.
    --meta <key>=<value>      Optional metadata. Repeatable. Duplicate keys
                              produce an array — single key produces a scalar.
                              Common keys:
                                tags       one --meta per tag value
                                summary    short summary (auto-generated
                                           if not provided)

**Output (default)**

```
Created: RDS slow query debugging  [20260305-rds-slow-query-debugging]
```

When a curated note was auto-removed to make room:

```
Created: RDS slow query debugging  [20260305-rds-slow-query-debugging]
Removed: Spindle bearing diagnosis  [20260110-spindle-bearing-diagnosis]  (oldest curated)
```

When an uncurated note was auto-removed (no curated notes available):

```
Created: RDS slow query debugging  [20260305-rds-slow-query-debugging]
Removed: SQS visibility timeout    [20260120-sqs-visibility-timeout]  (oldest — UNCURATED, knowledge may be lost. Run /kno.curate)
```

**Output (--json)**

```json
{
  "id": "20260305-rds-slow-query-debugging",
  "title": "RDS slow query debugging",
  "created_at": "2026-03-05T14:22:00Z",
  "auto_removed": null
}
```

When a curated note was auto-removed:

```json
{
  "id": "20260305-rds-slow-query-debugging",
  "title": "RDS slow query debugging",
  "created_at": "2026-03-05T14:22:00Z",
  "auto_removed": "20260110-spindle-bearing-diagnosis"
}
```

When an uncurated note was auto-removed:

```json
{
  "id": "20260305-rds-slow-query-debugging",
  "title": "RDS slow query debugging",
  "created_at": "2026-03-05T14:22:00Z",
  "auto_removed": "20260120-sqs-visibility-timeout",
  "auto_removed_uncurated": true
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
                            Read before writing when appending to an
                            existing array (e.g. curated_into).

**Examples**

```bash
# stamp curation into one page
kno note update 20260305-rds-slow-query-debugging \
  --meta curated_at=2026-03-05T14:22:00Z \
  --meta curated_into=aws-infrastructure

# stamp curation into two pages — read first, then write all values
kno note update 20260305-rds-slow-query-debugging \
  --meta curated_at=2026-03-05T14:22:00Z \
  --meta curated_into=aws-infrastructure \
  --meta curated_into=kubernetes-migration

# update content only
echo "<revised content>" | kno note update 20260305-rds-slow-query-debugging

# update content and tags
echo "<revised content>" | kno note update 20260305-rds-slow-query-debugging \
  --meta tags=aws \
  --meta tags=rds
```

**Output (default)**

```
Updated: RDS slow query debugging  [20260305-rds-slow-query-debugging]
```

**Output (--json)**

```json
{
  "id": "20260305-rds-slow-query-debugging",
  "updated_at": "2026-03-05T15:00:00Z"
}
```

---

### kno note delete

```
kno note delete <id>  [--json]
```

Permanently delete a note. The search index is updated automatically.

**Output (default)**

```
Deleted: RDS slow query debugging  [20260305-rds-slow-query-debugging]
```

**Output (--json)**

```json
{
  "id": "20260305-rds-slow-query-debugging",
  "title": "RDS slow query debugging",
  "deleted": true
}
```

---

### kno note prune

```
kno note prune --count <n>  [--dry-run]  [--json]
```

Remove the N oldest notes regardless of curate status. Use for bulk
cleanup when you want to reduce vault size beyond what auto-removal
handles. `--dry-run` shows what would be removed without deleting.

**Options**

    --count <n>    Required. Number of notes to remove.
    --dry-run      Preview removals without deleting.
    --json         Machine-readable output

**Output (--dry-run, default)**

```
Would remove 5 notes (oldest first):

  20260110-spindle-bearing-diagnosis   Spindle bearing diagnosis   2026-01-10    curated
  20260115-mysql-index-tuning         MySQL index tuning          2026-01-15    curated
  20260120-sqs-visibility-timeout     SQS visibility timeout      2026-01-20    not curated

Run without --dry-run to proceed.
```

**Output (--json)**

```json
{
  "removed": 5,
  "ids": ["20260110-spindle-bearing-diagnosis", "20260115-mysql-index-tuning", "20260120-sqs-visibility-timeout"]
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
ID                                          TITLE                            SCORE   STATUS
20260305-rds-slow-query-debugging           RDS slow query debugging         0.92    not curated
20260215-aurora-connection-pool-tuning      Aurora connection pool tuning     0.81    curated
```

**Output (--json)**

```json
[
  {
    "id": "20260305-rds-slow-query-debugging",
    "title": "RDS slow query debugging",
    "score": 0.92,
    "metadata": {
      "tags": ["aws", "databases", "performance"],
      "summary": "Query planner regression after minor version upgrade...",
      "curated_at": null,
      "curated_into": null
    },
    "created_at": "2026-03-05T14:22:00Z"
  }
]
```

Note: `summary` in `metadata` is always included in JSON output when present
enabling load decisions without follow-up `show` calls.

---

## PAGES

Pages are curated, living knowledge documents. They are durable — notes
are ephemeral, pages are not. Page content is stored and returned without
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
ID                    NAME                    LAST CURATED
aws-infrastructure       AWS Infrastructure        2026-03-01
cnc-machine-maintenance  CNC Machine Maintenance   2026-02-15
customer-onboarding      Customer Onboarding       —
```

**Output (--json)**

```json
[
  {
    "id": "aws-infrastructure",
    "name": "AWS Infrastructure",
    "metadata": {
      "last_curated_at": "2026-03-01T10:00:00Z"
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
━━━ AWS Infrastructure  [aws-infrastructure]  last curated 2026-03-01 ━━━━━━━━━━━━

## AWS Infrastructure — Current Understanding
...
[full page content]
```

**Output (--json)**

```json
{
  "id": "aws-infrastructure",
  "name": "AWS Infrastructure",
  "content": "## AWS Infrastructure — Current Understanding\n...",
  "metadata": {
    "last_curated_at": "2026-03-01T10:00:00Z"
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
provided the page starts empty and is populated on the first curate pass.

Content is stored and returned without interpretation.

**Options**

    --name <name>           Required. Display name for the page.
    --meta <key>=<value>    Optional metadata. Repeatable.

**Output (default)**

```
Created: AWS Infrastructure  [aws-infrastructure]
```

**Output (--json)**

```json
{
  "id": "aws-infrastructure",
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
# update content — primary curate write-back
echo "<updated content>" | kno page update aws-infrastructure

# update metadata only
kno page update aws-infrastructure --meta last_curated_at=2026-03-05T14:22:00Z

# update both — typical curate write-back
echo "<updated content>" | kno page update aws-infrastructure \
  --meta last_curated_at=2026-03-05T14:22:00Z
```

**Output (default)**

```
Updated: AWS Infrastructure  [aws-infrastructure]
```

**Output (--json)**

```json
{
  "id": "aws-infrastructure",
  "updated_at": "2026-03-05T14:22:00Z"
}
```

---

### kno page rename

```
kno page rename <id>  --name <name>  [--json]
```

Rename a page. Renames the underlying files, updates the search index, and
fixes `curated_into` references on any notes that pointed to the old ID.
The page ID is derived from the name (slugified), so renaming typically
changes the ID.

**Options**

    --name <name>    Required. New display name for the page.
    --json           Machine-readable output

**Output (default)**

```
Renamed: aws-infrastructure → AWS Cloud Ops  [aws-cloud-ops]
```

**Output (--json)**

```json
{
  "old_id": "aws-infrastructure",
  "new_id": "aws-cloud-ops",
  "name": "AWS Cloud Ops"
}
```

If the slug doesn't change (e.g. only capitalization differs), the name
is updated in metadata but files are not renamed.

---

### kno page delete

```
kno page delete <id>  [--json]
```

Permanently delete a page document. The search index is updated.
Associated notes are not deleted — their `curated_into` arrays are
updated to remove the deleted page id. Notes that referenced only this
page have `curated_into` and `curated_at` cleared, making them eligible
to be curated again.

**Output (default)**

```
Deleted: AWS Infrastructure  [aws-infrastructure]
```

**Output (--json)**

```json
{
  "id": "aws-infrastructure",
  "name": "AWS Infrastructure",
  "deleted": true
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
ID                    NAME                  SCORE
aws-infrastructure       AWS Infrastructure        0.95
cnc-machine-maintenance  CNC Machine Maintenance   0.42
```

**Output (--json)**

```json
[
  {
    "id": "aws-infrastructure",
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

Return a snapshot of vault health, capacity, and configuration.

**Output (default)**

```
Vault: ~/kno

Notes: 143 / 500  (357 remaining)
  Curated:      121
  Uncurated:      22

Pages: 4
  aws-infrastructure       AWS Infrastructure        last curated 3 days ago
  cnc-machine-maintenance  CNC Machine Maintenance   last curated 2 weeks ago
  customer-onboarding      Customer Onboarding       last curated 1 month ago
  mysql-optimization       MySQL Optimization        never curated

Config:
  notes.max_count           500
  notes.default_list_limit   50
  notes.summary_max_tokens  100
  notes.max_content_tokens  3000
  pages.max_content_tokens   12000
  curate.max_notes_per_run  50
  search.default_limit          10
```

**Output (--json)**

```json
{
  "vault_path": "/Users/kevin/kno",
  "notes": {
    "total": 143,
    "max_count": 500,
    "remaining": 357,
    "curated": 121,
    "uncurated": 22
  },
  "pages": [
    {
      "id": "aws-infrastructure",
      "name": "AWS Infrastructure",
      "metadata": {
        "last_curated_at": "2026-03-02T10:00:00Z"
      }
    },
    {
      "id": "cnc-machine-maintenance",
      "name": "CNC Machine Maintenance",
      "metadata": {
        "last_curated_at": "2026-02-19T10:00:00Z"
      }
    }
  ],
  "config": {
    "notes": {
      "max_count": 500,
      "default_list_limit": 50,
      "summary_max_tokens": 100,
      "max_content_tokens": 3000
    },
    "pages": {
      "max_content_tokens": 12000
    },
    "curate": {
      "max_notes_per_run": 50
    },
    "search": {
      "default_limit": 10
    }
  }
}
```

Note: Config values are included in JSON output for programmatic access.
`vault_path` is always the fully resolved absolute path — never a
tilde-expanded shorthand.

---

### kno vault rebuild-index

```
kno vault rebuild-index  [--json]
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
max_count = 500                  # vault capacity; oldest curated removed first
default_list_limit = 50          # default for kno note list
summary_max_tokens = 100         # hint to skill: target summary length
max_content_tokens = 3000        # hard limit on note content (~12KB)

[pages]
max_content_tokens = 12000        # hard limit on page content (~48KB)

[curate]
max_notes_per_run = 50        # max notes processed in a single curate pass

[search]
default_limit = 10                # default result count for all search commands

[nudges]
level = "light"                   # "off" | "light" | "active"
                                  # off: no awareness, slash commands only
                                  # light: high-signal nudges only (default)
                                  # active: broader nudging
```

Defaults are designed to be predictably successful. Any consumer operating
within default limits always knows the upper bound of what it will receive.

Key behaviors:
- Exceeding `notes.max_count` at create time removes the oldest curated
  note. If none exist, the oldest note is removed regardless of status.
- Exceeding `notes.max_content_tokens` or `pages.max_content_tokens`
  rejects the write with an error. Content limits are enforced on both
  create and update.
- Exceeding `curate.max_notes_per_run` processes up to the limit and
  reports how many notes remain for a follow-up run.
