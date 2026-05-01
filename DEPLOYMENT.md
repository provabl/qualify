# Deploying qualify

qualify runs **inside your institution's AWS account** — no data leaves your environment. This guide covers deploying the qualify backend and dashboard in a production-ready configuration.

---

## Quick start (Docker Compose)

The fastest path to a running qualify instance:

```bash
# 1. Copy environment template
cp compose.env.example .env

# 2. Edit .env — fill in DB_PASSWORD, JWT_SECRET, and optionally LICENSE_KEY
nano .env

# 3. Start PostgreSQL + backend
docker compose --env-file .env up -d

# 4. Verify
curl http://localhost:8081/health
# → {"status":"healthy","version":"..."}

# 5. Get a dev token (dev mode only)
curl http://localhost:8081/api/auth/dev-token
```

---

## Environment variables

### Required

| Variable | Description |
|---|---|
| `DB_HOST` | PostgreSQL host (default: `localhost`) |
| `DB_PORT` | PostgreSQL port (default: `5432`) |
| `DB_USER` | PostgreSQL user (default: `qualify`) |
| `DB_PASSWORD` | PostgreSQL password — **set a strong value** |
| `DB_NAME` | PostgreSQL database name (default: `qualify`) |
| `DB_SSLMODE` | SSL mode: `disable` (dev) / `require` / `verify-full` (prod) |

### Authentication

| Variable | Description |
|---|---|
| `AUTH_DEV_MODE` | `true` = disable JWT checks, inject a dev user. **Never use in production.** |
| `AUTH_DEV_USER_ID` | UUID injected as the dev user (default: `00000000-0000-0000-0000-000000000001`) |
| `AUTH_DEV_EMAIL` | Email injected in dev mode (default: `dev@example.edu`) |
| `JWT_SECRET` | HMAC-SHA256 signing secret — minimum 32 bytes. Required in production. Generate with: `openssl rand -hex 32` |
| `JWT_EXPIRY` | Token lifetime (default: `24h`). Accepts Go duration strings: `1h`, `8h`, `24h`. |

### License

