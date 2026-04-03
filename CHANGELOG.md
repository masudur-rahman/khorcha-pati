# Changelog

All notable changes to this project are documented here.
Format follows [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

---

## [Unreleased] — natural branch

### Added
- **Unit of Work (UoW) Fix**: Refactored service layer to use named return variables, ensuring robust error capturing and transactional integrity for all balance-modifying operations.
- **Bot UX Overhaul**: Replaced boxed tables and images with premium hierarchical MarkdownV2 summaries using tree connectors (├, └).
- **Stateful Pagination**: Implemented `⬅️ Previous` and `Next ➡️` navigation for `/list` and `/expense` commands.
- **AutoKeyboardReset Middleware**: Automatically removes sticky `ForceReply` or `ReplyKeyboardMarkup` when switching between commands.
- **Natural Language Refinement**: Added `add` and `plus` keywords to quickly initialize or adjust balances; improved entity disambiguation hierarchy (Contact > Wallet > Remark).
- **Architecture**: Centralized AI provider selection (Gemini/OpenRouter) and keys into a configuration-driven model with secure environment overrides.
- **Scenario Testing**: Added end-to-end "User Journey" tests simulating full cycles of income, expenses, and lending.
- **Docker Optimization**: 
    - Optimized build context via strict `.dockerignore` (246MB -> <1MB).
    - Implemented a "Base Layer" pattern for Chromium dependencies to speed up builds.
    - Fixed architecture-aware `wkhtmltopdf` downloads for ARM64/AMD64 compatibility.
- **Security**: Integrated `govulncheck` into the CI/CD pipeline.

### Changed
- /list command now limits output to the **last 30 days** for improved performance and scalability.
- Default AI selection now prefers **Gemini** if the API key is provided in the environment.
- Deprecated `pkg/printer.go` and `pkg/formatter.go` for Telegram-based responses.

---

## [v1.0.0] — 2026-02-25

- Initial public release
- Telegram bot for tracking daily transactions
- Interactive /newtxn flow and natural language text parsing
- 13-category / 80+ subcategory taxonomy
- PDF report generation
- Back4App database backend via styx library
- Railway deployment support
