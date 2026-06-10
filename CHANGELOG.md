# Changelog

All notable changes to qualify will be documented in this file.

The format is based on [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- **Versioned `attest:*` tag schema contract** (closes #32): the `attest:*` IAM tag namespace shared
  with attest is now anchored by a canonical, machine-readable
  `internal/training/attest-tags-schema.json` (byte-identical with attest's copy) plus a
  `SchemaVersion` constant. A conformance test (`internal/training/tags_schema_test.go`) locks
  qualify's tag constants, `moduleTagMap`, and `ModuleExpiryTag` to the schema, so any key drift
  fails CI instead of silently breaking Cedar evaluation. No new module dependency (the two
  products stay decoupled — the byte-identical schema + shared `SchemaVersion` are the contract).
  See attest's `docs/integrations/qualify.md`.
- **`attest:*` IAM tags written through the provabl/evidence kernel** (closes #38): training-module
  completion now appraises via the evidence kernel `(ASP, appraiser)` pair and lowers the verdict to
  the `attest:*` role tags attest's principal resolver reads, instead of writing tags directly.
  Ephemeral per-run AM key; the kernel is the single producer of the lowered attributes.
- **Identity + countries-of-concern routed through the kernel** (closes #40): the researcher's
  identity attributes and countries-of-concern (COC) determination flow through the same kernel
  lowering as the training tags, so `attest:*` identity/COC tags carry the same provenance as the
  training tags rather than being set out-of-band.

### Changed

- **Completed the `ark` → `qualify` rename** (closes #34): renamed `cmd/ark-agent` → `cmd/qualify-agent` and `cmd/ark-backend` → `cmd/qualify-backend`. Build paths, Dockerfiles, release artifact names, and docs updated to match. The binaries were already emitted as `qualify-agent` / `qualify-backend`; the command directories now agree.
- **Bumped Go directive to 1.26.4** to align with the patched standard library (clears the GO-2026-* advisories).

### Fixed

- **Agent daemon binary lookup**: `findAgentBinary` searched `PATH` and the executable directory for `ark-agent`, but the build emits `qualify-agent` — so daemon-managed agent startup could never find it. Now looks up `qualify-agent`.

### Removed

- Removed a stray ~10 MB `ark-backend` binary that had been committed to the repo root; added the root-built binary names to `.gitignore`.

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
