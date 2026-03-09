# Start Skill

The user typed `/kno.start` to connect to their vault. Check the vault
and get them working.

**Skill prefix:** When referencing other slash commands, match the prefix the
user used to invoke this skill. If they invoked `/kno-personal.start`,
reference `/kno-personal.curate`, not `/kno.curate`.

## Project vault

Call `kno_vault_status`. If `vault_type` is `"project"` and the response
includes `project.page`, load the bound page immediately with
`kno_page_show` — no need to ask.

Check whether the page has been curated (`last_curated_at` in page
metadata). This tells you whether the page has real content or is still
the guidance template from `kno init`.

**Page has curated content:**

> kno active — loaded **[page name]** (project vault).
> [1-2 sentence demonstration of understanding the page content]

**Page is fresh (no `last_curated_at`):**

> kno active — **[page name]** project vault is set up. Start working
> and I'll capture what matters — your first curate will bring this
> page to life.

If the bound page doesn't exist in the vault:

> kno active — config references page "[page name]" but it's not in your
> vault. Want to create it?

Mention uncurated notes if applicable. Do not list other pages.

## Git detected, no project vault

If `vault_status` includes `git` but `vault_type` is not `"project"`,
this is a git repo using the personal vault.

If `skill.prompt_project_setup` is `false`, skip the offer. Just note
the repo and fall through to the standard flow:

> kno active — detected: [repo_name].

Otherwise, suggest setting up a project vault. Keep it to one offer —
the value is a shared knowledge base the team can build on:

> kno active — detected: [repo_name]. Want to set up a project vault?
> Run `kno init` to create a shared knowledge base in this repo —
> it'll take effect next session.

On decline: offer once to disable future prompts. If yes, call
`kno_set_option(key: "prompt_project_setup", value: "false")`. Then
fall through to the standard flow with the personal vault.

## Standard flow

**Has pages:** List names, offer to load.

> Your vault has **AWS Infrastructure**, **CNC Maintenance**, and
> **Customer Onboarding**. Want me to load any of these?

**Has notes, no pages:** Suggest creating a page — notes without a page
can't surface automatically in future sessions. Keep it brief.

**Empty vault:**

> Your vault is empty. As we work, I'll spot decisions, insights, and
> solutions worth keeping. Just start working.

**Uncurated notes:** In any flow, append:

> You also have N uncurated notes — `/kno.curate` when you're ready.

## Rules

- 2-4 lines total. No tutorials, no command listings.
- Auto-load when bound. Do not ask.
- Do not explain the knowledge loop.
- If `skill.nudge_level` is `off`, append: "Suggestions are off — use
  slash commands for captures and loads."
- After the response, you're done.
