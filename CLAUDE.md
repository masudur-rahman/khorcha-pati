# Project: expense-tracker-bot

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
See `.env.example`. Required: `TELEGRAM_BOT_TOKEN`, `PARSE_APP_ID`, `PARSE_REST_API_KEY`, `PARSE_SERVER_URL`

## Pending Work
1. `TransactionService.Undo()` — soft-delete last txn + revert balances
2. Wire `api/wizard/state.go` into /newtxn handler (64-byte callback_data limit)
3. Wire `pkg/telegram/SplitMessage()` into /expense, /summary, /allsummary
4. Optimistic locking in WalletRepo.UpdateBalance using Wallet.Version

## Active Context
Completed full Account/DebtorCreditor -> Wallet/Contact rename. CI/CD passing. Bot responses use emoji-rich formatting.

## Learned
- styx ORM cannot handle `*time.Time` pointer fields — use `int64` unix timestamps
- `for { select { case <-ticker.C: } }` triggers gosimple S1000 — use `for range ticker.C`
- golangci-lint v1.64+ removed `run.skip-dirs` — use `issues.exclude-dirs` instead
- Parser test: "friends" contains "fri" which matches Friday — avoid day-name substrings in test inputs
- `fmt.Fprintln` already adds newline — don't pass `"text\n"` (triggers go vet)
- Cache types: `cache.CacheMap` (memory), `cache.CacheRedis` — no `CacheMemory`
