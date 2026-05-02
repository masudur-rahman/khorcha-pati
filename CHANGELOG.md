# Changelog

All notable changes to this project are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [v1.4.0] — 2026-05-02

### Web Dashboard (new)
- Full React SPA with dashboard, transactions, wallets, budgets, and settings pages
- JWT authentication with OTP, QR login (Telegram deep link), and refresh token rotation
- Chi router with REST API handlers for transactions, wallets, contacts, budgets, and summary
- Summary service for dashboard charts and overview
- Transaction CRUD (GetByID, Update, Delete) in repo and service layers
- Report-data API and Statement page for browser-based PDF generation
- Statement PDF print layout with repeating header, page numbers, and auto-named file
- Runtime-configurable Docker setup with web frontend CI release
- Dashboard and landing page UI redesign with mobile improvements
- Currency symbols, layout fixes, favicon, mobile logout, PDF download

### Authentication
- Auth service with OTP, QR login, and refresh token rotation
- Auth repository with tests and mocks
- JWT token generation/parsing and OTP modules
- QR login deep link, contact share handler, and `/start login_` prefix
- Auto-confirm QR login with magic link fallback on expiry
- RefreshToken model, Profile.MobileNumber, and identifier-based user lookup

### Budgeting
- Budget model, repository, and service layer
- `/budget` command with set, delete, and status display
- Budget alerts appended to transaction confirmations
- Budget alert and overall spending tracking fixes

### Bot UX
- Overhaul bot UX with hierarchical Markdown, paginated sorted lists, and configuration-driven AI
- Standardize Telegram response formatting across all handlers
- ForceReply and placeholder for wallet and contact prompts
- AutoKeyboardReset middleware

### AI / NLP
- Overhaul NLP and AI intent engine with structured outputs and entity disambiguation
- Improve AI intent classification and cache accuracy
- Validate AI subcategory IDs and strengthen classification prompts
- Skip measurement units (kg, g, ml, km, pcs) when parsing transaction amounts

### PDF Reports
- Redesign transaction report templates and fix chromedp converter parity
- Split Dockerfile for wkhtmltopdf/chromedp targets
- Unify header layout across PDF converters

### Infrastructure
- Migrate to styx v1.4.0 (context API, MustFilterCols, req tags)
- Environment variable overrides for config with logger upgrade
- Docker/CI multi-stage caching and multi-arch build fixes
- OCI source labels on all Dockerfiles
- Rename WebDashboard config to Server, port changed to int
- Web dashboard HTTP server gated on `WEB_ENABLED=true`

### Bug Fixes
- Fix Unit of Work error handling and zero-value protection in repositories
- Fix styx zero-value filter bug
- Guard against empty contact name in balance update and lookup

### Dependencies (CI)
- Bump actions/checkout 4 → 6
- Bump actions/setup-go 5 → 6
- Bump actions/cache 4 → 5
- Bump codecov/codecov-action 4 → 6
- Bump golangci/golangci-lint-action 7 → 9
- Bump docker/setup-buildx-action 3 → 4
- Bump docker/setup-qemu-action 3 → 4

---

## [v1.0.0] — 2026-02-25

- Initial public release
- Telegram bot for tracking daily transactions
- Interactive /newtxn flow and natural language text parsing
- 13-category / 80+ subcategory taxonomy
- PDF report generation
- Back4App database backend via styx library
- Railway deployment support
