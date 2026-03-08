This is a software project page for {{repo_name}}.

CAPTURE PRIORITIES:
Focus on decisions and their rationale, known issues and their status,
non-obvious setup and environment details, and solutions to hard problems.
The WHY behind technical choices matters more than the WHAT.

SKIP:
Implementation details that belong in code comments. Anything already
in the README. Things obvious from reading the code. Transient debugging
notes that didn't lead to durable knowledge.

PAGE STRUCTURE:
Organize into sections that reflect what's actually present:
- Decisions (chronological, with dates — dates matter as context ages)
- Known Issues (open items first, resolved with resolution noted)
- Setup & Environment (commands, config, non-obvious steps)
- Solved Problems (hard bugs and their solution paths)

Keep sections scannable — this is reference material, not narrative.
Use fenced code blocks with language tags for all code.
Group commands and environment variables in their own sections.

LIFECYCLE:
Decisions age — preserve dates so staleness is visible.
Debt items should clearly show current status.
Setup notes may become stale after dependency or infrastructure changes —
note the context they were written in (e.g., "before the Docker migration").

CROSS-REFERENCES:
If content references another service or shared library, name it explicitly.
Awareness uses these references to surface this page in related sessions.
