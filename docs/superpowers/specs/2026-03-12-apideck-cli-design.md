# Apideck CLI — Design Document

**Date**: 2026-03-12
**Status**: Draft
**Project**: apideck/cli — Beautiful, agent-friendly CLI for the Apideck Unified API

## Vision

A standalone CLI that turns the Apideck Unified API into a beautiful, secure, AI-agent-friendly command-line experience. OpenAPI is the source of truth — no SDK dependency, no code generation. Humans get a polished TUI with Speakeasy-grade DX. AI agents get token-efficient progressive disclosure.

## Core Differentiators

1. **OpenAPI-native** — Parses the unified Apideck OpenAPI spec directly. No sdk-go dependency, no Speakeasy-generated code.
2. **Beautiful TUI explorer** — Interactive API playground in the terminal (bubbletea/lipgloss).
3. **Permission layers** — Auto-classified operations (read/write/dangerous) solve AI agent security.
4. **Self-describing for AI** — `apideck agent-prompt` outputs a token-optimized prompt (~80 tokens vs ~3,600 for MCP).
5. **Speakeasy-grade DX** — Styled output, interactive wizards, spinners, semantic colors, adaptive themes.

## Target Users

- **Apideck customers**: Interact with the Unified API from the terminal.
- **AI agent developers**: Point any agent at Apideck via CLI instead of MCP.
- The CLI is the shared interface between both personas.

## Scope — v1

- OpenAPI parser + cached parse tree
- Dynamic command router (spec → Cobra subcommands)
- Auth manager (env vars > config file, interactive setup wizard)
- Permission engine (read/write/dangerous auto-classification)
- HTTP client with retry, rate-limit handling
- Output formatter (JSON, table, YAML, CSV with TTY detection)
- TUI explorer (bubbletea three-panel design)
- Agent interface (agent-prompt, --list, --help progressive disclosure)
- Distribution: Homebrew + GitHub Releases + Docker
- Claude Code skill generation

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
│  │  read (auto) / write (prompt) / dangerous (block)│ │
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
│                                                      │
│  ┌────────────────────────────────────────────────┐  │
│  │           Spec Cache (unified spec + parse tree)│  │
│  │  ~/.apideck-cli/cache/                          │  │
│  └────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

## Project Structure

```
apideck-cli/
├── cmd/
│   └── apideck/
│       └── main.go                 # Entry point
├── internal/
│   ├── spec/
│   │   ├── parser.go               # libopenapi wrapper, spec → internal model
│   │   ├── model.go                # APISpec, Operation, Parameter types
│   │   ├── cache.go                # TTL-based spec caching + serialized parse tree
│   │   └── embed.go                # go:embed baseline spec
│   ├── router/
│   │   └── router.go               # spec operations → Cobra subcommands + flags
│   ├── permission/
│   │   ├── engine.go               # Classify operations (read/write/dangerous)
│   │   └── config.go               # Load/save ~/.apideck-cli/permissions.yaml
│   ├── auth/
│   │   ├── manager.go              # Resolve credentials (env > config file)
│   │   ├── config.go               # ~/.apideck-cli/config.yaml read/write
│   │   └── setup.go                # Interactive `apideck auth setup` wizard (huh)
│   ├── http/
│   │   ├── client.go               # net/http + retryablehttp, rate-limit handling
│   │   └── response.go             # Status code interpretation, error normalization
│   ├── output/
│   │   ├── formatter.go            # TTY detection, format dispatch
│   │   ├── json.go                 # JSON output
│   │   ├── table.go                # Table output (lipgloss styled)
│   │   ├── yaml.go                 # YAML output
│   │   └── csv.go                  # CSV output
│   ├── tui/
│   │   ├── explorer.go             # Bubbletea main model
│   │   ├── endpoint_list.go        # Left panel: tag-grouped endpoints
│   │   ├── detail_panel.go         # Right panel: params, permissions, response
│   │   └── styles.go               # Lipgloss styles, brand colors, adaptive theme
│   ├── agent/
│   │   └── prompt.go               # Generate token-optimized agent prompt
│   └── ui/
│       ├── styles.go               # Shared brand colors, semantic colors, typography
│       ├── spinner.go              # Styled spinner component
│       ├── confirm.go              # Styled confirmation prompts
│       └── error.go                # Styled error/success/warning boxes
├── specs/
│   └── speakeasy-spec.yml          # Embedded baseline unified spec
├── go.mod
├── go.sum
├── Dockerfile
├── .goreleaser.yml
└── Makefile
```

## CLI Interface

