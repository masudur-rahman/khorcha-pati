---
globs: "**/*.go"
---

# Code Quality Rules (Go)

- Every function has a single responsibility.
- No function exceeds 30 lines. Extract if it does.
- No file exceeds 300 lines. Split if it does.
- All exported functions/methods have a godoc comment (one line: purpose).
- Imports: stdlib -> internal packages -> external libs. Blank line between groups.
- No magic numbers/strings. Extract to named constants.
- Prefer early returns over nested conditionals.
- Always check errors. Use `_` only when explicitly intentional (with comment).
- Use `context.Context` as first parameter where applicable.
- No `interface{}` — use `any` (Go 1.18+) or typed interfaces.
- Prefer value receivers unless mutation is needed.
- Handlers are standalone functions, not struct methods. Access via `all.GetServices()`.
