# Apideck CLI

Beautiful, agentic-friendly CLI for [Apideck Unified API](https://www.apideck.com).

OpenAPI is the source of truth — no SDK dependency, no code generation. Humans get a polished TUI. AI agents get token-efficient progressive disclosure.

> Read the [blog post](https://www.apideck.com/blog/mcp-server-eating-context-window-cli-alternative)

## Features

- **OpenAPI-native** — Parses the unified Apideck OpenAPI spec directly. No generated code, no SDK dependency.
- **Beautiful TUI explorer** — Interactive API playground in the terminal.
- **AI agent-ready** — `apideck agent-prompt` outputs a token-optimized prompt (~80 tokens vs ~3,600 for MCP).
- **Permission layers** — Auto-classified operations (read/write/dangerous) with built-in safety rails.
- **DX Ready** — Styled output, interactive wizards, spinners, semantic colors, adaptive themes.

## Install

### Homebrew

```bash
brew install apideck-libraries/tap/apideck
```

### Go

```bash
go install github.com/apideck-libraries/cli/cmd/apideck@latest
```

### Docker

```bash
docker run -e APIDECK_API_KEY=xxx apideck/cli accounting invoices list
```

### From source

```bash
git clone https://github.com/apideck-libraries/cli.git
cd cli
make build
./bin/apideck --help
```

## Quick Start

### 1. Configure authentication

```bash
# Option A: Environment variables
export APIDECK_API_KEY=your-api-key
export APIDECK_APP_ID=your-app-id
export APIDECK_CONSUMER_ID=your-consumer-id

# Option B: Interactive setup wizard
apideck auth setup
```

### 2. Start using the API

```bash
# List available API groups
apideck --list

# List invoices
apideck accounting invoices list

# Create a contact
apideck crm contacts create --name "Jane Doe" --email "jane@example.com"

# Get a specific invoice
apideck accounting invoices get --id inv_001

# Use a specific connector
apideck accounting invoices list --service-id xero
```

### 3. Explore interactively

```bash
# Launch the TUI explorer
apideck explore

# Scoped to a specific API
apideck accounting explore
```

## Usage

```bash
# Discovery
apideck --list                              # List available API groups
apideck accounting --list                   # List resources + operations
apideck accounting invoices --help          # Show operations + params

# API calls
apideck accounting invoices list --limit 10
apideck accounting invoices create --customer-id "cus_123" --total 150
apideck crm contacts list --output json

# Complex requests via raw JSON
apideck accounting invoices create --data '{"customer_id": "cus_123", "line_items": [...]}'
apideck accounting invoices create --data @invoice.json

# Management
apideck auth status                         # Show credential sources
apideck sync                                # Update to latest API spec
apideck info                                # Spec version, cache status
apideck history                             # Recent API calls timeline
apideck permissions                         # Show/edit permission config
```

### Output Formats

The CLI auto-detects your environment:

| Context | Default | Override with |
|---------|---------|---------------|
| Interactive terminal | `table` | `--output json\|yaml\|csv` |
| Piped / non-TTY | `json` | `--output table\|yaml\|csv` |

```bash
# Select specific fields
apideck accounting invoices list --fields id,customer,total

# Raw API response
apideck accounting invoices list --raw

# Quiet mode (no spinners, no status lines)
apideck accounting invoices list -q -o json
```

### Pagination

```bash
# Manual pagination
apideck accounting invoices list --cursor "next_page_token"

# Fetch all pages automatically
apideck accounting invoices list --all

# Limit total pages fetched
apideck accounting invoices list --all --max-pages 10
```

## Permission Engine

Operations are auto-classified based on HTTP method:

| HTTP Method | Level | Behavior |
|-------------|-------|----------|
| GET, HEAD, OPTIONS | `read` | Auto-approved |
| POST, PUT, PATCH | `write` | Confirmation prompt |
| DELETE | `dangerous` | Blocked unless overridden |

Override defaults in `~/.apideck-cli/permissions.yaml`:

```yaml
defaults:
  read: allow
  write: prompt
  dangerous: block

overrides:
  accounting.payments.create: dangerous   # upgrade: payments are sensitive
  crm.contacts.delete: write              # downgrade: allow with prompt
```

Use `--yes` to skip write prompts. Use `--force` to override dangerous blocks.

## AI Agent Integration

The Apideck CLI is designed to be the most efficient way for AI agents to interact with the Apideck API.

### Agent prompt (~80 tokens)

```bash
apideck agent-prompt
```

Outputs a token-optimized prompt that teaches any AI agent how to use the CLI. Compare this to loading a full OpenAPI spec (~50,000 tokens) or MCP tool schemas (~3,600 tokens per API).

### Progressive disclosure

Agents discover capabilities lazily — only loading what they need:

```bash
apideck --list                              # What APIs exist?
apideck accounting --list                   # What resources?
apideck accounting invoices create --help   # What parameters?
```

### Install as a Claude Code skill

```bash
apideck skill install
```

This writes a skill file to `~/.claude/skills/apideck.md` so Claude Code automatically knows how to use the Apideck CLI.

### Non-TTY behavior

When running in a non-interactive environment (CI/CD, agent frameworks):
- `read` operations execute normally
- `write` / `dangerous` operations fail with exit code 1 and structured JSON error
- Use `--yes` and `--force` flags to allow operations programmatically

## Why Not curl, OpenAPI, or MCP?

The obvious question: AI agents can already read OpenAPI specs and make HTTP requests with curl. Why build another tool?

Because **"can" ≠ "should"**. Agents *can* write raw assembly too — we still give them Python.

### Context Window Explosion

A typical OpenAPI spec is **5,000–50,000+ tokens**. An AI agent that loads the full spec to list customers just burned thousands of tokens on endpoints it will never use.

**Apideck CLI**: `apideck agent-prompt` → ~80 tokens. Progressive disclosure via `--list` and `--help` means the agent only loads what it needs, when it needs it.

### Auth Is a Nightmare

An AI agent using curl must parse security schemes, manage token storage and refresh, and construct auth headers correctly every time. One mistake = leaked credentials or broken auth.

**Apideck CLI**: Auth is configured once and injected automatically. The agent never sees or handles credentials directly.

### No Safety Rails

Give an AI agent curl and it can `DELETE /accounting/invoices/{id}` just as easily as `GET /accounting/invoices`. There's no distinction between reading data and destroying it.

**Apideck CLI**: Auto-classified permission layers. GET is auto-approved. POST/PUT prompts for confirmation. DELETE is blocked unless explicitly allowed. Safety is structural, not an afterthought.

### Discovery Is Wasteful

With curl + OpenAPI, every API call requires: load spec → parse → search → extract params → construct request → execute → parse response. That's 7 steps of repeated token-burning work.

**Apideck CLI**: `apideck accounting invoices list --limit 10`. One step.

### Why Not MCP?

| Dimension | curl + OpenAPI | MCP | **Apideck CLI** |
|-----------|---------------|-----|-----------------|
| Context cost | ~5,000–50,000 tokens | ~3,600 tokens/API | **~80 tokens** |
| Auth handling | Manual, per-request | Per-server, inconsistent | **Automatic** |
| Safety/Permissions | None | Optional, per-server | **Built-in** |
| Discovery | Load entire spec | Schema dump upfront | **Lazy, progressive** |
| Runtime deps | None | Server process | **None (static binary)** |
| Agent compatibility | Any (shell) | MCP clients only | **Any (shell)** |
| Custom code needed | Yes | Yes, per server | **None** |

The best AI agent interface isn't a new protocol — it's the oldest one we have: a well-designed command-line tool.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    apideck binary                     │
│                                                      │
│  ┌──────────┐  ┌──────────┐  ┌───────────────────┐  │
│  │  Parser   │  │  TUI     │  │  Agent Interface  │  │
│  │ OpenAPI   │  │ Explorer │  │  --help/--list    │  │
│  │ libopenapi│  │ bubbletea│  │  agent-prompt     │  │
│  └────┬─────┘  └────┬─────┘  └────────┬──────────┘  │
│       │              │                 │             │
│  ┌────▼──────────────▼─────────────────▼──────────┐  │
│  │              Command Router                     │  │
│  │  spec → operations → subcommands + flags        │  │
│  └────────────────────┬───────────────────────────┘  │
│                       │                              │
│  ┌────────────────────▼───────────────────────────┐  │
│  │           Permission Engine                     │  │
│  │  read / write (prompt) / dangerous (block)      │  │
│  └────────────────────┬───────────────────────────┘  │
│                       │                              │
│  ┌────────────────────▼───────────────────────────┐  │
│  │           Auth Manager                          │  │
│  │  env vars > config file > setup wizard          │  │
│  └────────────────────┬───────────────────────────┘  │
│                       │                              │
│  ┌────────────────────▼───────────────────────────┐  │
│  │           HTTP Client + Response Formatter      │  │
│  │  retry · rate-limit · JSON/table/YAML/CSV       │  │
│  └────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

## Tech Stack

| Component | Choice |
|---|---|
| Language | Go |
| CLI framework | [Cobra](https://github.com/spf13/cobra) |
| TUI framework | [Bubbletea](https://github.com/charmbracelet/bubbletea) + [Lipgloss](https://github.com/charmbracelet/lipgloss) |
| Interactive forms | [Huh](https://github.com/charmbracelet/huh) |
| OpenAPI parsing | [libopenapi](https://github.com/pb33f/libopenapi) |
| HTTP client | net/http + [retryablehttp](https://github.com/hashicorp/go-retryablehttp) |

## Development

```bash
# Build
make build

# Run
make run

# Test
make test

# Clean
make clean
```

## License

MIT
