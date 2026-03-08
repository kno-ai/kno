# Start Skill

The user typed `/kno` to connect to their vault. This is the zero-friction
entry point — check the vault and get them working.

**Skill prefix:** When referencing other slash commands, match the prefix the
user used to invoke this skill. If they invoked `/kno-personal`,
reference `/kno-personal.curate`, not `/kno.curate`.

## Process

1. Call `kno_vault_status` to get the vault snapshot.

2. Respond based on what you find:

### Vault has pages

List page names and offer to load. Names only — no summaries, no
descriptions. The user knows what their own pages are about.

> Your vault has **AWS Infrastructure**, **CNC Maintenance**, and
> **Customer Onboarding**. Want me to load any of these?
>
> Otherwise just start working — kno will suggest relevant context as we go.
> You can also ask kno to load something specific, or use `/kno.load`.

If there are also uncurated notes, add one line:

> You also have 5 uncurated notes — `/kno.curate` when you're ready.

### Vault has notes but no pages

If there are 3 or more notes, actively suggest page creation — notes
without pages can't be loaded automatically in future sessions:

> Your vault has N notes but no pages yet. Pages collect related sessions
> so kno can load them automatically — want to create one now?
> Otherwise just start working — `/kno.page` when you're ready.

If fewer than 3, keep it lighter:

> Your vault has N notes but no pages yet. Just start working —
> `/kno.page` or `/kno.curate` when you're ready.

### Empty vault

> Your vault is empty. As we work, I'll spot decisions, insights, and
> solutions worth keeping and offer to save them. Just start working.

## Developer context

When `vault_status` includes a `git` field, this is a project session.
Adjust the response to acknowledge the detected repo.

### Vault has pages matching the repo

> kno active — detected: [repo_name].
> Your vault has **[matching page]** and **[other pages]**. Want me to load any?

### No matching pages but vault has other pages

> kno active — detected: [repo_name]. No project page yet.
> Your vault has **[other pages]**. Want me to load any?
> Otherwise just start working — kno will capture as we go.

### Notes but no pages in git context

If there are 3 or more notes (especially if any are tagged with the repo name),
suggest a project page:

> kno active — detected: [repo_name]. You have N notes but no project page yet.
> Want to create one? It'll give future sessions on [repo_name] a head start.
> Otherwise just start working — `/kno.page` when you're ready.

### Empty vault in git context

> kno active — detected: [repo_name]. Vault is empty. As we work, I'll
> spot decisions, solutions, and project knowledge worth keeping. Just
> start working.

## Rules

- **Keep it short.** The entire response should be 2-4 lines. No tutorials,
  no explanations of how kno works, no command listings.
- **Do not load anything.** This skill checks the vault and offers — loading
  happens if the user says yes (handle it like an awareness-initiated load)
  or via `/kno.load`.
- **Do not explain the knowledge loop.** The user installed kno. They want
  to use it, not learn about it.
- **If `skill.nudge_level` is `off`**, append: "Suggestions are off — use
  slash commands for captures and loads." This sets expectations so the
  user doesn't wonder why kno is quiet.
- **After the response, you're done.** kno takes over from here.
