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
    "summary": "<one-line summary of what this page currently covers>",
    "tags": ["tag1", "tag2", "..."],
    "note_count": "<number of notes curated into this page>"
  }
})
```

**`summary`** — a concise description of the page's current scope. Update
on every curate pass to reflect the page as it stands now (not just what
changed). Example: "AWS operational lessons — RDS parameter tuning, ECS
deployment patterns, cross-AZ cost optimization." Powers topic awareness.

**`tags`** — the union of all tags from notes curated into this page,
deduplicated and sorted alphabetically. Collect tags from each note being
curated, merge with any existing `tags` on the page, deduplicate. These
tags power publishing (frontmatter) and topic awareness.

**`note_count`** — total number of notes curated into this page (including
previously curated notes, not just this pass). Read the current value from
the page metadata and add the count of newly curated notes.

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

3. If there are no pages, offer to create one now. Look at common tags across
   sessions and suggest a page name: "You have N sessions but no pages yet.
   Based on your tags, a [suggested name] page would cover most of them —
   want to create it and curate in one go?" If yes, create the page with
   `kno_page_create` and proceed with curation. Don't bounce the user to a
   separate command.

4. Show the user: uncurated count, pages sorted by staleness (longest since
   last curate, using `last_curated_at` from each page's metadata).

5. If this curate was initiated from a post-capture offer (Tier 3), the
   target page is already established — skip step 4 and proceed directly
   with that page. Otherwise: if there's only one page, proceed with it
   directly — don't ask. If there are multiple, ask: "Curate all pages,
   or start with one?"

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
   f. Synthesize an update to the page content, following the page's guidance
      and the default voice below.
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

## Page voice

Pages should read like a helpful, efficient assistant left you notes —
direct, specific, and immediately useful. The target voice:

- **Direct and imperative.** "Pin parameter groups before minor version
  upgrades." Not "It is recommended that parameter groups be pinned."
- **Specific numbers and values.** "20 connections per instance, hard max."
  "Drain window: 60 seconds minimum." These concrete details are what make
  a page worth loading.
- **Context for why, not just what.** "The default 30s caused dropped
  requests during deploys." One sentence of context turns a rule into
  knowledge.
- **No filler.** No "In this section we'll cover..." No "It's important
  to note that..." Just the knowledge.

Page guidance can override this voice — if the user wants a different tone,
follow their guidance. This is the default when no guidance exists.

## Partial runs

If uncurated sessions exceed `curate.max_notes_per_run`, the list call
returns up to that limit. Process what you received and let the user know:
"Processed N sessions. There are more — ask kno to continue, or run
`/kno.curate` again."

## After curating

Connect curate to the next step in the loop:

**First curate:** The user just saw sessions turn into a page document for
the first time. Keep it concrete: "Your [page] document is ready — kno will pull this in next time you're
working on the topic."

**Subsequent curates:** Brief reinforcement: "Pages updated. This context
will be available automatically in future sessions on these topics." If the
page has grown substantially, acknowledge it — "Your [page] page has real
depth now" — the user should feel the accumulation.

**Published pages:** If the `kno_page_update` response includes
`published_to`, the page was automatically published to configured
targets. Mention it naturally: "Pages updated and published." Don't
explain what publishing does — just confirm it happened.

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
