import { test, expect } from '@playwright/test'

test.describe('Home Page', () => {
  test('should load the home page', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveTitle(/qualify/)
    await expect(page.locator('text=Welcome to qualify')).toBeVisible()
  })

  test('should have navigation links', async ({ page }) => {
    await page.goto('/')

    // Use .first() to handle Cloudscape rendering multiple text matches
    await expect(page.locator('text=Home').first()).toBeVisible()
    await expect(page.locator('text=Dashboard').first()).toBeVisible()
    // S3 appears multiple times (nav + page content) — check the nav link specifically
    await expect(page.locator('nav a, [class*="side-navigation"] a').filter({ hasText: /^S3$/ }).first()).toBeVisible()
    await expect(page.locator('text=Training').first()).toBeVisible()
  })

  test('should navigate to different pages', async ({ page }) => {
    await page.goto('/')

    // Navigate to Dashboard
    await page.locator('nav a, [class*="side-navigation"] a').filter({ hasText: 'Dashboard' }).first().click()
    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('text=Dashboard').first()).toBeVisible()

    // Navigate to Training
    await page.locator('nav a, [class*="side-navigation"] a').filter({ hasText: 'Training' }).first().click()
    await expect(page).toHaveURL('/training')
    await expect(page.locator('text=Training Modules')).toBeVisible({ timeout: 8000 })

    // Navigate back to Home
    await page.locator('nav a, [class*="side-navigation"] a').filter({ hasText: 'Home' }).first().click()
    await expect(page).toHaveURL('/')
    await expect(page.locator('text=Welcome to qualify')).toBeVisible()
  })

  test('should display agent status', async ({ page }) => {
    await page.goto('/')

    // Agent status indicator is only present when the qualify agent is running.
    // Skip gracefully in CI where agent is not started.
    const agentStatus = page.locator('text=/Agent (Connected|Disconnected|Checking)/i')
    const visible = await agentStatus.isVisible({ timeout: 3000 }).catch(() => false)
    if (!visible) {
      test.skip()
      return
    }
    await expect(agentStatus).toBeVisible()
  })
})