```bash
# Discovery
apideck --list                              # List available API groups

# Usage (mirrors the API structure)
apideck accounting invoices list
apideck accounting invoices create --customer-id "cus_123" --total 150
apideck accounting invoices get --id inv_001
apideck crm contacts list --limit 10

# Exploration
apideck accounting --list                   # List all resources + operations
apideck accounting invoices --help          # Show operations + params (styled)
apideck accounting explore                  # Launch TUI explorer
apideck explore                             # Launch TUI explorer (all APIs)

# AI Agent interface
apideck agent-prompt                        # ~80 token global prompt
apideck accounting agent-prompt             # ~60 token scoped prompt
apideck accounting invoices create --help   # ~120 tokens, structured

# Auth
apideck auth setup                          # Interactive credential wizard
apideck auth status                         # Show credential sources

# Management
apideck info                                # Spec version, cache status
apideck sync                                # Force re-fetch + re-parse spec
# apideck sync v10.22.0                     # Version pinning (deferred, needs versioned URLs)
apideck history                             # Recent API calls timeline
apideck permissions                         # Show/edit permission config
```

## Internal Model

```go
type APISpec struct {
    Name        string                // "apideck"
    Version     string                // "10.24.3"
    BaseURL     string                // "https://unify.apideck.com"
    Description string
    APIGroups   map[string]*APIGroup  // "accounting", "crm", etc.
    Auth        AuthScheme
}

type APIGroup struct {
    Name        string                // "accounting"
    Description string
    Resources   map[string]*Resource  // "invoices", "bank-accounts"
}

type Resource struct {
    Name        string
    Description string
    Operations  []*Operation
}

type Operation struct {
    ID          string               // "accountingInvoicesList"
    Method      string               // "GET", "POST", etc.
    Path        string               // "/accounting/invoices"
    Summary     string
    Description string
    Parameters  []*Parameter
    RequestBody *RequestBody
    Responses   map[int]*Response
    Permission  PermissionLevel      // derived from Method
}

type Parameter struct {
    Name        string
    In          string               // "query", "path", "header"
    Type        string               // "string", "integer", "boolean"
    Required    bool
    Default     any
    Description string
    Enum        []string
}

type RequestBody struct {
    ContentType string
    Fields      []*BodyField         // flattened from JSON schema
    Required    bool
}

type BodyField struct {
    Name        string
    Type        string               // "string", "integer", "boolean", "object", "array"
    Required    bool
    Default     any
    Description string
    Enum        []string
    Items       *BodyField           // for array types
    Children    []*BodyField         // for nested objects (max depth 3)
    Format      string               // "date", "date-time", "email", etc.
}
```

### Schema Flattening (request bodies → CLI flags)

- Top-level fields → `--field-name`
- Nested objects → `--parent.child` dot notation (max depth: 3 levels)
- Arrays → `--items '[{"key": "val"}]'` JSON string input
- Simple arrays → `--tags val1,val2` comma-separated
- `oneOf`/`anyOf` schemas → fall back to `--data` raw JSON input
- `pass_through` / `additionalProperties` → `--pass-through '{"key": "val"}'` JSON string

**Raw body escape hatch** for complex requests:
```bash
# Instead of many flags, pass a JSON body directly
apideck accounting invoices create --data '{"customer_id": "cus_123", "line_items": [...]}'
apideck accounting invoices create --data @invoice.json  # from file
```

When `--data` is provided, it takes precedence over individual field flags.

## Command Router

Dynamic Cobra command generation from the parsed spec:

```
apideck                              # root
├── <api-group>                      # accounting, crm, hris, etc.
│   ├── <resource>                   # invoices, customers, etc.
│   │   ├── list                     # GET collection
│   │   ├── get                      # GET by ID
│   │   ├── create                   # POST
│   │   ├── update                   # PATCH/PUT
│   │   └── delete                   # DELETE
│   ├── --list                       # list all resources
│   ├── --help                       # styled help
│   ├── explore                      # launch TUI
│   └── agent-prompt                 # token-optimized prompt
├── auth
│   ├── setup                        # interactive wizard
│   └── status                       # show credential sources
├── sync                             # re-fetch + re-parse spec
├── info                             # spec version, cache status
├── history                          # recent API call timeline
├── permissions                      # show/edit permission config
├── explore                          # TUI for all APIs
├── agent-prompt                     # global agent prompt
├── skill
│   └── install                      # write Claude Code skill to ~/.claude/skills/
└── completion [bash|zsh|fish]       # generate shell completions
```

**Lazy loading:** Only the requested API group's resources are built into Cobra commands. `apideck accounting invoices list` doesn't build commands for CRM or HRIS.

**Verb mapping from HTTP methods:**