| Variable | Description |
|---|---|
| `LICENSE_KEY` | License key from [provabl.co](https://provabl.co). Empty = community tier (basic modules, no SSO). |
| `LICENSE_ENDPOINT` | Provabl licensing API URL (default: `https://licensing.provabl.co/api/v1/validate`) |
| `LICENSE_CACHE_TTL` | How long to cache a valid license response (default: `24h`). Useful for air-gapped deployments. |

### Backend

| Variable | Description |
|---|---|
| `PORT` | HTTP port the backend listens on (default: `8080`) |
| `MIGRATIONS_PATH` | Path to SQL migration files (default: `./migrations`) |
| `LOG_LEVEL` | Log level: `debug` / `info` / `warn` / `error` (default: `info`) |

### Web frontend (Vite env vars)

Set in `web/.env.local` for development, or pass to the Docker build:

| Variable | Description |
|---|---|
| `VITE_BACKEND_URL` | qualify backend URL from the browser (default: `http://127.0.0.1:8081`) |
| `VITE_AGENT_URL` | qualify agent URL from the browser (default: `http://127.0.0.1:8737`) |

---

## Production deployment checklist

- [ ] `AUTH_DEV_MODE` is **not** set (or set to `false`)
- [ ] `JWT_SECRET` is set to a randomly-generated 32+ byte secret
- [ ] `DB_SSLMODE=require` (or `verify-full` with a CA cert)
- [ ] `DB_PASSWORD` is not the default `qualify_dev_password`
- [ ] PostgreSQL is not exposed on a public network interface
- [ ] Backend is behind a reverse proxy (nginx/ALB) with TLS
- [ ] `LICENSE_KEY` is set (or documented that you're on community tier)
- [ ] Backup and restore procedures tested

---

## Authentication flow

### Development (current)

```
Browser loads qualify dashboard
  → GET /api/auth/dev-token (only available when AUTH_DEV_MODE=true)
  → Returns a signed JWT for AUTH_DEV_USER_ID
  → Frontend stores JWT in sessionStorage
  → All subsequent /api/* requests include Authorization: Bearer <token>
```

### Production v0.2.0 (JWT)

```
Browser loads qualify dashboard
  → No token in sessionStorage → redirect to /login
  → POST /api/auth/login with credentials
  → Backend verifies, issues JWT signed with JWT_SECRET
  → Frontend stores JWT, sends on all requests
  → JWT expires after JWT_EXPIRY; frontend prompts re-login
```

### Production v0.3.0 (OIDC/SAML — planned)

```
Browser → OIDC redirect to institutional IdP (Shibboleth, Azure AD, Okta)
  → IdP issues OIDC token after MFA
  → qualify backend validates OIDC token, maps to user
  → Issues internal JWT with mapped attributes (lab_id, role, institution)
```

---

## Content packs

Training content is loaded via numbered PostgreSQL migrations. The open-source repository includes foundation modules (security-awareness, CUI, HIPAA, FERPA, etc.).

Commercial content packs (expert-validated framework-specific modules) are distributed as signed SQL migration bundles from [provabl.co](https://provabl.co):

```bash
# Download and verify a content pack
curl -O https://content.provabl.co/packs/fips-140-v1.2.tar.gz
cosign verify-blob --certificate-identity-regexp 'provabl.co' \
  --certificate-oidc-issuer https://accounts.google.com \
  fips-140-v1.2.tar.gz

# Extract migrations
tar xzf fips-140-v1.2.tar.gz -C ./migrations/

# Restart backend (migrations run automatically on startup)
docker compose restart backend
```

The backend's license validator checks which packs are authorized for your license tier at startup.

---

## Kubernetes deployment

Reference manifests are in `kubernetes/`. The kustomize overlay includes namespace, ConfigMap, Deployment, and Service resources.

```bash
# 1. Create your Secret (never commit this file)
cp kubernetes/secret.yaml.example kubernetes/secret.yaml
# Edit kubernetes/secret.yaml — fill in DB_PASSWORD, JWT_SECRET, LICENSE_KEY

# 2. Optionally edit ConfigMap defaults
#    kubernetes/configmap.yaml — DB_HOST, AUTH_DEV_MODE, LOG_LEVEL, etc.

# 3. Apply
kubectl apply -k kubernetes/
kubectl rollout status deployment/qualify-backend -n qualify

# 4. Verify
kubectl get pods -n qualify
kubectl logs -n qualify deployment/qualify-backend | tail -20
```

To pin a specific release version, add to `kubernetes/kustomization.yaml`:
```yaml
images:
  - name: ghcr.io/provabl/qualify-backend
    newTag: v0.1.2
```

---

## Backup and restore

```bash
# Backup (from inside the Docker network)
docker compose exec postgres pg_dump -U qualify qualify > qualify-backup-$(date +%Y%m%d).sql

# Restore
docker compose exec -T postgres psql -U qualify qualify < qualify-backup-20260501.sql
```

---

## Upgrading

```bash
# Pull latest images
docker compose pull

# Restart (migrations run automatically on startup)
docker compose up -d

# Verify migration version
docker compose exec backend ./qualify-backend --version
```

---

## Troubleshooting

**Backend won't start — "failed to connect to database"**
- Check `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`
- Verify PostgreSQL is running: `docker compose ps postgres`

**"license validation failed — running as community tier"**
- Check `LICENSE_KEY` is correct
- Verify network connectivity to `licensing.provabl.co`
- Increase `LICENSE_CACHE_TTL` for air-gapped environments

**"invalid or expired token" on API calls**
- Token may have expired (default: 24h). Clear sessionStorage and reload.
- If using dev mode: verify `AUTH_DEV_MODE=true` is set on the backend.

**CORS errors in browser**
- Verify `VITE_BACKEND_URL` in `web/.env.local` matches where the backend is running
- The backend allows `localhost:5173` and `localhost:5174` by default
