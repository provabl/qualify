# qualify

**Compliance training and access gating for AWS Secure Research Environments.**

qualify is the training layer of the [Provabl](https://provabl.dev) suite. Researchers complete compliance training through an interactive CLI; completion writes IAM tags to their role; attest's Cedar PDP grants or denies access based on those tags.

```
qualify train required                 # see what training your SRE needs
qualify train start cui-fundamentals   # interactive sections + quiz
qualify train status                   # show completion and expiry
```

[![Tests](https://github.com/provabl/qualify/actions/workflows/test.yml/badge.svg)](https://github.com/provabl/qualify/actions/workflows/test.yml)
[![Apache 2.0](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

---

## How it works

1. **`attest compile`** produces a crosswalk that maps compliance frameworks to required training modules (e.g., CMMC Level 2 → CUI Fundamentals + Security Awareness).
2. **`qualify train required`** reads the active attest frameworks and shows which modules the user needs.
3. **`qualify train start <module>`** presents sections of content and a quiz. Progress is saved between sessions (`~/.qualify/progress/`).
4. On passing, qualify calls `svc.CompleteModule()` which writes an IAM tag to the user's role: `attest:cui-training = true`.
5. attest's Cedar PDP evaluates that tag on the next access request. Until the tag is present (and unexpired), access to CUI environments is denied.

Completion certificates are saved to `~/.qualify/certificates/`. Re-display any certificate with `qualify train certificate <module-id>`.

---

## Included training modules

| Module ID | Title | Frameworks |
|---|---|---|
| `security-awareness` | Security Awareness | All |
| `data-classification` | Data Classification | All |
| `cui-fundamentals` | CUI Fundamentals | CMMC L1/L2, NIST 800-171 |
| `hipaa-privacy-security` | HIPAA Privacy & Security | HIPAA |
| `ferpa-basics` | FERPA Basics | FERPA |
| `itar-export-control` | ITAR Export Control | ITAR |
| `nih-research-security` | NIH Research Security | NIH GDS |
| `countries-of-concern-awareness` | Countries of Concern | NIH GDS |

Each module has 3 sections and a 5-question quiz (80% to pass, 2 attempts).

---

## Install

```bash
go install github.com/provabl/qualify/cmd/qualify@latest
```

Requires Go 1.24+ and a PostgreSQL database (for progress tracking and IAM tag writes).

---

## Quick start

```bash
# 1. Start the backend (PostgreSQL required)
make docker-up          # starts PostgreSQL on localhost:5433
make build-backend      # builds ./bin/qualify-backend
DB_HOST=localhost DB_PORT=5433 DB_USER=qualify \
  DB_PASSWORD=qualify_dev_password DB_NAME=qualify \
  ./bin/qualify-backend &

# 2. Run migrations
qualify lab setup       # runs golang-migrate up

# 3. Register your IAM role for tag writes
qualify lab register-role --role-arn arn:aws:iam::123456789012:role/ResearchRole

# 4. Check what training is required
qualify train required --framework cmmc-level-2

# 5. Start a module
qualify train start cui-fundamentals

# 6. Check status
qualify train status
```

---

## Architecture

```
qualify CLI                 qualify backend (HTTP)     attest Cedar PDP
─────────────────           ──────────────────────     ─────────────────
train start                 POST /api/training/         evaluates:
  ├── reads module            complete                    attest:cui-training
  ├── runs quiz             GET /api/training/            attest:cui-training-expiry
  ├── svc.CompleteModule      progress                    (written by qualify)
  └── writes IAM tag        GET /api/dashboard/
                              stats
qualify agent (local)
─────────────────────
  intercepts AWS CLI
  checks training tags
  blocks unqualified ops
```

### Key packages

| Package | Purpose |
|---|---|
| `cmd/qualify/cmd/` | CLI commands (train, lab, onboard, s3, credentials) |
| `internal/training/` | Module storage, quiz scoring, completion records |
| `internal/localaudit/` | JSONL audit log at `~/.qualify/audit.log` |
| `cmd/ark-backend/` | HTTP API for dashboard and training progress |
| `cmd/ark-agent/` | Local agent intercepting AWS operations |
| `web/` | React + Radix UI + Tailwind dashboard (TypeScript, Vite, optional) |

### Integration with attest

qualify writes IAM tags in the format `attest:<training-id> = true` with a companion `attest:<training-id>-expiry = <ISO8601>`. attest's Cedar PDP reads these during access evaluation. Tags are written to the role registered via `qualify lab register-role`.

### Integration with ground

ground's `external_services` config declares that qualify is deployed. ground exports this to `ground-meta.json`; attest reads it to know that training-gated access is in effect.

---

## Configuration

qualify reads database connection settings from environment variables:

| Variable | Default | Description |
|---|---|---|
| `DB_HOST` | `localhost` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `qualify` | Database user |
| `DB_PASSWORD` | *(required)* | Database password |
| `DB_NAME` | `qualify` | Database name |
| `DB_SSLMODE` | `disable` | SSL mode (`disable`, `require`, `verify-full`) |
| `MIGRATIONS_PATH` | `./migrations` | Path to migration files |
| `PORT` | `8080` | Backend HTTP port |

Set these in your shell or use a `.env` file for local development.

---

## Local audit log

Every training event is written to `~/.qualify/audit.log` in JSONL format regardless of backend availability:

```json
{"ts":"2026-04-30T14:22:00Z","event":"module_completed","user":"alice@example.edu","module":"cui-fundamentals","details":{"score":80}}
{"ts":"2026-04-30T14:22:01Z","event":"iam_tag_written","user":"alice@example.edu","module":"cui-fundamentals","details":{"tag":"attest:cui-training"}}
```

---

## Development

```bash
git clone https://github.com/provabl/qualify
cd qualify

# Install tools
make install-tools

# Run all checks (fmt, vet, staticcheck, short tests)
make check

# Run full test suite
make test

# Start dev environment (PostgreSQL)
make docker-up

# Build all binaries
make build
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full development guide.

---

## Suite

| Tool | Role |
|---|---|
| **[ground](https://ground.provabl.dev)** | Deploys the AWS org foundation (OUs, networking, logging, Identity Center) |
| **[attest](https://attest.provabl.dev)** | Compiles compliance frameworks → SCPs, Cedar policies, Config rules |
| **qualify** | Training and access gating for researchers |
| **[vet](https://vet.provabl.dev)** | Software supply chain verification (sign, verify, sbom, gate) |

---

## Open-core model

The CLI, backend API, training engine, IAM tag integration, and basic web dashboard are **open source** (Apache 2.0, always free).

The commercial tier (**qualify Cloud**) adds expert-validated compliance content packs, advanced web dashboard, SSO/LDAP integration, multi-institution management, and compliance report generation.

See [COMMERCIAL.md](COMMERCIAL.md) for the full boundary.

---

## License

Apache 2.0. Copyright 2026 Scott Friedman.
