/**
 * Test utilities for component testing
 */

import { ReactElement } from 'react'
import { render, RenderOptions } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import { vi } from 'vitest'

/**
 * Render component with router wrapper
 */
export function renderWithRouter(
  ui: ReactElement,
  options?: Omit<RenderOptions, 'wrapper'>
) {
  function Wrapper({ children }: { children: React.ReactNode }) {
    return <BrowserRouter>{children}</BrowserRouter>
  }

  return render(ui, { wrapper: Wrapper, ...options })
}

/**
 * Create mock navigate function for router tests
 */
export function createMockNavigate() {
  return vi.fn()
}

/**
 * Mock react-router-dom with custom navigate function
 */
export function mockReactRouter(mockNavigate: ReturnType<typeof vi.fn>) {
  vi.mock('react-router-dom', async () => {
    const actual = await vi.importActual('react-router-dom')
    return {
      ...actual,
      useNavigate: () => mockNavigate,
    }
  })
}

/**
 * Wait for async operations
 */
export function waitForAsync(ms: number = 100) {
  return new Promise(resolve => setTimeout(resolve, ms))
}

/**
 * Create mock API error
 */
export function createMockError(message: string = 'Network error') {
  return new Error(message)
}

/**
 * Mock successful API call
 */
export function mockApiSuccess<T>(data: T) {
  return vi.fn().mockResolvedValue(data)
}

/**
 * Mock failed API call
 */
export function mockApiError(error: Error = new Error('Network error')) {
  return vi.fn().mockRejectedValue(error)
}

/**
 * Mock API call with delay
 */
export function mockApiDelayed<T>(data: T, delay: number = 1000) {
  return vi.fn().mockImplementation(
    () =>
      new Promise(resolve => {
        setTimeout(() => resolve(data), delay)
      })
  )
}

/**
 * Mock API call that never resolves (for testing loading states)
 */
export function mockApiPending() {
  return vi.fn().mockImplementation(() => new Promise(() => {}))
}

/**
 * Create mock agent service
 */
export function createMockAgentService() {
  return {
    getDashboardStats: vi.fn(),
    getUserProfile: vi.fn(),
    updateUserPreferences: vi.fn(),
    listTrainingModules: vi.fn(),
    getTrainingModule: vi.fn(),
    submitQuizAnswers: vi.fn(),
    getUserActivity: vi.fn(),
    listBuckets: vi.fn(),
    createBucket: vi.fn(),
    deleteBucket: vi.fn(),
  }
}

/**
 * Reset all mocks
 */
export function resetAllMocks() {
  vi.clearAllMocks()
  vi.resetAllMocks()
}

/**
 * Mock console methods to suppress expected errors in tests
 */
export function mockConsole() {
  const originalError = console.error
  const originalWarn = console.warn

  beforeEach(() => {
    console.error = vi.fn()
    console.warn = vi.fn()
  })

  afterEach(() => {
    console.error = originalError
    console.warn = originalWarn
  })
}

/**
 * Suppress console errors in a specific test
 */
export function suppressConsoleError(test: () => void | Promise<void>) {
  const originalError = console.error
  console.error = vi.fn()

  try {
    return test()
  } finally {
    console.error = originalError
  }
}
