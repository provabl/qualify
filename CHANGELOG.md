# Changelog

All notable changes to qualify will be documented in this file.

The format is based on [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-04-30

First release — Foundation milestone complete.

### Added

- **`qualify train start <module>`**: interactive CLI training loop with section-by-section presentation, markdown-lite rendering (ANSI bold/headers/blockquotes on TTY, plain text in CI), interactive quiz, retry on fail. Progress saved to `~/.qualify/progress/` between sessions.
- **`qualify train required`**: reads `.attest/sre.yaml` for active compliance frameworks and shows required modules. Works offline.
- **`qualify train status`**: shows completion, expiry, and unlock context per module (what AWS access each training gates).
- **`qualify train certificate <module>`**: displays or re-displays a completion certificate (box-drawing format). Certificates auto-issued on pass and saved to `~/.qualify/certificates/`.
- **`qualify lab setup`**: assigns researcher to a lab; writes `attest:lab-id` and `attest:admin-level` IAM tags.
- **`qualify lab register-role`**: stores IAM role ARN for tag writes.
- **`qualify lab record-check --user --country --performed-by`**: records a countries-of-concern compliance check. Writes `attest:country`, `attest:coc-check-current`, `attest:coc-check-expiry` IAM tags; stores check metadata in DB.
- **`qualify onboard`**: guided new-user onboarding.
- **`internal/training/tags.go`**: all `attest:*` IAM tag key constants and `ModuleTagMap` — single authoritative source shared between service (writer) and CLI (display). Schema version 1.
- **`internal/localaudit/`**: JSONL audit log at `~/.qualify/audit.log`. Records all training events with UTC timestamps. Always available without backend.
- **8 training modules**: security-awareness, data-classification, cui-fundamentals, hipaa-privacy-security, ferpa-basics, itar-export-control, nih-research-security (NOT-OD-26-017), countries-of-concern-awareness (NOT-OD-25-083). Each: 3 sections + 5-question quiz, 80% passing score.
- **Migration 000008**: adds `institutional_affiliation_country`, `affiliation_check_performed_at/by` to users table.
- **Backend** (`cmd/ark-backend`): slog JSON structured logging, request ID middleware, `/health`, `/ping`, training and dashboard API endpoints.
- **Docker Compose**: local dev environment with PostgreSQL (`make docker-up`).
- **CI**: `test.yml` (backend + frontend), `check.yml` (fast pre-commit), `release.yml` (SLSA L2).
- **`README.md`** + **`CONTRIBUTING.md`**: complete project documentation.

### Security

- `parseAnswer`: rune arithmetic prevents byte overflow for option counts > 9.
- `renderText`: strips pre-existing ANSI escape sequences from DB content before processing.
- `database.New`: error messages never include the DSN (which contains the password).
- Probe interface (ground): validated against `^[a-z0-9][a-z0-9-]{0,62}$`; relative paths rejected.

[Unreleased]: https://github.com/provabl/qualify/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/provabl/qualify/releases/tag/v0.1.0
