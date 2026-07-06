# Project: khorcha-pati

## Overview
Telegram bot for personal finance tracking. Supports natural language transaction input, interactive menus, PDF reports, category classification via Gemini AI, and multiple database backends (SQLite + PostgreSQL).

## Stack
Go 1.24 | telebot.v3 | styx ORM (SQLite/PostgreSQL) | Cobra CLI | Zap logger | Redis/go-cache | Gemini AI | Google Drive sync | wkhtmltopdf | Docker multi-stage

## Architecture
```
models/             Data structs (Profile, Wallet, Contacts, Transaction, Event)
repos/              Data access interfaces + styx ORM implementations
services/           Business logic interfaces + implementations
  services/all/     DI registry — all.GetServices()
api/
  tele.go           Bot route registration
  handlers/         Standalone handler functions (NOT struct methods)
  wizard/           Server-side wizard state for /newtxn (10-min TTL)
modules/
  cache/            Pluggable cache (memory map / Redis)
  transaction/      Natural language parser (amount, date, wallet, contact)
  ai/               Gemini LLM category classifier
  google/           Google Drive SQLite backup/restore
  convert/          Transaction model to GraphQL type conversion
configs/            Config loading, DB init, env validation
cmd/                Cobra CLI (root + serve)
pkg/                Utilities (PDF, formatting, health endpoint, Levenshtein)
infra/logr/         Structured logger (Zap wrapper)
```

## Commands
```bash
go run . serve                # run bot locally
go build ./...                # verify compilation
go test ./...                 # run tests
go test -v -race ./...        # tests with race detector
golangci-lint run ./...       # lint
make run                      # native go run
make test                     # test with coverage
make lint                     # golangci-lint
make check                    # vet + test
make docker-build             # single-arch Docker image
make all-container            # multi-arch Docker build
```

## Conventions
- Handlers are standalone functions, access services via `all.GetServices().User / .Wallet / .Contact / .Txn`
- Register handlers in `api/tele.go`
- Import order: stdlib -> internal -> external (blank line between groups)
- Models synced in `configs/database.go:syncTables()` — add new models there
- Naming: Profile (bot user), Wallet (cash/bank), Contacts (people), Transaction
- Bot responses use Telegram Markdown (`*bold*`, `` `code` ``, `_italic_`)
- Transaction.Summary() is the standard emoji-rich formatting pattern for bot output

## Key Naming (post-refactor)
| Old               | New       | Notes                                       |
|-------------------|-----------|---------------------------------------------|
| Account           | Wallet    | models.Wallet, WalletService, repos/wallets |
| DebtorCreditor    | Contact   | models.Contacts, ContactService             |
| User (bot owner)  | Profile   | models.Profile; telebot.User unchanged      |

## Database
- SQLite (default, Google Drive sync) + PostgreSQL supported
- Schema auto-synced via styx ORM on startup
- Wallet.Version field for optimistic concurrency (not yet enforced)

## Environment
See `.env.example`. Required: `TELEGRAM_BOT_TOKEN`

## Pending Work
1. Wire `api/wizard/state.go` into /newtxn handler (64-byte callback_data limit)
2. Telegram response format audit — standardize primitive response formats
3. Budgeting — allow users to set and track spending budgets

## Active Context
Completed full Account/DebtorCreditor -> Wallet/Contact rename. CI/CD passing. Bot responses use emoji-rich formatting.

## Learned
- styx ORM cannot handle `*time.Time` pointer fields — use `int64` unix timestamps
- styx ORM zero-value filtering is FIXED in v1.3.0+ — use `MustFilterCols("col")` to force zero-value fields in WHERE clauses. Also available: `req` struct tag (e.g. `db:"deleted_at,req"`). The old bug where empty strings/zero ints were silently dropped from WHERE is solved.
- `for { select { case <-ticker.C: } }` triggers gosimple S1000 — use `for range ticker.C`
- golangci-lint v1.64+ removed `run.skip-dirs` — use `issues.exclude-dirs` instead
- Parser test: "friends" contains "fri" which matches Friday — avoid day-name substrings in test inputs
- `fmt.Fprintln` already adds newline — don't pass `"text\n"` (triggers go vet)
- Cache types: `cache.CacheMap` (memory), `cache.CacheRedis` — no `CacheMemory`
- Amount parser: numbers followed by units (kg, km, g, ml, l, pcs) must be skipped — they're quantities, not monetary amounts

<!-- code-review-graph MCP tools -->
## MCP Tools: code-review-graph

**IMPORTANT: This project has a knowledge graph. ALWAYS use the
code-review-graph MCP tools BEFORE using Grep/Glob/Read to explore
the codebase.** The graph is faster, cheaper (fewer tokens), and gives
you structural context (callers, dependents, test coverage) that file
scanning cannot.

### When to use graph tools FIRST

- **Exploring code**: `semantic_search_nodes` or `query_graph` instead of Grep
- **Understanding impact**: `get_impact_radius` instead of manually tracing imports
- **Code review**: `detect_changes` + `get_review_context` instead of reading entire files
- **Finding relationships**: `query_graph` with callers_of/callees_of/imports_of/tests_for
- **Architecture questions**: `get_architecture_overview` + `list_communities`

Fall back to Grep/Glob/Read **only** when the graph doesn't cover what you need.

### Key Tools

| Tool | Use when |
|------|----------|
| `detect_changes` | Reviewing code changes — gives risk-scored analysis |
| `get_review_context` | Need source snippets for review — token-efficient |
| `get_impact_radius` | Understanding blast radius of a change |
| `get_affected_flows` | Finding which execution paths are impacted |
| `query_graph` | Tracing callers, callees, imports, tests, dependencies |
| `semantic_search_nodes` | Finding functions/classes by name or keyword |
| `get_architecture_overview` | Understanding high-level codebase structure |
| `refactor_tool` | Planning renames, finding dead code |

### Workflow

1. The graph auto-updates on file changes (via hooks).
2. Use `detect_changes` for code review.
3. Use `get_affected_flows` to understand impact.
4. Use `query_graph` pattern="tests_for" to check coverage.
