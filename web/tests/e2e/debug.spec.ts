import { test, expect } from '@playwright/test'

test('check home page loads and capture errors', async ({ page }) => {
  const consoleMessages: string[] = []
  const errors: string[] = []

  // Capture console messages
  page.on('console', msg => {
    consoleMessages.push(`[${msg.type()}] ${msg.text()}`)
  })

  // Capture page errors
  page.on('pageerror', error => {
    errors.push(error.message)
  })

  // Navigate to home page
  await page.goto('/')

  // Wait a bit for any async errors
  await page.waitForTimeout(2000)

  // Log what we found
  console.log('\n=== Console Messages ===')
  consoleMessages.forEach(msg => console.log(msg))

  console.log('\n=== Errors ===')
  if (errors.length > 0) {
    errors.forEach(err => console.log(err))
  } else {
    console.log('No errors found')
  }

  // Take a screenshot
  await page.screenshot({ path: 'debug-screenshot.png', fullPage: true })
  console.log('\nScreenshot saved to debug-screenshot.png')

  // Check if page has content
  const bodyText = await page.textContent('body')
  console.log('\n=== Body Text Length ===')
  console.log(bodyText?.length || 0)
})
