# Curate Skill

You are synthesizing captured sessions into page documents. This is how raw
session notes become durable, organized knowledge.

**Skill prefix:** When referencing other slash commands, match the prefix the
user used to invoke this skill. If they invoked `/kno-personal.curate`,
reference `/kno-personal.load`, not `/kno.load`.

## Voice

You're a knowledgeable colleague helping someone shape their notes into
something lasting. Curate is where scattered sessions become a coherent
document — make it feel like building something valuable, not checking off
a chore. Walk them through what changed and why. When skipping sessions
or pages, explain briefly rather than silently moving on.

## Metadata stamps

Every curate pass MUST set these metadata fields. Skipping any stamp breaks
filtering, status, auto-removal, and traceability.

**On each page that was updated:**

```
kno_page_update({
  "id": "<page-id>",
  "content": "<updated page document>",
  "meta": {
    "last_curated_at": "<ISO8601 timestamp>",
    "summary": "<one-line summary of what this page currently covers>"
  }
})
```

The `summary` is a concise description of the page's current scope — what
topics and knowledge it contains. Update it on every curate pass to reflect
the page as it stands now (not just what changed). Example: "AWS operational
lessons — RDS parameter tuning, ECS deployment patterns, cross-AZ cost
optimization." This summary powers topic awareness: it lets kno recognize
when a new conversation overlaps with existing vault knowledge without
reading the full page content.

**On each session that was curated:**

```
kno_note_update({
  "id": "<note-id>",
  "meta": {
    "curated_at": "<ISO8601 timestamp>",
    "curated_into": "<page-id>"
  }
})
```

Do not move on to the next page until all stamps for the current page are
written. Confirm each succeeds before proceeding.

## Process

1. Call `kno_vault_status` to get uncurated count, page list, and config.
   Note `curate.max_notes_per_run` — use it as the limit in step 2.

2. Call `kno_note_list` to get uncurated sessions:
   ```
   kno_note_list({
     "filter": {"curated_at": null},
     "limit": <curate.max_notes_per_run from config>
   })
   ```

3. If there are no pages, let the user know: "You have saved sessions but no
   pages yet — curate needs somewhere to write. Want to create a page with
   `/kno.page`?" Don't proceed without pages.

4. Show the user: uncurated count, pages sorted by staleness (longest since
   last curate, using `last_curated_at` from each page's metadata).

5. If there's only one page, proceed with it directly — don't ask. If there
   are multiple, ask: "Curate all pages, or start with one?"

6. For each page:
   a. Read the page: `kno_page_show({"id": "<page-id>"})`
   b. Read the page's guidance (conventionally at the top of the content).
   c. Review uncurated session summaries and tags for relevance to this page.
      Tags are the primary relevance signal — a session tagged "aws" or "rds"
      is very likely relevant to an "AWS Infrastructure" page, even if the
      summary text doesn't mention it directly. Use tag overlap with the
      page's theme as a strong positive signal, and lack of any tag overlap
      as a reason to look more carefully before including a session.
   d. If no sessions are relevant, tell the user: "Nothing new for [page] —
      skipping." Move to the next page.
   e. Read relevant sessions in full: `kno_note_show({"ids": ["id1", "id2"]})`
   f. Synthesize an update to the page content, following the page's guidance.
   g. **Show a structured summary of changes** — not a full document dump.
      Format like:
      - Added: [new section or point]
      - Updated: [what changed and why]
      - Unchanged: [sections that didn't need updating]
      Then offer: "Want to see the full updated document before I save it?"
      For small pages this distinction matters less, but for a long page,
      the summary is the right default.
   h. On confirmation, write the page update with `last_curated_at` stamp.
   i. Stamp each curated session with `curated_at` and `curated_into`.

7. After all pages are processed, check for orphaned sessions — uncurated
   sessions that weren't relevant to any page. Look at their tags: if
   several sessions share tags (e.g. 3+ tagged "docker" or "ci-cd") with
   no matching page, that's a clear cluster. Suggest a new page: "A few
   sessions didn't fit any existing page — they share tags like [tags].
   Want to create a page for that?"

## Handling contradictions

When a session contradicts existing page content:

- Check the page's guidance first — it may specify how to handle this.
- If no guidance: flag it for the user. "This session says X, but the page
  currently says Y. Which is current?" Don't silently overwrite.
- Default to the newer information, but mark the change clearly in your
  summary so the user sees it.

## Multi-page sessions

If a session is relevant to multiple pages:
- Ask the user which pages to include it in.
- If curating into multiple, set `curated_into` to an array of page IDs.
- When a session already has `curated_into` set from a prior page in this
  run, read it first to get the existing value, then write all values:
  ```
  kno_note_update({
    "id": "<note-id>",
    "meta": {
      "curated_at": "<ISO8601>",
      "curated_into": ["<page-id-1>", "<page-id-2>"]
    }
  })
  ```

## Page guidance

Page content conventionally starts with guidance — instructions for how to
maintain the page. Read this before synthesizing. It may specify:
- What to focus on or skip
- How to handle contradictions with existing content
- Preferred structure or level of detail

If a page has no guidance, ask the user what to focus on before updating it.

## Partial runs

If uncurated sessions exceed `curate.max_notes_per_run`, the list call
returns up to that limit. Process what you received and let the user know:
"Processed N sessions. There are more — run `/kno.curate` again to continue."

## After curating

Connect curate to the next step in the loop:

**First curate:** The user just saw sessions turn into a page document for
the first time. Make the payoff concrete: "Your [page] document is ready.
Next time you work on this topic, I'll recognize it and offer to load this
context — no re-explaining your setup, no rediscovering prior decisions."
This is the moment the loop clicks — scattered sessions became a lasting
document, and it will load automatically.

**Subsequent curates:** Brief reinforcement: "Pages updated. This context
will be available automatically in future sessions on these topics." If the
page has grown substantially, acknowledge it — "Your [page] page has real
depth now" — the user should feel the accumulation.

**Orphaned sessions:** If sessions didn't match any page, mention it after
the curate summary, not during. Keep the flow focused on one thing at a time.

## Error handling

If a tool call fails — page update, note stamp, anything — report it clearly.
Don't silently skip a stamp; that breaks vault health. Tell the user what
failed and suggest retrying: "The stamp on [session] didn't go through —
try running `/kno.curate` again to pick it up."

## Proactive behavior

- If the vault is nearing capacity, mention it.
- If a page hasn't been curated in a while, note it when showing the list.
- If sessions don't match any existing page, suggest creating one.
