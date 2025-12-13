# Testing Guide

Comprehensive testing infrastructure for Ark.

## Quick Start

```bash
# Run all tests (backend + frontend)
make check           # Fast pre-commit checks
make test            # Backend unit tests with coverage
make web-test-unit   # Frontend unit tests
make web-test-e2e    # Frontend E2E tests (requires backend)
```

## Backend Testing

### Unit Tests

Located in `*_test.go` files alongside source code.

```bash
# Run all backend tests
make test

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/training/...

# Run short tests (fast, used in pre-commit)
go test -short ./...
```

### Test Structure

```go
func TestServiceMethod(t *testing.T) {
    // Table-driven tests
    tests := []struct {
        name    string
        input   Input
        want    Output
        wantErr bool
    }{
        {name: "success case", input: ..., want: ...},
        {name: "error case", input: ..., wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### Test Helpers

Use `internal/training/testhelpers.go` for database mocking:

```go
import "github.com/scttfrdmn/ark/internal/training"

// Create test helper
sqlDB, mock, _ := sqlmock.New()
db := &database.DB{DB: sqlDB}
helper := training.NewTestHelper(mock)

// Mock database queries
helper.MockModuleQuery("mod-1", mockContent)
helper.MockUserProgressUpdate("user-1", "mod-1")

// Create mock data
content := training.CreateMockQuizContent(3)
submission := training.CreateMockQuizSubmission("user-1", 3, 2)
```

### Integration Tests

Tag integration tests with `// +build integration`:

```bash
# Run integration tests (requires LocalStack)
make integration
```

## Frontend Testing

### Unit Tests

Located in `web/tests/unit/*.test.tsx`.

```bash
# Run unit tests (watch mode)
cd web && npm run test:unit

# Run once (CI mode)
make web-test-unit

# Generate coverage report
make web-coverage
```

### Test Structure

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { renderWithRouter } from '@/tests/helpers/testUtils'
import { createMockDashboardStats } from '@/tests/helpers/mockData'

describe('Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders correctly', async () => {
    const mockData = createMockDashboardStats()
    vi.mocked(agentService.getData).mockResolvedValue(mockData)

    renderWithRouter(<Component />)

    await waitFor(() => {
      expect(screen.getByText('Expected')).toBeInTheDocument()
    })
  })
})
```

### Test Helpers

Use helpers from `web/tests/helpers/`:

```typescript
// Mock data generation
import {
  createMockQuizQuestions,
  createMockDashboardStats,
  createMockTrainingModules,
} from '@/tests/helpers/mockData'

// Test utilities
import {
  renderWithRouter,
  mockApiSuccess,
  mockApiError,
  createMockAgentService,
} from '@/tests/helpers/testUtils'

// Example usage
const stats = createMockDashboardStats({ completed: 3, inProgress: 1 })
const { getByText } = renderWithRouter(<Dashboard />)
vi.mocked(service.get).mockImplementation(mockApiSuccess(stats))
```

### E2E Tests

Located in `web/tests/e2e/*.spec.ts`.

```bash
# Run E2E tests (requires backend + frontend)
make web-test-e2e

# Debug mode
cd web && npm run test:e2e:debug

# UI mode
cd web && npm run test:e2e:ui
```

**E2E Test Suites:**
- `onboarding.spec.ts` - Onboarding wizard flow (4 tests)
- `training-flow.spec.ts` - Training and quiz flows (8 tests)
- `training-gate.spec.ts` - Operation locking (7 tests)
- `home.spec.ts` - Basic navigation (4 tests)

### E2E Test Setup

E2E tests require:
1. Backend server running on port 8080
2. Database with migrations applied
3. Frontend dev server on port 5174

```bash
# Terminal 1: Start backend
make backend-dev

# Terminal 2: Start frontend (test port)
cd web && npm run dev:test

# Terminal 3: Run E2E tests
cd web && npm run test:e2e
```

## Test Coverage

### Backend

```bash
# Generate coverage report
make test

# View HTML coverage report
go tool cover -html=coverage.out
```

**Coverage Targets:**
- Minimum: 60%
- Target: 80%+

### Frontend

```bash
# Generate coverage report
make web-coverage

# View HTML report
open web/coverage/index.html
```

## CI/CD

### GitHub Actions Workflows

**test.yml** - Comprehensive testing on main/develop:
- Backend tests with coverage
- Frontend unit tests with coverage
- E2E tests with Playwright
- Integration tests with LocalStack

**check.yml** - Quick checks on feature branches:
- Go fmt, vet, staticcheck
- Backend short tests
- Frontend unit tests
- Build verification

### Running Locally

Simulate CI environment:

```bash
# Backend checks (fast)
make check

# Full backend tests
make test

# Frontend tests
make web-test-unit

# Integration tests (requires Docker)
docker-compose -f docker/docker-compose.dev.yml up -d
make integration
```

## Pre-commit Hook

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
set -e

echo "Running pre-commit checks..."

# Backend checks
make check

# Frontend unit tests
cd web && npm run test:unit:run

echo "✓ All checks passed"
```

Make executable:
```bash
chmod +x .git/hooks/pre-commit
```

## Best Practices

### Backend

1. **Table-driven tests** for multiple cases
2. **Use testhelpers** for database mocking
3. **Test error paths** not just happy path
4. **Use t.Helper()** in test utilities
5. **Short flag** for fast tests: `if testing.Short() { t.Skip() }`

### Frontend

1. **Mock external dependencies** (API services)
2. **Use renderWithRouter** for components with routing
3. **waitFor** for async operations
4. **userEvent** for user interactions (not fireEvent)
5. **Query by accessibility** (getByRole, getByLabelText)
6. **Use test helpers** to reduce boilerplate

### E2E

1. **Test user flows** not implementation
2. **Wait for elements** before assertions
3. **Use timeouts** for slow operations
4. **Graceful degradation** if services unavailable
5. **Clean test data** between runs

## Troubleshooting

### Backend Tests Fail

```bash
# Clean and rebuild
make clean
go clean -testcache
make test
```

### Frontend Tests Fail

```bash
# Clear node modules and reinstall
cd web
rm -rf node_modules package-lock.json
npm install
npm run test:unit:run
```

### E2E Tests Fail

```bash
# Check backend is running
curl http://localhost:8080/health

# Check frontend is running
curl http://localhost:5174

# View Playwright report
cd web && npx playwright show-report
```

### Coverage Too Low

1. Check `coverage.out` (backend) or `web/coverage/` (frontend)
2. Identify untested files
3. Add tests for critical paths first
4. Use `// +build integration` for slow tests

## Resources

- [Go Testing](https://go.dev/doc/tutorial/add-a-test)
- [Vitest](https://vitest.dev/)
- [React Testing Library](https://testing-library.com/react)
- [Playwright](https://playwright.dev/)
- [go-sqlmock](https://github.com/DATA-DOG/go-sqlmock)
