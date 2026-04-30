# Contributing to qualify

## Development setup

**Requirements:** Go 1.24+, Node 20+, Docker, PostgreSQL 15+ (or use `make docker-up`)

```bash
git clone https://github.com/provabl/qualify
cd qualify
make install-tools    # installs staticcheck, golangci-lint
make docker-up        # PostgreSQL on localhost:5433
make build            # builds qualify, qualify-backend, qualify-agent → bin/
make check            # fmt + vet + staticcheck + short tests (fast)
make test             # full test suite with coverage
```

## Project layout

```
cmd/
  qualify/           CLI entry point
    cmd/             cobra subcommands: train, lab, onboard, s3, credentials
  ark-backend/       HTTP API server
  ark-agent/         Local AWS operation interceptor
internal/
  training/          Module storage, quiz scoring, completion records
  localaudit/        JSONL audit log (~/.qualify/audit.log)
  audit/             DB-backed audit service (used by backend)
  database/          DB connection helpers
migrations/          golang-migrate SQL files (numbered 000001+)
web/                 React + Radix UI + Tailwind dashboard (TypeScript, Vite)
scripts/
  init-db.sql        PostgreSQL init (extensions only — migrations handle schema)
```

## Adding a training module

1. Write a new migration in `migrations/00000N_<name>.up.sql`:

```sql
INSERT INTO training_modules (name, title, description, category, difficulty,
  estimated_minutes, required_for_frameworks, content) VALUES (
  'my-module',
  'My Module Title',
  'One-sentence description.',
  'compliance',
  'intermediate',
  20,
  '["framework-id"]',
  '{
    "sections": [
      {"title": "Section 1", "content": "Content here. **Bold** and `code` work."},
      {"title": "Section 2", "content": "More content."}
    ],
    "quiz": [
      {
        "question": "What is the answer?",
        "options": ["Option A", "Option B", "Option C", "Option D"],
        "correct": 1,
        "explanation": "Option B is correct because..."
      }
    ],
    "passing_score": 80
  }'
);
```

2. Write the down migration (`000N_<name>.down.sql`):

```sql
DELETE FROM training_modules WHERE name = 'my-module';
```

3. Add the module to `training.moduleTagMap` (in `internal/training/tags.go`) and `moduleUnlocks` in `cmd/qualify/cmd/train.go`.

4. Add the module ID to the relevant framework mapping in migration `000007`.

## Running tests

```bash
# Unit tests only (no DB required)
go test -short ./...

# Full tests (requires DB via docker-up)
make test

# Specific package
go test ./internal/training/... -v

# Frontend
cd web && npm run test:unit:run
```

## Environment variables for tests

```bash
export DB_HOST=localhost
export DB_PORT=5433
export DB_USER=qualify
export DB_PASSWORD=qualify_dev_password
export DB_NAME=qualify
```

## Web frontend environment variables

```bash
# Override backend URL (default: http://127.0.0.1:8081)
export VITE_BACKEND_URL=http://localhost:8081

# Override agent URL (default: http://127.0.0.1:8737)
export VITE_AGENT_URL=http://localhost:8737
```

Set these in `web/.env.local` for persistent overrides (`.env.local` is gitignored).

## Commit style

- Imperative mood: "Add feature" not "Added feature"
- Reference issues: `Fixes #N` or `Closes #N`
- Keep subject under 72 characters

## Code conventions

- No `init()` functions, no global mutable state
- Errors returned, not logged-and-continued
- Table-driven tests preferred
- `go vet ./...` and `go test ./...` must pass before push
- `gofmt -w .` before committing (pre-commit hook enforces this)
