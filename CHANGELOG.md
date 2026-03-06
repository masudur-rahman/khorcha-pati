# Changelog

All notable changes to this project are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [Unreleased] — natural branch

### Added
- **Naming refactor**: Account → Wallet, Debtor/Creditor → Contact, User (bot owner) → Profile
- **Soft-delete**: Transaction.DeletedAt field; /undo command to reverse the most recent transaction
- **Optimistic locking**: Wallet.Version field to prevent concurrent balance race conditions
- **Wizard state store**: Server-side state for /newtxn flow (fixes 64-byte callback_data limit)
- **Message splitting**: SplitMessage() helper to stay within Telegram's 4096-byte message limit
- **Health endpoint**: pkg/health HTTP handler for Railway / Docker HEALTHCHECK
- **TTL cache**: pkg/cache for wallet lists and category taxonomy (reduces Back4App round-trips)
- **Config validation**: configs.Validate() fails fast on startup if required env vars are missing
- **.env.example**: Template listing all required/optional environment variables
- **.golangci.yml**: Comprehensive lint configuration
- **Multi-arch Docker**: Makefile docker-buildx target and CI release.yml using docker/build-push-action
- **Cross-compilation**: Makefile cross-build target for linux, darwin, windows (amd64 + arm64)
- **ldflags**: VERSION, BUILD_DATE, GIT_COMMIT embedded into binary at build time
- **BuildKit**: DOCKER_BUILDKIT=1 enabled globally in Makefile; cache mounts in Dockerfile
- **Distroless runtime**: Dockerfile now uses gcr.io/distroless/static:nonroot (from Alpine)
- **Non-root container**: USER nonroot:nonroot in Dockerfile
- **CI pipeline**: .github/workflows/ci.yml with test, lint, vuln scan, build jobs
- **Release pipeline**: .github/workflows/release.yml with multi-arch Docker + GitHub Release
- **Parser tests**: modules/parser_test.go with table-driven cases
- **CHANGELOG.md**: This file

### Changed
- /users command renamed to /contacts
- Menu labels: "Account" → "Wallet", "Person" → "Contact"
- Contact.Balance renamed to Contact.NetBalance (positive = they owe you)
- Contact struct gains Handle field (short name for text parsing)
- Profile struct gains Timezone field

### Natural branch parser improvements
- Wider action-verb vocabulary (paid, received, repaid, collected, ...)
- Quoted note fields: note "Lunch with team"
- Relative date expressions: yesterday, last monday, -3d
- Fuzzy wallet name matching (partial names resolve if unambiguous)
- Descriptive error messages on parse failure

---

## [v1.0.0] — 2026-02-25

- Initial public release
- Telegram bot for tracking daily transactions
- Interactive /newtxn flow and natural language text parsing
- 13-category / 80+ subcategory taxonomy
- PDF report generation
- Back4App database backend via styx library
- Railway deployment support
