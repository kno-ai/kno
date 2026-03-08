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

> Your vault has N notes but no pages yet. Just start working —
> `/kno.curate` will turn them into pages when you're ready.

### Empty vault

> Your vault is empty — just start working and kno will notice when
> something worth keeping comes up.

## Rules

- **Keep it short.** The entire response should be 2-4 lines. No tutorials,
  no explanations of how kno works, no command listings.
- **Do not load anything.** This skill checks the vault and offers — loading
  happens if the user says yes (handle it like an awareness-initiated load)
  or via `/kno.load`.
- **Do not explain the knowledge loop.** The user installed kno. They want
  to use it, not learn about it.
- **If `nudges.level` is `off`**, append: "Suggestions are off — use
  slash commands for captures and loads." This sets expectations so the
  user doesn't wonder why kno is quiet.
- **After the response, you're done.** kno takes over from here.
