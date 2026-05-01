# Changelog

All notable changes to qualify will be documented in this file.

The format is based on [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.2] - 2026-05-01

### Added

- **`internal/auth/` package**: JWT-based authentication for the qualify backend (closes #33). Includes `Config`, `IssueToken`, `ValidateToken`, chi middleware, and context helpers. 12 unit tests.
- **`GET /api/auth/dev-token`**: Issues a signed JWT for the configured dev user. Only available when `AUTH_DEV_MODE=true`. Returns 404 in production.
- **`GET /api/auth/me`**: Returns the authenticated user's identity from their token.
- **`internal/license/` package**: Network-based license validation against `https://licensing.provabl.co`. Results cached in `system_config` table with configurable TTL. Falls back to `CommunityLicense()` (open-source tier) when key is absent or server unreachable.
- **Migration 000009** (`system_config`): Key/value cache table with TTL for license validation and future feature flags.
- **`compose.env.example`**: Deployment environment template documenting all variables.
- **`DEPLOYMENT.md`**: Comprehensive self-hosted deployment guide — quick start, all env vars, auth flow progression (dev → JWT → OIDC), content packs, Kubernetes, backup, troubleshooting.
- **`kubernetes/`**: Reference Kubernetes manifests (namespace, ConfigMap, Deployment, Service, kustomization). Non-root, read-only filesystem, resource limits.

### Changed

- All `cmd/ark-backend` handlers now extract `user_id` from JWT context rather than URL params or request body — prevents privilege escalation.
- `setupRouter` requires `auth.Config`; protected routes grouped under `auth.Middleware`. Public routes (health, module listing, auth endpoints) bypass auth.
- `training.Service` gains `GetUserProfile()` and `UpdateUserProfile()` — real DB queries replacing hardcoded mock responses.
- `web/src/App.tsx`: removed hardcoded `USER_ID`. App fetches JWT from `/api/auth/dev-token` on mount, stores in `sessionStorage`, resolves user via `/api/auth/me`.
- `web/src/services/agent.ts`: all backend requests include `Authorization: Bearer <token>` header. Added `getToken/setToken/clearToken` and `getMe()`.
- `docker-compose.yml`: added `AUTH_DEV_MODE`, `JWT_SECRET`, `LICENSE_KEY` env vars.
- `go.mod`: updated Go directive from `1.24.0` to `1.26.0` to match toolchain.

## [0.1.1] - 2026-04-30

### Security

- **`moduleTagMap` unexported** (`internal/training/tags.go`): was an exported mutable global; external callers could inject or overwrite module→tag mappings. Now unexported with `TagForModule()` and `ModuleIDs()` read-only accessors.
- **CORS explicit origins** (`cmd/ark-backend/main.go`): replaced `localhost:*` wildcard (matches any port) with explicit ports 5173 and 5174. Added auth warning comment on unauthenticated `/api/*` routes.

### Fixed

- **`TrainingContent` TypeScript type** (`web/src/types/api.ts`): added `quiz?: QuizQuestion[]` and `passing_score?: number` fields. `TrainingModule.tsx` now parses module content as `TrainingContent` with a typed cast instead of duck-typing on `any`.
- **`OnboardingWizard` type cast** (`web/src/components/onboarding/OnboardingWizard.tsx`): removed `as any` cast on `updateUserProfile` call; uses `satisfies Partial<UserProfile>` with explicit `UserPreferences` structure.
- **Backend URL configurable** (`web/src/services/agent.ts`): `BACKEND_URL` and `AGENT_URL` now read from `VITE_BACKEND_URL` / `VITE_AGENT_URL` env vars with localhost fallback. Documented in `CONTRIBUTING.md`.

### Added

- **7 new tests for `RecordCountryCheck`**: invalid code rejection (5 cases), DB update, IAM tag writes (`attest:country`, `attest:coc-check-current`, `attest:coc-check-expiry`), expiry ~1 year, `TagForModule`/`ModuleIDs` accessors, immutability.

### Docs

- **README/CONTRIBUTING**: updated `React + Cloudscape` → `React + Radix UI + Tailwind`.
- **README**: removed stale pending-rename notes for `cmd/ark-*` directories.
- **CONTRIBUTING**: updated `moduleTagMap` reference to new location; added web env vars section.

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
