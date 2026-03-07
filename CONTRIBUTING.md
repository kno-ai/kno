# Contributing to kno

## Prerequisites

- Go 1.25+
- Claude Desktop (for MCP integration testing)

## Build and test

The Makefile is the single entry point:

```bash
make all        # check + unit tests + e2e tests (the default)
make build      # compile to /tmp/kno
make test       # go test ./...
make e2e        # build, then run test/e2e_test.sh
make fmt        # gofmt -w . (+ goimports if installed)
make check      # fmt-check + go vet (fails if code isn't formatted)
make lint       # check + staticcheck (if installed)
```

Or run the pieces directly:

```bash
go build -o /tmp/kno ./cmd/kno
go test ./...
./test/e2e_test.sh
```

## Run locally with a test vault

Create an isolated vault that won't touch your real data:

```bash
# Build
go build -o /tmp/kno ./cmd/kno

# Create a test vault (skips Claude Desktop registration)
/tmp/kno setup --vault /tmp/kno-test --no-claude-desktop

# Use it
echo "## TL;DR\n\nTest session content." | \
  /tmp/kno --vault /tmp/kno-test note create \
    --title "Test session" \
    --meta tags=testing \
    --meta summary="A test note"

/tmp/kno --vault /tmp/kno-test note list --json
/tmp/kno --vault /tmp/kno-test vault status
```

To start fresh, delete the directory:

```bash
rm -rf /tmp/kno-test
```

## Test with Claude Desktop

To test MCP integration against a dev build:

```bash
# Build and install to a known path
go build -o /tmp/kno ./cmd/kno

# Set up a test vault with MCP registration
/tmp/kno setup --vault /tmp/kno-dev --name kno-dev

# Restart Claude Desktop — /kno-dev.capture, /kno-dev.load, etc. will appear
```

The MCP registration points Claude Desktop at `/tmp/kno` with
`--vault /tmp/kno-dev`. Your real vault (if any) is untouched.

To remove the test registration, delete the `kno-dev` entry from your
Claude Desktop config (`~/Library/Application Support/Claude/claude_desktop_config.json`
on macOS).

## Run the MCP server manually

Useful for debugging MCP tool calls without Claude Desktop:

```bash
/tmp/kno --vault /tmp/kno-test mcp
```

This starts the MCP server over stdio. Send JSON-RPC messages on stdin.

## Project structure

```
cmd/kno/main.go          Entry point
internal/
  cli/                   Cobra commands (note, page, vault, setup)
  mcp/                   MCP server, tools, and prompts
  model/                 Data types: Note, Page, MetaMap
  vault/                 Vault interface
  vault/fs/              Filesystem vault implementation
  config/                Config loading and defaults (config.toml)
  search/                Bleve full-text search index
  skills/                Skill store interface
  skills/embedded/       Embedded skill markdown files
  app/                   App struct that wires vault + config + skills
  integration_test.go    Integration tests (Go API level)
test/
  e2e_test.sh            End-to-end CLI tests (builds binary, exercises every command)
Makefile                 Build, test, e2e, vet, lint
```

## Commit messages

We use [Conventional Commits](https://www.conventionalcommits.org/) to drive
automated releases. Prefix your commit message with a type:

| Prefix | Meaning | Version bump |
|---|---|---|
| `feat:` | New feature or capability | minor (0.2.0 → 0.3.0) |
| `fix:` | Bug fix | patch (0.2.0 → 0.2.1) |
| `docs:` | Documentation only | no release |
| `chore:` | Maintenance, CI, deps | no release |
| `refactor:` | Code change, no new behavior | no release |
| `test:` | Adding or updating tests | no release |
| `feat!:` or `BREAKING CHANGE:` | Breaking change | major (0.2.0 → 1.0.0) |

Examples:

```
feat: add note export command
fix: search index not created on fresh vaults
docs: update MCP setup instructions
chore: bump mcp-go dependency
```

## Release process

Releases are fully automated via [release-please](https://github.com/googleapis/release-please):

1. Merge PRs to `main` using conventional commit messages.
2. release-please opens (or updates) a "Release" PR with a changelog and
   version bump based on the commit types since the last release.
3. When you're ready to ship, merge the Release PR.
4. This creates a git tag → GoReleaser builds binaries → Homebrew tap is
   updated automatically.

You never need to manually create tags or GitHub releases.

**Version policy:** We follow [semver](https://semver.org/). While pre-1.0,
minor bumps may include breaking changes. After 1.0, breaking changes require
a major bump via `feat!:` or `BREAKING CHANGE:` in the commit footer.

## Language conventions

kno uses different language at different layers:

| Layer | Noun | Verb | Example |
|---|---|---|---|
| CLI / data model | note, page | create, list, show, update, search | `kno note create` |
| MCP tools | note, page | create, list, show, update, search | `kno_note_create` |
| MCP prompts / skills | session, page | capture, curate, load | `/kno.capture` |

- **CLI and MCP tools** use "note" — the data resource in the vault.
- **Skills and user-facing text** use "capture" (verb) and "session" (noun) — what the user experiences.
- These layers don't cross. CLI never says "capture". Skills never tell the user to run `kno note`.