| HTTP Method | CLI Verb |
|-------------|----------|
| GET (collection) | `list` |
| GET (by ID) | `get` |
| POST | `create` |
| PATCH / PUT | `update` |
| DELETE | `delete` |

## Permission Engine

**Auto-classification from HTTP methods:**

| HTTP Method | Level | Behavior |
|-------------|-------|----------|
| GET, HEAD, OPTIONS | `read` | Auto-approved, silent execution |
| POST, PUT, PATCH | `write` | Styled confirmation prompt |
| DELETE | `dangerous` | Blocked unless overridden |

**Config file** at `~/.apideck-cli/permissions.yaml`:

```yaml
defaults:
  read: allow
  write: prompt
  dangerous: block

overrides:
  accounting.payments.create: dangerous   # upgrade: payments are sensitive
  crm.contacts.delete: write              # downgrade: allow with prompt
```

**Runtime behavior:**

- `read` → execute immediately
- `write` → styled `huh` confirmation prompt: "This will create a new invoice. Proceed?"
- `dangerous` → styled error box with instructions to override

**Agent-friendly behavior (non-TTY):**

- `read` → executes
- `write` / `dangerous` → fails with exit code 1 and structured JSON error
- `--yes` flag → skips write prompts
- `--force` flag → overrides dangerous blocks

## Auth Manager

**Credential resolution chain:**

```
Environment variables → Config file → Error with setup instructions
```

**Environment variables:**
```bash
APIDECK_API_KEY=xxx
APIDECK_APP_ID=yyy
APIDECK_CONSUMER_ID=zzz
APIDECK_SERVICE_ID=quickbooks    # optional, targets a specific connector
```

**Config file** at `~/.apideck-cli/config.yaml`:
```yaml
api_key: "xxx"
app_id: "yyy"
consumer_id: "zzz"
service_id: "quickbooks"         # optional default connector
```

**`apideck auth setup`** — interactive wizard via `huh`:
- Step-by-step form: API Key, App ID, Consumer ID, Service ID (optional)
- Saves to config file
- Makes a test API call to verify credentials
- Shows styled success/failure

**`apideck auth status`** — shows where each credential is sourced:
```
✓ API Key:       from env (APIDECK_API_KEY)
✓ App ID:        from config (~/.apideck-cli/config.yaml)
✓ Consumer ID:   from config (~/.apideck-cli/config.yaml)
✓ Service ID:    from config (quickbooks)
```

**Global `--service-id` flag** overrides config/env per-request:
```bash
apideck accounting invoices list --service-id xero
```

**Header injection per request (exact names from OpenAPI spec):**
```
Authorization: Bearer <api_key>
x-apideck-app-id: <app_id>
x-apideck-consumer-id: <consumer_id>
x-apideck-service-id: <service_id>    # optional, omitted if not set
```

## HTTP Client

**Stack:** `net/http` wrapped with `hashicorp/go-retryablehttp`

**Timeouts:**
- Default request timeout: 30 seconds
- Configurable via `--timeout` flag (in seconds)

**Retry policy:**
- Retry on: 429, 500, 502, 503, 504
- Max retries: 3
- Backoff: exponential with jitter
- Respect `Retry-After` header on 429s

**Request lifecycle:**
1. Auth manager injects credential headers
2. Build URL: baseURL + path + path params
3. Query params from flags → URL query string
4. Request body from flags → JSON body (POST/PUT/PATCH)
5. Execute with retry
6. Response → normalize into `APIResponse`

**Response model:**

```go
type APIResponse struct {
    StatusCode int
    Success    bool
    Data       any
    Error      *APIError
    Meta       *ResponseMeta
}

type APIError struct {
    Message    string
    Detail     string
    StatusCode int
    RequestID  string
}

type ResponseMeta struct {
    Cursor     string
    HasMore    bool
    RateLimit  int
    RatePeriod int
}
```

**Pagination:** Apideck uses cursor-based pagination (`meta.cursors.next` in response body).
- Default: returns first page only
- `--cursor <value>` — manual pagination to a specific page
- `--all` — fetches all pages automatically, streams results
- `--max-pages <n>` — safeguard with `--all` (default: 50)
- `--limit` — per-page limit, not total limit

**Request history:** Logged to `~/.apideck-cli/history.json` (FIFO, last 100 calls).

History entry schema:
```json
{"timestamp": "...", "method": "GET", "path": "/accounting/invoices", "status": 200, "duration_ms": 142, "service_id": "quickbooks"}
```

- Request/response bodies are NOT stored (security: may contain sensitive data)
- File writes use atomic replacement (write to temp, then rename) to avoid corruption from concurrent CLI invocations
- `apideck history` renders as a styled timeline table

