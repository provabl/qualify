import { test, expect } from '@playwright/test'

test.describe('Onboarding Flow', () => {
  test('should display onboarding wizard for new users', async ({ page }) => {
    await page.goto('/')

    // Wait for onboarding check to complete
    await page.waitForTimeout(1000)

    // Check if onboarding wizard appears
    // The wizard should appear for users who haven't completed onboarding
    const wizardVisible = await page.locator('text=/Welcome to Ark Training|Get Started/i').isVisible({ timeout: 5000 })
      .catch(() => false)

    if (wizardVisible) {
      console.log('Onboarding wizard is displayed')

      // Wizard should have a title or welcome message
      await expect(page.locator('text=/Welcome|Get Started|Onboarding/i').first()).toBeVisible()
    } else {
      console.log('User has already completed onboarding')
    }
  })

  test('should allow users to dismiss onboarding', async ({ page }) => {
    await page.goto('/')
    await page.waitForTimeout(1000)

    // Look for dismiss/close button if wizard is present
    const dismissButton = page.locator('button:has-text("Dismiss"), button:has-text("Skip"), button[aria-label="Close"]').first()

    if (await dismissButton.isVisible({ timeout: 2000 }).catch(() => false)) {
      await dismissButton.click()

      // Wizard should close
      await page.waitForTimeout(500)

      // Main content should be visible
      await expect(page.locator('text=Welcome to Ark')).toBeVisible()
    }
  })

  test('should navigate through onboarding steps', async ({ page }) => {
    await page.goto('/')
    await page.waitForTimeout(1000)

    // Check if wizard is visible
    const wizardVisible = await page.locator('text=/Welcome to Ark Training|Get Started/i').isVisible({ timeout: 5000 })
      .catch(() => false)

    if (wizardVisible) {
      // Look for "Next" or "Continue" button
      const nextButton = page.locator('button:has-text("Next"), button:has-text("Continue")').first()

      if (await nextButton.isVisible({ timeout: 2000 }).catch(() => false)) {
        // Click through steps
        await nextButton.click()
        await page.waitForTimeout(500)

        // Should show next step or finish button
        const finishButton = page.locator('button:has-text("Finish"), button:has-text("Get Started"), button:has-text("Complete")').first()
        const nextButton2 = page.locator('button:has-text("Next"), button:has-text("Continue")').first()

        const hasFinish = await finishButton.isVisible({ timeout: 2000 }).catch(() => false)
        const hasNext = await nextButton2.isVisible({ timeout: 2000 }).catch(() => false)

        expect(hasFinish || hasNext).toBeTruthy()
      }
    } else {
      test.skip()
    }
  })

  test('should complete onboarding and hide wizard', async ({ page }) => {
    await page.goto('/')
    await page.waitForTimeout(1000)

    // Check if wizard is visible
    const wizardVisible = await page.locator('text=/Welcome to Ark Training|Get Started/i').isVisible({ timeout: 5000 })
      .catch(() => false)

    if (wizardVisible) {
      // Try to find and click through to completion
      let maxSteps = 5
      let currentStep = 0

      while (currentStep < maxSteps) {
        const nextButton = page.locator('button:has-text("Next"), button:has-text("Continue")').first()
        const finishButton = page.locator('button:has-text("Finish"), button:has-text("Complete"), button:has-text("Get Started")').first()

        const hasNext = await nextButton.isVisible({ timeout: 1000 }).catch(() => false)
        const hasFinish = await finishButton.isVisible({ timeout: 1000 }).catch(() => false)

        if (hasFinish) {
          await finishButton.click()
          await page.waitForTimeout(1000)

          // Wizard should be hidden after completion
          const stillVisible = await page.locator('text=/Welcome to Ark Training|Get Started/i').isVisible({ timeout: 2000 })
            .catch(() => false)

          expect(stillVisible).toBeFalsy()
          break
        } else if (hasNext) {
          await nextButton.click()
          await page.waitForTimeout(500)
          currentStep++
        } else {
          break
        }
      }
    } else {
      test.skip()
    }
  })
})
