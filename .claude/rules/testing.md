---
globs: "**/*_test.go"
---

# Testing Rules (Go)

- Every new function gets a test. No exceptions.
- Test name format: `TestFunctionName_condition` or table-driven with descriptive names.
- Arrange-Act-Assert pattern.
- No test interdependencies. Each test sets up its own state.
- Mock external services (DB, APIs, Redis), not internal logic.
- Use `t.Skip()` for tests requiring external deps (wkhtmltopdf, Redis, AI API keys).
- Use `t.Parallel()` where safe.
- Use `testify/assert` for assertions (already a project dependency).
- Table-driven tests for functions with multiple input/output combinations.
- Use `cache.CacheMap` (not Redis) in unit tests for cache-dependent code.
