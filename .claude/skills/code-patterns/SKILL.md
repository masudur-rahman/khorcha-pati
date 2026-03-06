---
name: code-patterns
description: Reference for Go code patterns in this project. Read before creating new components.
user-invocable: false
---

# Code Patterns

Read existing code to discover patterns before writing new code.

## Discovery Process
1. Find 2-3 similar existing files to what you're creating.
2. Note: file structure, naming, imports, error handling, exports.
3. Follow the same patterns exactly.

## Key Patterns

### New Service
```
services/<name>.go          — interface definition
services/<name>/<name>.go   — implementation (var _ services.Interface = &impl{})
```

### New Repository
```
repos/<name>.go             — interface definition (with WithUnitOfWork)
repos/<name>/<name>.go      — SQL implementation using styx ORM
```

### New Handler
```go
// api/handlers/<name>.go — standalone function, NOT a method
func HandleSomething(ctx telebot.Context) error {
    svc := all.GetServices()
    user, err := svc.User.GetUserByTelegramID(ctx.Sender().ID)
    if err != nil {
        return ctx.Send(models.ErrCommonResponse(err))
    }
    // ... business logic via svc.Wallet / svc.Contact / svc.Txn
    return ctx.Send("result", telebot.ModeMarkdown)
}
```
Register in `api/tele.go`: `bot.Handle("/command", handlers.HandleSomething)`

### New Model
Add struct to `models/`. Add to `configs/database.go:syncTables()` for auto-sync.

## If No Pattern Exists
- Ask the user which existing file to use as reference.
- Do not invent new patterns without approval.
