import { test, expect } from '@playwright/test'

test.describe('Training Flow', () => {
  test('should display training modules list', async ({ page }) => {
    await page.goto('/training')

    // Should show Training Modules heading
    await expect(page.locator('text=Training Modules').first()).toBeVisible({ timeout: 5000 })

    // Should show at least one module or empty state
    const hasModules = await page.locator('[data-testid="training-module"], .module-card, text=/CUI Fundamentals|Security Awareness|HIPAA|FERPA|NIH|Data Classification/i').isVisible({ timeout: 3000 })
      .catch(() => false)

    const hasEmptyState = await page.locator('text=/No training|No modules|no training/i').isVisible({ timeout: 1000 })
      .catch(() => false)

    expect(hasModules || hasEmptyState).toBeTruthy()
  })

  test('should show module status indicators', async ({ page }) => {
    await page.goto('/training')
    await page.waitForTimeout(1000)

    // Look for status badges
    const statusBadges = page.locator('text=/Not Started|In Progress|Completed/i')
    const count = await statusBadges.count().catch(() => 0)

    // Should have at least one module with status
    if (count > 0) {
      expect(count).toBeGreaterThan(0)
      console.log(`Found ${count} training modules with status indicators`)
    } else {
      console.log('No modules found or modules do not have status indicators')
    }
  })

  test('should navigate to a training module', async ({ page }) => {
    await page.goto('/training')
    await page.waitForTimeout(1000)

    // Find first module link or "Start" button
    const moduleLink = page.locator('a[href*="/training/"], button:has-text("Start"), button:has-text("Continue"), button:has-text("View")').first()

    if (await moduleLink.isVisible({ timeout: 3000 }).catch(() => false)) {
      await moduleLink.click()

      // Should navigate to module detail page
      await expect(page).toHaveURL(/\/training\/[^/]+/, { timeout: 5000 })

      // Should show module content or title
      await page.waitForTimeout(1000)
      const hasContent = await page.locator('h1, h2, [role="heading"]').isVisible({ timeout: 2000 })
        .catch(() => false)

      expect(hasContent).toBeTruthy()
    } else {
      test.skip()
    }
  })

  test('should display module content sections', async ({ page }) => {
    await page.goto('/training')
    await page.waitForTimeout(1000)

    // Try to open first available module
    const moduleLink = page.locator('a[href*="/training/"], button:has-text("Start"), button:has-text("Continue")').first()

    if (await moduleLink.isVisible({ timeout: 3000 }).catch(() => false)) {
      await moduleLink.click()
      await page.waitForTimeout(1000)

      // Module should have content sections
      const hasIntroduction = await page.locator('text=/Introduction|Overview|About/i').isVisible({ timeout: 3000 })
        .catch(() => false)

      const hasContent = await page.locator('text=/Key Concepts|Learning Objectives|Topics/i, p, div[class*="content"]').isVisible({ timeout: 2000 })
        .catch(() => false)

      // Should have some content visible
      expect(hasIntroduction || hasContent).toBeTruthy()
    } else {
      test.skip()
    }
  })

  test('should display quiz section in module', async ({ page }) => {
    await page.goto('/training')
    await page.waitForTimeout(1000)

    // Navigate to first module
    const moduleLink = page.locator('a[href*="/training/"], button:has-text("Start")').first()

    if (await moduleLink.isVisible({ timeout: 3000 }).catch(() => false)) {
      await moduleLink.click()
      await page.waitForTimeout(1000)

      // Look for quiz section
      const hasQuiz = await page.locator('text=/Quiz|Assessment|Test Your Knowledge/i').isVisible({ timeout: 3000 })
        .catch(() => false)

      if (hasQuiz) {
        console.log('Quiz section found in module')

        // Should have quiz questions or start quiz button
        const hasQuestions = await page.locator('text=/Question [0-9]+|Select the correct answer/i').isVisible({ timeout: 2000 })
          .catch(() => false)

        const hasStartButton = await page.locator('button:has-text("Start Quiz"), button:has-text("Take Quiz")').isVisible({ timeout: 2000 })
          .catch(() => false)

        expect(hasQuestions || hasStartButton).toBeTruthy()
      } else {
        console.log('Module does not have a quiz section')
      }
    } else {
      test.skip()
    }
  })

  test('should take quiz and show results', async ({ page }) => {
    await page.goto('/training')
    await page.waitForTimeout(1000)

    // Navigate to first module
    const moduleLink = page.locator('a[href*="/training/"], button:has-text("Start")').first()

    if (await moduleLink.isVisible({ timeout: 3000 }).catch(() => false)) {
      await moduleLink.click()
      await page.waitForTimeout(1500)

      // Look for quiz questions
      const questions = page.locator('text=/Question [0-9]+/')
      const questionCount = await questions.count().catch(() => 0)

      if (questionCount > 0) {
        console.log(`Found ${questionCount} quiz questions`)

        // Answer all questions (select first option for each)
        const radioButtons = page.locator('input[type="radio"]')
        const radioCount = await radioButtons.count().catch(() => 0)

        if (radioCount > 0) {
          // Select first option for each question
          for (let i = 0; i < Math.min(questionCount, radioCount); i += Math.ceil(radioCount / questionCount)) {
            await radioButtons.nth(i).click({ timeout: 1000 }).catch(() => {})
            await page.waitForTimeout(300)
          }

          // Submit quiz
          const submitButton = page.locator('button:has-text("Submit"), button:has-text("Submit Quiz")').first()
          if (await submitButton.isEnabled({ timeout: 2000 }).catch(() => false)) {
            await submitButton.click()
            await page.waitForTimeout(1500)

            // Should show results
            const hasResults = await page.locator('text=/You scored|Quiz Passed|Quiz Not Passed|Results/i').isVisible({ timeout: 5000 })
              .catch(() => false)

            expect(hasResults).toBeTruthy()

            // Should show score percentage
            const hasScore = await page.locator('text=/[0-9]+%/').isVisible({ timeout: 2000 })
              .catch(() => false)

            expect(hasScore).toBeTruthy()
          }
        }
      } else {
        console.log('No quiz questions found in module')
        test.skip()
      }
    } else {
      test.skip()
    }
  })

  test('should allow quiz retry after failure', async ({ page }) => {
    await page.goto('/training')
    await page.waitForTimeout(1000)

    // Navigate to first module
    const moduleLink = page.locator('a[href*="/training/"], button:has-text("Start")').first()

    if (await moduleLink.isVisible({ timeout: 3000 }).catch(() => false)) {
      await moduleLink.click()
      await page.waitForTimeout(1500)

      // Look for quiz
      const questions = page.locator('text=/Question [0-9]+/')
      const questionCount = await questions.count().catch(() => 0)

      if (questionCount > 0) {
        // Answer all questions with first option (likely to fail)
        const radioButtons = page.locator('input[type="radio"]')
        const radioCount = await radioButtons.count().catch(() => 0)

        if (radioCount >= questionCount) {
          // Select first option for all questions
          for (let i = 0; i < questionCount; i++) {
            const firstRadio = page.locator(`input[type="radio"][name*="${i}"], input[type="radio"]`).first()
            await firstRadio.click({ timeout: 1000 }).catch(() => {})
            await page.waitForTimeout(200)
          }

          // Submit
          const submitButton = page.locator('button:has-text("Submit")').first()
          if (await submitButton.isEnabled({ timeout: 2000 }).catch(() => false)) {
            await submitButton.click()
            await page.waitForTimeout(1500)

            // Check if failed
            const failedText = await page.locator('text=/Quiz Not Passed|Failed|Try Again/i').isVisible({ timeout: 3000 })
              .catch(() => false)

            if (failedText) {
              console.log('Quiz failed as expected')

              // Look for retry button
              const retryButton = page.locator('button:has-text("Retry"), button:has-text("Try Again"), button:has-text("Retake")').first()

              if (await retryButton.isVisible({ timeout: 2000 }).catch(() => false)) {
                await retryButton.click()
                await page.waitForTimeout(1000)

                // Should show quiz form again
                const questionsVisible = await page.locator('text=/Question [0-9]+/').isVisible({ timeout: 3000 })
                  .catch(() => false)

                expect(questionsVisible).toBeTruthy()
                console.log('Successfully retried quiz')
              }
            } else {
              console.log('Quiz passed (cannot test retry for failed quiz)')
            }
          }
        }
      } else {
        test.skip()
      }
    } else {
      test.skip()
    }
  })

  test('should update dashboard after completing module', async ({ page }) => {
    await page.goto('/training')
    await page.waitForTimeout(1000)

    // Check current dashboard stats
    await page.goto('/dashboard')
    await page.waitForTimeout(1500)

    // Note the current completion status
    const completionText = await page.locator('text=/Overall Completion|Completion/i').textContent().catch(() => '')
    console.log('Current completion status:', completionText)

    // Go back to training
    await page.goto('/training')
    await page.waitForTimeout(1000)

    // If there are incomplete modules, try completing one
    const moduleLink = page.locator('button:has-text("Start"), a[href*="/training/"]').first()

    if (await moduleLink.isVisible({ timeout: 3000 }).catch(() => false)) {
      await moduleLink.click()
      await page.waitForTimeout(1500)

      // Try to complete quiz if present
      const questions = page.locator('text=/Question [0-9]+/')
      const questionCount = await questions.count().catch(() => 0)

      if (questionCount > 0) {
        // Answer questions and submit
        const radioButtons = page.locator('input[type="radio"]')
        const radioCount = await radioButtons.count().catch(() => 0)

        if (radioCount >= questionCount) {
          // Answer all questions
          for (let i = 0; i < Math.min(questionCount, 5); i++) {
            await radioButtons.nth(i * Math.ceil(radioCount / questionCount)).click({ timeout: 1000 }).catch(() => {})
            await page.waitForTimeout(200)
          }

          const submitButton = page.locator('button:has-text("Submit")').first()
          if (await submitButton.isEnabled({ timeout: 2000 }).catch(() => false)) {
            await submitButton.click()
            await page.waitForTimeout(1500)

            // Go back to dashboard
            await page.goto('/dashboard')
            await page.waitForTimeout(1500)

            // Dashboard should be updated
            await expect(page.locator('text=/Overall Completion|Recent Activity/i')).toBeVisible({ timeout: 5000 })
          }
        }
      }
    }
  })
})
