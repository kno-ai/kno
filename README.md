# kno

Local-first knowledge capture for LLM conversations.

kno captures valuable LLM sessions into a local knowledge vault — so insights, decisions, and context compound over time instead of vanishing with each chat.

## The problem

You have great conversations with LLMs — debugging sessions, architecture decisions, deep dives. Then they're gone. Even saved transcripts are unsearchable walls of back-and-forth. Every new session starts cold.

## How kno works

kno is a single binary that provides:

- **CLI commands** for capturing content from clipboard, stdin, or files
- **MCP server** for zero-friction capture directly from Claude Desktop
- **Structured storage** — each capture is a directory with clean markdown and JSON metadata

### Capture via Claude Desktop

Type `/kno.capture` at the end of a conversation. Claude summarizes the session, extracts decisions and key points, and saves it to your vault automatically.

### Capture via CLI

```bash
pbpaste | kno capture --stdin --title "SQS debugging session"
```

### Browse captures

```bash
kno list
kno show <capture-name>
```

Or use `/kno.load` in Claude Desktop to browse and load a previous capture into a new conversation.

## Quick start

```bash
go install github.com/kno-ai/kno/cmd/kno@latest
kno setup
```

That's it. `kno setup` creates a vault at `~/kno` and registers with Claude Desktop if installed.

## Vision

kno is the foundation for a personal knowledge workflow:

1. **Capture** — save structured summaries from LLM sessions (today)
2. **Distill** — merge insights from captures into maintained topic documents (planned)
3. **Context** — load relevant knowledge into new sessions automatically (planned)
4. **Knowledge** — readable, maintained documents that represent your current understanding (planned)

The goal: your knowledge compounds across hundreds of conversations over years, and every new session starts with the right context.

## Project status

Early development. The capture layer is functional — CLI, MCP server, and skills work end-to-end. The distillation and context layers are planned.

## License

[MIT](LICENSE)
