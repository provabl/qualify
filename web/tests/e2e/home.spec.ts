import { test, expect } from '@playwright/test'

test.describe('Home Page', () => {
  test('should load the home page', async ({ page }) => {
    await page.goto('/')

    // Check that the page title contains expected text
    await expect(page).toHaveTitle(/Ark/)

    // Check that the main header is present
    await expect(page.locator('text=Welcome to Ark')).toBeVisible()
  })

  test('should have navigation links', async ({ page }) => {
    await page.goto('/')

    // Check for navigation items
    await expect(page.locator('text=Home')).toBeVisible()
    await expect(page.locator('text=Dashboard')).toBeVisible()
    await expect(page.locator('text=S3')).toBeVisible()
    await expect(page.locator('text=Training')).toBeVisible()
  })

  test('should navigate to different pages', async ({ page }) => {
    await page.goto('/')

    // Navigate to Dashboard
    await page.click('text=Dashboard')
    await expect(page).toHaveURL('/dashboard')
    await expect(page.locator('text=Dashboard').first()).toBeVisible()

    // Navigate to S3
    await page.click('text=S3')
    await expect(page).toHaveURL('/s3')
    await expect(page.locator('text=S3 Buckets')).toBeVisible()

    // Navigate to Training
    await page.click('text=Training')
    await expect(page).toHaveURL('/training')
    await expect(page.locator('text=Training Modules')).toBeVisible()

    // Navigate back to Home
    await page.click('text=Home')
    await expect(page).toHaveURL('/')
    await expect(page.locator('text=Welcome to Ark')).toBeVisible()
  })

  test('should display agent status', async ({ page }) => {
    await page.goto('/')

    // Check for agent status indicator
    // It may show as disconnected if agent is not running
    await expect(page.locator('text=/Agent (Connected|Disconnected|Checking)/i')).toBeVisible()
  })
})
