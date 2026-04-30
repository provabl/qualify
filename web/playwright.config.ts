import { defineConfig, devices } from '@playwright/test'

// Use unique port 5174 for Playwright testing to avoid conflicts
const TEST_PORT = process.env.TEST_PORT || '5174'
const BASE_URL = `http://127.0.0.1:${TEST_PORT}`

export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 2 : undefined,
  reporter: 'html',
  timeout: process.env.CI ? 30000 : 60000,
  use: {
    baseURL: BASE_URL,
    trace: 'on-first-retry',
    actionTimeout: process.env.CI ? 8000 : 15000,
  },

  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },

    {
      name: 'firefox',
      use: { ...devices['Desktop Firefox'] },
    },

    {
      name: 'webkit',
      use: { ...devices['Desktop Safari'] },
    },
  ],

  // Start dedicated dev server on port 5174 for testing
  // This avoids conflicts with dev servers from other projects
  webServer: {
    command: `VITE_PORT=${TEST_PORT} npm run dev`,
    url: BASE_URL,
    reuseExistingServer: !process.env.CI,
    timeout: 120000,
  },
})
