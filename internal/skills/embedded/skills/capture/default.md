# Capture Skill

You are capturing a conversation into the user's knowledge vault. Review the full conversation and produce a structured summary, then save it using the `kno_capture` tool.

## Process

1. Read through the entire conversation.
2. Identify the key outcomes: what was decided, what was learned, what changed.
3. Work backwards from those outcomes to capture *how* they were reached — the reasoning, tradeoffs, and alternatives considered.
4. Generate a short, descriptive title (e.g. "SQS retry strategy for failed messages", "React auth flow with refresh tokens"). The title should tell a future reader what this capture is about without opening it.
5. Write the structured markdown summary (see format below).
6. Call the `kno_capture` tool with `title` and `markdown`.

## Markdown format

Write the summary as a single markdown document with these sections, in this order. Include all sections even if some are brief.

```
## TL;DR

1-3 sentences summarizing what happened and the main outcome.

## Decisions

Each decision made during the session. For each:
- What was decided
- Why (the reasoning or constraint that drove it)
- What alternatives were considered and why they were rejected

This is the most important section. Be thorough here.

## Key points

Bullet list of important things learned, discovered, or confirmed.
Focus on facts and insights a future reader would need.

## Next steps

Concrete actions that follow from this session.
Only include real next steps, not vague intentions.

## Snippets

Code, commands, configurations, or examples worth preserving exactly.
Use fenced code blocks with language tags.
Omit this section if there are no meaningful snippets.
```

## Guidelines

- Write for a future reader who has no context about this conversation.
- Lead with outcomes, not process. "We decided X because Y" not "We discussed X and Y and Z."
- Preserve technical precision — exact names, versions, error messages, flag values.
- Capture the *why* behind decisions as carefully as the *what*.
- Omit conversational back-and-forth, pleasantries, and false starts.
- If the user provided a hint, use it to guide what you emphasize.
- Keep it concise. A good capture is 100-400 words, not a transcript.

## Calling the tool

After writing the summary, call the `kno_capture` tool:

- `title`: the short descriptive title you generated
- `markdown`: the full structured summary
- `meta`: include if the conversation has an obvious topic (e.g. `{"topic": "aws/sqs"}`)

Confirm to the user what was captured and the title you chose.
