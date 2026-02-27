# CLAUDE.md — Expense Tracker Bot

## Project Overview

A Telegram bot for personal finance tracking built in Go 1.24. Supports natural language transaction input (e.g. "spent 500 for food-rest from dbbl"), interactive menus, PDF reports, and multiple database backends.

**Branch**: `natural` (active development) — upstream PR target: `main`

## Architecture

Clean layered architecture with interface-based design:

```
models/          → Pure data structs (Profile, Contacts, Wallet, Transaction, …)
repos/           → Data access interfaces + SQL implementations (styx ORM)
services/        → Business logic interfaces + implementations
  services/all/  → Dependency injection registry (GetServices())
api/
  api/tele.go    → Telegram bot route registration
  api/handlers/  → Standalone handler functions (not struct methods)
  api/wizard/    → Server-side wizard state store (10-min TTL)
modules/
  modules/cache/ → Pluggable cache (memory / Redis)
  modules/transaction/parser.go → Natural language parser
  modules/ai/    → Gemini / OpenRouter AI integrations
  modules/google/→ Google Drive DB backup/restore
configs/         → Config loading, database init, cache init
cmd/             → Cobra CLI (root + serve commands)
pkg/             → Utility packages (PDF, formatter, time helpers, …)
  pkg/health/    → HTTP health-check handler (JSON /health endpoint)
  pkg/telegram/  → SplitMessage() helper (≤4000 chars per Telegram message)
infra/logr/      → Structured logger (Zap)
```

## Key Naming Conventions (post-refactor)

| Old name            | New name        | Notes                                   |
|---------------------|-----------------|-----------------------------------------|
| Account             | Wallet          | models.Wallet, WalletService, WalletRepo |
| DebtorCreditor/DrCr | Contact         | models.Contacts, ContactService, etc.   |
| User (bot owner)    | Profile         | models.Profile; telebot.User unchanged  |
| models/user.go      | models/profile.go | Contains both Profile and Contacts     |
| repos/accounts/     | repos/wallets/  | Package name: `wallets`                 |
| services/accounts/  | services/wallets/ | Package name: `wallets`               |
| /users command      | /contacts       | Telegram bot command                    |

## Handler Pattern

Handlers are **standalone functions**, not struct methods. They access services via:

```go
all.GetServices().User          // ProfileService
all.GetServices().Wallet        // WalletService
all.GetServices().Transaction   // TransactionService
all.GetServices().Contact       // ContactService (formerly DrCr)
```

Register handlers in `api/tele.go`.

## Database

Supports SQLite (default, syncs to Google Drive) and PostgreSQL. Schema auto-synced via `styx` ORM on startup. All models must be listed in `configs/database.go:syncTables()`.

Models synced: `Profile`, `Contacts`, `Wallet`, `Transaction`, `TxnCategory`, `TxnSubcategory`, `Event`

## Common Commands

```bash
# Run locally
go run . serve

# Build
go build -o bin/expense-tracker .

# Tests (native)
go test ./...
go test -v -race ./...

# Lint (requires golangci-lint)
golangci-lint run ./...

# Check build
go build ./...

# Docker (multi-stage, preserves wkhtmltopdf for PDF)
docker build -t expense-tracker .

# Makefile targets
make run           # native go run
make test          # native go test with coverage
make vet           # go vet
make lint          # golangci-lint
make check         # vet + test
make tidy          # go mod tidy + verify
make all-build     # docker-based cross-compile
make all-container # docker-based container build
make release       # docker multi-arch push + manifest
make version       # print version info
make help          # list all targets
```

## Environment Variables

See `.env.example` for the full list. Key required vars:

- `TELEGRAM_BOT_TOKEN` — from @BotFather
- `PARSE_APP_ID`, `PARSE_REST_API_KEY`, `PARSE_SERVER_URL` — Back4App credentials
- `BOT_MODE` — `polling` (local dev) or `webhook` (production)
- `BASE_URL` — used for periodic health ping (optional)

Config validation: `configs.Validate()` (call at startup in main if desired).

## Transaction Parser

`modules/transaction/parser.go` — parses natural language like:
- `"spent 1000 for food-rest from dbbl yesterday"`
- `"lend 5000 to karim from brac"`
- `"earn 65k fin-sal to dbbl on 2024-03-15"`

Requires `ContactVerifier` and `AccountVerifier` function types for resolving names.

## New Files Added (refactor)

| File | Purpose |
|------|---------|
| `api/wizard/state.go` | Thread-safe wizard state store for /newtxn flow |
| `api/handlers/undo.go` | /undo handler stub (needs TransactionService.Undo impl) |
| `pkg/health/health.go` | JSON health endpoint (alternative to existing /healthz) |
| `pkg/telegram/helpers.go` | SplitMessage(), FormatAmount() helpers |
| `configs/validate.go` | Required env var validation |
| `.env.example` | Environment variable template |
| `.golangci.yml` | golangci-lint configuration |
| `CHANGELOG.md` | Change log |

## Pending Manual Implementation

These features require code to be written — see CHANGELOG.md for details:

1. `TransactionService.Undo(userID int64)` — soft-delete last transaction + revert balances
2. Wire `api/wizard/state.go` into `/newtxn` handler to fix 64-byte callback_data limit
3. Wire `pkg/telegram/SplitMessage()` into `/expense`, `/summary`, `/allsummary` handlers
4. Register `/contacts` and `/undo` in `api/tele.go`
5. Optimistic locking in `WalletRepo.UpdateBalance` using `Wallet.Version` field

## Refactor Script

The original automated refactor script is at `hack/refactor/main.go`.
Run with: `go run hack/refactor/main.go --dry-run --verbose`