## Output Formatter

**TTY detection drives defaults:**

| Context | Default format | Override with |
|---------|---------------|---------------|
| Interactive terminal | `table` | `--output json\|yaml\|csv` |
| Piped / non-TTY | `json` | `--output table\|yaml\|csv` |

**Table output** — lipgloss styled with colored headers, borders, status icons.

**Global flags:**
- `--output, -o` — `json|table|yaml|csv`
- `--fields` — select specific fields: `--fields id,customer,total`
- `--raw` — raw API response body, no normalization
- `--quiet, -q` — suppress spinners, status lines (agent-friendly)
- `--service-id` — target a specific connector (overrides config/env)
- `--version` — print CLI version and loaded spec version

**Agent usage:** `-q -o json` for clean, parseable output with zero visual noise.

## TUI Explorer

**Launch:** `apideck explore` (all APIs) or `apideck accounting explore` (scoped)

**Two-panel layout** (resource list + detail/response panel):

```
┌─ apideck explore · Accounting API v10.24.3 ──────────────────────────┐
│                                                                       │
│  RESOURCES          │  GET /accounting/invoices                        │
│  ──────────         │  ─────────────────────────                      │
│  > Invoices         │  List all invoices                              │
│    Bank Accounts    │                                                  │
│    Bills            │  PARAMETERS                                      │
│    Credit Notes     │  ┌───────────┬─────────┬──────────┬───────────┐ │
│    Customers        │  │ Name      │ Type    │ Required │ Default   │ │
│    Ledger Accounts  │  ├───────────┼─────────┼──────────┼───────────┤ │
│    Payments         │  │ limit     │ integer │ no       │ 20        │ │
│    Suppliers        │  │ cursor    │ string  │ no       │ —         │ │
│    Tax Rates        │  │ filter    │ string  │ no       │ —         │ │
│                     │  └───────────┴─────────┴──────────┴───────────┘ │
│                     │                                                  │
│                     │  Permission: read (auto-approved)                │
│                     │                                                  │
│  ──────────         │  [Try it]  [Copy curl]  [Copy CLI]              │
│  / fuzzy search     │                                                  │
│  ? help             │  RESPONSE                                        │
│                     │  { "status_code": 200, "data": [...] }          │
└─────────────────────┴────────────────────────────────────────────────┘
```

**Navigation model:**
- Left panel shows resources grouped by API. Selecting a resource expands its operations (list, get, create, update, delete) as a sub-list.
- Right panel shows details for the currently highlighted operation.

**Key interactions:**
- `↑/↓` — navigate resources; within an expanded resource, navigate operations
- `←/→` — switch panels
- `Enter` on resource — expand/collapse operation sub-list
- `/` — fuzzy search across all endpoints
- `Enter` on [Try it] — execute request live, show response
- `c` — copy as curl
- `C` — copy as CLI command
- `?` — keybinding help overlay
- `q` / `Esc` — quit

**[Try it] flow:** Permission check → spinner → execute → render response with syntax highlighting → log to history.

## Agent Interface

**`apideck agent-prompt`** (~80 tokens):
```
Use `apideck` to interact with the Apideck Unified API.
Available APIs: `apideck --list`
List resources: `apideck <api> --list`
Operation help: `apideck <api> <resource> <verb> --help`
APIs: accounting, ats, connector, crm, ecommerce, file-storage, hris, issue-tracking, sms, vault, webhook, proxy
Auth is pre-configured. GET auto-approved. POST/PUT/PATCH prompt (use --yes). DELETE blocked (use --force).
Use --service-id <connector> to target a specific integration.
For clean output: -q -o json
```

**`apideck accounting agent-prompt`** (~60 tokens):
```
Use `apideck accounting` to interact with accounting resources.
Resources: invoices, bank-accounts, bills, credit-notes, customers, ledger-accounts, payments, suppliers, tax-rates
List operations: `apideck accounting --list`
Details: `apideck accounting <resource> <verb> --help`
Auth pre-configured. GET auto-approved. POST/PUT/PATCH: use --yes. DELETE: use --force. For JSON: -q -o json
```

**Ships as a Claude Code skill:**
```markdown
---
description: Interact with Apideck Unified API via CLI
---
<output of apideck agent-prompt>
```

Installable via `apideck skill install` which writes to `~/.claude/skills/apideck.md`.

## Spec Caching & Sync

**Source:** Single unified spec from `https://ci-spec-unify.s3.eu-central-1.amazonaws.com/speakeasy-spec.yml`

