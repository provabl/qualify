import { test, expect } from '@playwright/test'

test.describe('Home Page', () => {
  test('should load the home page', async ({ page }) => {
    await page.goto('/')
    await expect(page).toHaveTitle(/qualify/)
    await expect(page.locator('text=Welcome to qualify')).toBeVisible()
  })

  test('should have navigation links', async ({ page }) => {
    await page.goto('/')

    // Use role-based selectors — Cloudscape SideNavigation renders links as <a> elements
    await expect(page.getByRole('link', { name: 'Home' }).first()).toBeVisible()
    await expect(page.getByRole('link', { name: 'Dashboard' }).first()).toBeVisible()
    await expect(page.getByRole('link', { name: 'S3' }).first()).toBeVisible()
    await expect(page.getByRole('link', { name: 'Training' }).first()).toBeVisible()
  })

  test('should navigate to different pages', async ({ page }) => {
    await page.goto('/')

    // Use role-based link clicks — more reliable than text= with Cloudscape
    await page.getByRole('link', { name: 'Dashboard' }).first().click()
    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('text=Dashboard').first()).toBeVisible()

    await page.getByRole('link', { name: 'Training' }).first().click()
    await expect(page).toHaveURL('/training')
    await expect(page.locator('text=Training Modules').first()).toBeVisible({ timeout: 8000 })

    await page.getByRole('link', { name: 'Home' }).first().click()
    await expect(page).toHaveURL('/')
    await expect(page.locator('text=Welcome to qualify')).toBeVisible()
  })

  test('should display agent status', async ({ page }) => {
    // The qualify agent runs locally on :8737 and is not started in CI.
    test.skip(!!process.env.CI, 'qualify agent not started in CI environment')

    await page.goto('/')
    await expect(page.locator('text=/Agent (Connected|Disconnected|Checking)/i').first()).toBeVisible({ timeout: 5000 })
  })
})
