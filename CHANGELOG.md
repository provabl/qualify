# Changelog

All notable changes to qualify will be documented in this file.

The format is based on [Keep a Changelog 1.1.0](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning 2.0.0](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `qualify lab setup` — writes `attest:lab-id` and `attest:admin-level` IAM tags during lab onboarding
- `qualify lab register-role` — registers a researcher's IAM role ARN for tag writing
- `SetIdentityTags()` and `RegisterRoleARN()` on training service

### Changed
- CLI description: "Ark consists of three components" → "qualify consists of three components"
- Default config path: `~/.ark/config.yml` → `~/.qualify/config.yml`

## [0.2.0] - 2026-04-30

### Added
- SLSA Level 2 release workflow (`actions/attest-build-provenance` + cosign keyless + SBOM via syft)
- SPDX-FileCopyrightText headers on all Go source files (2026 Scott Friedman)
- `LICENSE` (Apache 2.0), `LICENSES/Apache-2.0.txt`, `REUSE.toml` for supply chain tooling
- `NOTICE` file
- Migration 000004: compliance training module seed data — 7 modules with full content:
  `cui-fundamentals`, `hipaa-privacy-security`, `security-awareness`, `ferpa-basics`,
  `itar-export-control`, `data-classification`, `nih-research-security`
- `qualify.provabl.dev` documentation site
- Provabl org transfer: module path → `github.com/provabl/qualify`

### Changed
- CLI root command: `Use: "qualify"` (was `"ark"` with qualify alias)
- Fixed stale "Ark consists of three components" CLI help text

## [0.1.0] - 2026-01-01

### Added
- Core training service (`internal/training/service.go`)
- IAM tag writing on training completion: maps 7 module IDs to `attest:*` IAM tags
- `moduleTagMap`: cui-fundamentals, hipaa-privacy-security, security-awareness,
  ferpa-basics, itar-export-control, data-classification, nih-research-security
- Default training expiry: 365 days (RFC3339 timestamp written per module)
- Backend HTTP handlers for training module completion
- Quiz evaluation and progress tracking
- Dashboard statistics endpoint
- Policy check enforcement (training gate)
- Database schema with migrations (PostgreSQL)
- Docker Compose for local development
- Cobra CLI with agent, config, credentials, s3, completion subcommands

[Unreleased]: https://github.com/provabl/qualify/compare/v0.2.0...HEAD
[0.2.0]: https://github.com/provabl/qualify/releases/tag/v0.2.0
[0.1.0]: https://github.com/provabl/qualify/releases/tag/v0.1.0
