# Start Skill

The user typed `/kno.start` to connect to their vault. Check the vault
and get them working.

**Skill prefix:** When referencing other slash commands, match the prefix the
user used to invoke this skill. If they invoked `/kno-personal.start`,
reference `/kno-personal.curate`, not `/kno.curate`.

## Auto-load: project page bound

Call `kno_vault_status`. If the response includes `project.page`, a `.kno`
file binds a page to this directory. Load it immediately with
`kno_page_show` — no need to ask.

> kno active — loaded **[page name]**.
> [1-2 sentence demonstration of understanding the page content]

If the page doesn't exist in the vault:

> kno active — `.kno` references page "[page name]" but it's not in your
> vault. Want to create it?

Mention uncurated notes if applicable. Do not list other pages.

## Git detected, no project binding

If `vault_status` includes `git` but no `project`, this is a git repo
without a `.kno` file.

If `skill.prompt_project_setup` is `false`, skip the offer. Just add the
repo line and fall through to the standard flow:

> kno active — detected: [repo_name].

Otherwise, offer to bind a page:

- **Page matches repo name:** suggest binding it.
  > kno active — detected: [repo_name]. Want to bind **[page]** so it
  > loads automatically when you work here?

- **Pages exist but none match:** offer to create and bind.
  > kno active — detected: [repo_name]. No project page yet. Want to
  > create one and bind it for auto-load?

- **No pages:** mention the repo, fall through to standard flow.
  > kno active — detected: [repo_name].

On confirm: call `kno_set_option(key: "page", value: "[page name]")`.
Create the page first if needed. Mention once: "Saved to `.kno`. Commit
it to share with your team, or add it to `.gitignore` to keep it personal."

On decline: offer once to disable future prompts. If yes, call
`kno_set_option(key: "prompt_project_setup", value: "false")`.

## Standard flow

**Has pages:** List names, offer to load.

> Your vault has **AWS Infrastructure**, **CNC Maintenance**, and
> **Customer Onboarding**. Want me to load any of these?

**Has notes, no pages:** Suggest creating a page — notes without a page
can't be loaded automatically in future sessions. Keep it brief.

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
