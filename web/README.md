# qualify web dashboard

React + Tailwind CSS dashboard for the qualify compliance training platform.

## Stack

- **React 18** + TypeScript + Vite 7
- **Tailwind CSS v4** via `@tailwindcss/vite` plugin (no config file)
- **lucide-react** for icons
- **Vitest** for unit tests, **Playwright** for E2E

## Development

```bash
npm install
npm run dev          # starts Vite dev server on localhost:5173
```

Environment overrides (set in `.env.local`, gitignored):

```bash
VITE_BACKEND_URL=http://localhost:8081   # qualify backend (default: 127.0.0.1:8081)
VITE_AGENT_URL=http://localhost:8737     # qualify agent   (default: 127.0.0.1:8737)
VITE_PORT=5173                           # dev server port (default: 5173)
```

The backend must be running for the dashboard to load data. Start it with:

```bash
cd ..   # qualify repo root
make docker-up          # PostgreSQL on :5433
make build-backend      # builds ./bin/qualify-backend
PORT=8081 ./bin/qualify-backend &
```

## Testing

```bash
npm run test:unit:run        # Vitest unit tests (no backend needed)
npm run test:unit:coverage   # with coverage report
npm run test:e2e             # Playwright E2E (requires backend + agent running)
```

## Structure

```
src/
  views/           Dashboard, Training, TrainingModule, S3, Home
  components/
    training/      Quiz, TrainingGate
    onboarding/    OnboardingWizard
    common/        AgentStatus
  services/        agent.ts — HTTP client for backend + agent APIs
  types/           api.ts — TypeScript interfaces
  contexts/        AgentContext — agent connection status
  lib/             utils.ts — cn() helper (clsx + tailwind-merge)
```

See the root [CONTRIBUTING.md](../CONTRIBUTING.md) for the full development guide.
