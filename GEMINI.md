# Project: expense-tracker-bot

## Global Rules & Mandates
- **Response Style**: Code only. No explanations unless asked. No filler phrases. No preamble. When asked to explain, max 3 sentences per point.
- **Strict Workflow**:
  1. **Ask**: If requirements are ambiguous, ask targeted questions BEFORE coding (max 1 round).
  2. **Plan**: Before ANY code change, outline a plan.
  3. **Approve**: Wait for explicit user approval before writing code.
  4. **Implement**: Code the approved plan, following project rules.
  5. **Test**: Run tests/lints/builds immediately after implementation. Report pass/fail. Fix failures before proceeding.
  6. **Review**: Self-review changes. Report issues.
  7. **Remember**: Append new learnings to the `## Knowledge Base (Learnings)` section.

## Overview
Telegram bot for personal finance tracking. Supports natural language transaction input, interactive menus, PDF reports, category classification via Gemini AI, and multiple database backends (SQLite + PostgreSQL).

## Stack
Go 1.24 | telebot.v3 | styx ORM (SQLite/PostgreSQL) | Cobra CLI | Zap logger | Redis/go-cache | Gemini AI | Google Drive sync | wkhtmltopdf | Docker multi-stage | Ansible | Terraform | Kubernetes | Helm

## Architecture
This project follows a layered architecture:
- **models/**: Data structures (Profile, Wallet, Contacts, Transaction, Event).
- **repos/**: Data access interfaces + styx ORM implementations.
- **services/**: Business logic interfaces + implementations.
- **services/all/**: DI registry — use `all.GetServices()` to access services.
- **api/**: Bot route registration (`tele.go`) and standalone handler functions (`handlers/`).
- **modules/**: Pluggable components (cache, transaction parser, AI category classifier, Google Drive sync).
- **configs/**: Configuration loading, DB initialization, and environment validation.
- **cmd/**: Cobra CLI commands (root, serve).
- **pkg/**: Utility packages (PDF generation, formatting, health checks).
- **infra/logr/**: Structured logging wrapper (Zap).

## Core Mandates & Conventions
- **Naming Conventions (CRITICAL)**:
  - `Wallet` (instead of Account/Bank)
  - `Contact` (instead of DebtorCreditor)
  - `Profile` (instead of User/Bot Owner)
- **Handler Pattern**: Handlers MUST be standalone functions, not struct methods. They access services via `all.GetServices()`.
- **Registration**: All new handlers must be registered in `api/tele.go`.
- **Database**: New models must be added to `configs/database.go:syncTables()` for auto-sync via styx ORM.
- **Response Format**: Use Telegram Markdown (`*bold*`, `` `code` ``, `_italic_`). `Transaction.Summary()` is the standard emoji-rich formatting pattern.
- **Import Order**: stdlib -> internal -> external (blank line between groups).

## Engineering Standards

### Coding Rules (Go)
- **Single Responsibility**: Each function should have one purpose and not exceed 30 lines.
- **File Length**: Limit files to 300 lines; split if they grow larger.
- **Documentation**: All exported symbols must have a single-line godoc comment.
- **Conditionals**: Prefer early returns to avoid deeply nested code.
- **Error Handling**: Always check errors. Only use `_` when intentional and documented with a comment.
- **Context**: Pass `context.Context` as the first parameter where appropriate.
- **Types**: Use `any` instead of `interface{}` (Go 1.18+).

### Testing Rules (Go)
- **Mandatory Tests**: Every new function MUST have a corresponding test.
- **Pattern**: Use the Arrange-Act-Assert (AAA) pattern.
- **Mocks**: Mock external services (DB, APIs, Redis); do not mock internal logic.
- **Dependencies**: Use `t.Skip()` for tests requiring external dependencies (e.g., wkhtmltopdf, AI API keys).
- **Assertions**: Use `github.com/stretchr/testify/assert`.
- **Cache**: Use `cache.CacheMap` for unit tests instead of Redis.

## Key Code Patterns

### New Service
- Interface in `services/<name>.go`.
- Implementation in `services/<name>/<name>.go` (verify with `var _ services.Interface = &impl{}`).

### New Repository
- Interface in `repos/<name>.go` (include `WithUnitOfWork`).
- SQL implementation in `repos/<name>/<name>.go` using styx ORM.

### New Handler
```go
func HandleExample(ctx telebot.Context) error {
    svc := all.GetServices()
    profile, err := svc.User.GetUserByTelegramID(ctx.Sender().ID)
    if err != nil {
        return ctx.Send(models.ErrCommonResponse(err))
    }
    // Business logic...
    return ctx.Send("Success ✅", telebot.ModeMarkdown)
}
```

## Development Workflow
1. **Phased Workflow (CRITICAL)**: When a plan has multiple phases, implement ONE phase at a time. After each phase:
   - Run build/test.
   - Show summary of changes.
   - Wait for user review.
   - Ask if they want to commit.
   - Commit if yes.
   - ONLY then proceed to the next phase. NEVER implement all phases at once.
2. **Execution**: One refactor at a time; never mix refactoring with feature work.
3. **Validation**: Run `go build ./...`, `go vet ./...`, and `go test ./...`. Conduct a thorough review for error handling, naming consistency, and test coverage.

## Git & Source Control
- **Commits**: No conventional commit prefixes (no `feat:`, `refactor:`, etc.). Just a plain description.
- **Formatting**: Capitalize the first character of the commit message. Use `git commit -s` (sign-off) flag. Do NOT add Co-Authored-By.
- **Scope**: One logical change per commit.
- **Branching**: `type/short-description`

## Project Commands
- `go run . serve` / `make run`: Run the bot locally.
- `go build ./...`: Verify compilation.
- `go test ./...` / `make test`: Run tests (with coverage).
- `golangci-lint run ./...` / `make lint`: Run linting.
- `make check`: Run vet and tests.

## Knowledge Base (Learnings)
- **Memory Management**: 
  - "remember this" -> append to this file under `## Knowledge Base (Learnings)`
  - "remember globally" -> use the `save_memory` tool to persist globally.
- styx ORM: Use `int64` unix timestamps instead of `*time.Time`.
- Tickers: Use `for range ticker.C` instead of `for { select { case <-ticker.C: } }`.
- fmt: `fmt.Fprintln` already adds a newline; do not include `\n` in the string.
- Parser: Avoid using day-name substrings (e.g., "fri") in test inputs to prevent false matches.
- Hooks: Always call React Hooks (e.g., useMemo) at the top level, before any early returns (e.g., if (isLoading) return ...), to maintain consistent hook order across renders.

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
