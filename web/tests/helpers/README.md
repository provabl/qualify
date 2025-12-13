# Test Helpers

Utilities and mock data generators for testing.

## Mock Data (`mockData.ts`)

### Quiz Testing

```typescript
import { createMockQuizQuestions, createMockQuizResponse } from './helpers/mockData'

// Create 5 quiz questions
const questions = createMockQuizQuestions(5)

// Create a passing quiz response
const passResponse = createMockQuizResponse({
  score: 80,
  totalQuestions: 3,
  correctAnswers: 2,
  passed: true,
})

// Create a failing quiz response
const failResponse = createMockQuizResponse({
  score: 50,
  totalQuestions: 3,
  correctAnswers: 1,
  passed: false,
})
```

### Dashboard Testing

```typescript
import {
  createMockDashboardStats,
  createEmptyDashboardStats
} from './helpers/mockData'

// Create dashboard with custom stats
const stats = createMockDashboardStats({
  completed: 3,
  inProgress: 1,
  notStarted: 0,
  averageScore: 90,
  unlockedOps: ['s3:CreateBucket', 's3:ListBuckets'],
})

// Create empty dashboard for new user
const emptyStats = createEmptyDashboardStats('user-123')
```

### Training Modules

```typescript
import {
  createMockTrainingModule,
  createMockTrainingModules
} from './helpers/mockData'

// Create single module
const module = createMockTrainingModule({
  name: 's3-basics',
  title: 'S3 Basics',
  status: 'completed',
  score: 85,
})

// Create multiple modules
const modules = createMockTrainingModules(4)
```

### Activity Items

```typescript
import { createMockActivityItems } from './helpers/mockData'

// Create 10 activity items
const activities = createMockActivityItems(10)
```

## Test Utilities (`testUtils.tsx`)

### Component Rendering

```typescript
import { renderWithRouter } from './helpers/testUtils'

// Render component with router context
const { getByText } = renderWithRouter(<Dashboard />)
```

### Mock API Calls

```typescript
import {
  mockApiSuccess,
  mockApiError,
  mockApiDelayed,
  mockApiPending
} from './helpers/testUtils'

// Mock successful API call
vi.mocked(agentService.getDashboardStats).mockImplementation(
  mockApiSuccess(mockDashboardStats)
)

// Mock API error
vi.mocked(agentService.getDashboardStats).mockImplementation(
  mockApiError(new Error('Network error'))
)

// Mock delayed response (loading state)
vi.mocked(agentService.getDashboardStats).mockImplementation(
  mockApiDelayed(mockDashboardStats, 2000)
)

// Mock pending (never resolves)
vi.mocked(agentService.getDashboardStats).mockImplementation(
  mockApiPending()
)
```

### Mock Agent Service

```typescript
import { createMockAgentService } from './helpers/testUtils'

const mockService = createMockAgentService()

// Configure specific methods
mockService.getDashboardStats.mockResolvedValue(mockStats)
mockService.listBuckets.mockRejectedValue(new Error('Not authorized'))
```

### Router Testing

```typescript
import { createMockNavigate, mockReactRouter } from './helpers/testUtils'

const mockNavigate = createMockNavigate()
mockReactRouter(mockNavigate)

// Test navigation
button.click()
expect(mockNavigate).toHaveBeenCalledWith('/training')
```

### Console Suppression

```typescript
import { suppressConsoleError } from './helpers/testUtils'

it('handles error gracefully', async () => {
  await suppressConsoleError(async () => {
    // Test that logs expected errors
    // Errors won't clutter test output
  })
})
```

## Example Test

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { screen, waitFor } from '@testing-library/react'
import { renderWithRouter, mockApiSuccess } from './helpers/testUtils'
import { createMockDashboardStats } from './helpers/mockData'
import Dashboard from '@/views/Dashboard'
import { agentService } from '@/services/agent'

vi.mock('@/services/agent')

describe('Dashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('displays dashboard stats', async () => {
    const mockStats = createMockDashboardStats({
      completed: 2,
      inProgress: 1,
    })

    vi.mocked(agentService.getDashboardStats).mockImplementation(
      mockApiSuccess(mockStats)
    )

    renderWithRouter(<Dashboard />)

    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument()
    })

    expect(screen.getByText('Total Modules')).toBeInTheDocument()
  })
})
```