**Spec stats (v10.24.3):** 47,788 lines YAML, 360 operations, 12 API groups:
accounting, ats, connector, crm, ecommerce, file-storage, hris, issue-tracking, proxy, sms, vault, webhook

**Cache structure:**
```
~/.apideck-cli/
├── config.yaml              # auth credentials
├── permissions.yaml         # permission overrides
├── history.json             # API call log (last 100)
└── cache/
    ├── spec.yml             # raw unified spec
    ├── parsed.bin           # serialized internal model (gob encoding)
    └── meta.json            # version, fetched_at, ttl
```

**Resolution order:**
1. Check cache → if fresh (within TTL of 24h) → load `parsed.bin`
2. If stale → attempt background fetch → use cached while fetching
3. If no cache → use embedded baseline spec (`specs/speakeasy-spec.yml`)
4. If fetch succeeds → parse → write both `spec.yml` and `parsed.bin` (atomic file replacement)

**Parse failure handling:**
- If fetched spec fails to parse → keep existing cached version, log warning
- If no cache and embedded spec fails → fatal error with clear message
- `apideck sync` shows parse errors explicitly so users can report issues

**`apideck sync`** — force re-fetch + re-parse with styled step progress:
```
✓ Fetching spec from S3...          (v10.25.0)
✓ Parsing OpenAPI spec...           (360 operations)
✓ Caching parse tree...             (~/.apideck-cli/cache/)
  Updated: v10.24.3 → v10.25.0
```

**Version pinning** (deferred from v1 — requires versioned S3 URLs or a registry).

## Beautiful DX (Speakeasy-inspired)

**Charmbracelet stack:**

| Library | Purpose |
|---------|---------|
| `lipgloss` | Colors, borders, layout, adaptive themes |
| `bubbletea` | TUI explorer |
| `bubbles` | Spinners, lists, text inputs |
| `huh` | Interactive forms/wizards (auth setup, confirmations) |
| `glamour` | Markdown rendering in terminal |

**Brand identity:**
- Primary accent color (Apideck brand, consistent across all output)
- Adaptive light/dark terminal support via `lipgloss.AdaptiveColor`
- Semantic colors: green=success, red=error, yellow=warning, dim=secondary

**DX touches applied everywhere:**

1. **`apideck auth setup`** — wizard-style `huh` form, not raw prompts
2. **API calls** — spinner during HTTP call, styled table/JSON response
3. **`apideck sync`** — step-based progress tree with status icons
4. **`--help`** — custom Cobra template with lipgloss styling, grouped commands
5. **Errors** — styled boxes: what went wrong (red), why (dim), what to do (emphasized)
6. **TTY detection** — all beauty only when TTY. Piped = raw JSON, zero ANSI codes.

## Distribution

**goreleaser** builds for: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64

| Channel | Command |
|---------|---------|
| Homebrew | `brew install apideck-io/tap/apideck` |
| GitHub Releases | Pre-built binaries + checksums |
| go install | `go install github.com/apideck-io/cli/cmd/apideck@latest` |
| Docker | `docker run apideck/cli accounting invoices list` |

**Docker image:** Based on `gcr.io/distroless/static` (includes CA certificates for HTTPS). Auth via `-e APIDECK_API_KEY=xxx`.

## Tech Stack

| Component | Choice |
|---|---|
| Language | Go |
| CLI framework | Cobra |
| TUI framework | Bubbletea + Lipgloss + Bubbles |
| Interactive forms | Huh |
| Markdown rendering | Glamour |
| OpenAPI parsing | libopenapi (pb33f) |
| Spec source | Single unified spec (S3) |
| Spec cache | Embedded baseline + cached parse tree |
| Auth storage | Env vars > config file |
| HTTP client | net/http + retryablehttp |
| Output formats | JSON, YAML, table (lipgloss), CSV |
| Distribution | goreleaser → brew, GitHub Releases, Docker |
| Agent skill | Ships as Claude Code skill |

## Key Design Decisions

| Decision | Rationale |
|----------|-----------|
| No sdk-go dependency | Spec is the source of truth. SDK would mean two representations of the same API. |
| Single unified spec | Simpler than managing 15+ individual spec URLs. Cached parse tree eliminates perf concerns. |
| Env vars for auth | Standard pattern for AI agents (Claude Code, Codex, Cursor) and CI/CD. |
| libopenapi over kin-openapi | Most full-featured Go OpenAPI 3.x parser, active development. |
| Charmbracelet stack | Industry standard for beautiful Go CLIs. Same stack as Speakeasy. |
| Apideck-only (not generic) | Focused, purpose-built. Generic "any API" is onecli's job. |
