import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import Modal from '@cloudscape-design/components/modal'
import Box from '@cloudscape-design/components/box'
import SpaceBetween from '@cloudscape-design/components/space-between'
import Button from '@cloudscape-design/components/button'
import Container from '@cloudscape-design/components/container'
import ColumnLayout from '@cloudscape-design/components/column-layout'
import Badge from '@cloudscape-design/components/badge'
import Icon from '@cloudscape-design/components/icon'
import Alert from '@cloudscape-design/components/alert'
import { agentService } from '@/services/agent'
import type { TrainingModule } from '@/types/api'

interface OnboardingWizardProps {
  visible: boolean
  userId: string
  trainingModules: TrainingModule[]
  onDismiss: () => void
  onComplete: () => void
}

export default function OnboardingWizard({
  visible,
  userId,
  trainingModules,
  onDismiss,
  onComplete
}: OnboardingWizardProps) {
  const navigate = useNavigate()
  const [currentStep, setCurrentStep] = useState(0)
  const [isCompleting, setIsCompleting] = useState(false)

  const steps = [
    {
      title: 'Welcome to ARK',
      content: (
        <SpaceBetween size="m">
          <Box variant="h2" textAlign="center">
            qualify Training
          </Box>
          <Box variant="p" textAlign="center" fontSize="heading-m" padding={{ bottom: 'm' }}>
            Safe, Controlled AWS Access for Research Environments
          </Box>

          <Container>
            <SpaceBetween size="s">
              <Box variant="h3">Security-First Design</Box>
              <Box variant="p">
                ARK provides a training-gated approach to AWS operations. Before you can perform
                actions like creating S3 buckets or launching EC2 instances, you'll complete
                relevant training modules.
              </Box>
            </SpaceBetween>
          </Container>

          <ColumnLayout columns={3}>
            <div>
              <SpaceBetween size="xs">
                <Icon name="status-positive" variant="success" size="large" />
                <Box variant="h4">Learn First</Box>
                <Box fontSize="body-s" color="text-body-secondary">
                  Complete interactive training modules on AWS services
                </Box>
              </SpaceBetween>
            </div>
            <div>
              <SpaceBetween size="xs">
                <Icon name="unlocked" variant="success" size="large" />
                <Box variant="h4">Unlock Operations</Box>
                <Box fontSize="body-s" color="text-body-secondary">
                  Pass quizzes to prove understanding and unlock capabilities
                </Box>
              </SpaceBetween>
            </div>
            <div>
              <SpaceBetween size="xs">
                <Icon name="search" variant="success" size="large" />
                <Box variant="h4">Full Audit Trail</Box>
                <Box fontSize="body-s" color="text-body-secondary">
                  All actions are logged for compliance and accountability
                </Box>
              </SpaceBetween>
            </div>
          </ColumnLayout>
        </SpaceBetween>
      )
    },
    {
      title: 'Available Training',
      content: (
        <SpaceBetween size="m">
          <Box variant="p">
            ARK includes {trainingModules.length} training modules covering essential AWS services.
            Each module includes educational content and a quiz to test your understanding.
          </Box>

          <Alert type="info">
            You need a score of 70% or higher to pass each quiz and unlock the associated AWS operations.
          </Alert>

          <SpaceBetween size="s">
            {trainingModules.map((module) => (
              <Container key={module.id}>
                <ColumnLayout columns={2} variant="text-grid">
                  <SpaceBetween size="xxs">
                    <Box variant="h4">{module.title}</Box>
                    <SpaceBetween direction="horizontal" size="xs">
                      {module.category && <Badge color="blue">{module.category}</Badge>}
                      {module.difficulty && <Badge>{module.difficulty}</Badge>}
                    </SpaceBetween>
                    {module.description && (
                      <Box fontSize="body-s" color="text-body-secondary">
                        {module.description}
                      </Box>
                    )}
                  </SpaceBetween>
                  <Box textAlign="right">
                    <Box fontSize="body-s" color="text-body-secondary">
                      <Icon name="status-in-progress" /> {module.estimated_minutes} minutes
                    </Box>
                  </Box>
                </ColumnLayout>
              </Container>
            ))}
          </SpaceBetween>
        </SpaceBetween>
      )
    },
    {
      title: 'Quick Start Guide',
      content: (
        <SpaceBetween size="m">
          <Box variant="h3">Getting Started with ARK</Box>

          <Container>
            <SpaceBetween size="s">
              <Box variant="h4">
                <Icon name="angle-right" /> Step 1: Complete Training
              </Box>
              <Box variant="p" padding={{ left: 'l' }}>
                Navigate to the Training section and select a module. Read through the content
                and complete the quiz at the end.
              </Box>
            </SpaceBetween>
          </Container>

          <Container>
            <SpaceBetween size="s">
              <Box variant="h4">
                <Icon name="angle-right" /> Step 2: Monitor Your Progress
              </Box>
              <Box variant="p" padding={{ left: 'l' }}>
                Check the Dashboard to see your training completion status, recent activity,
                and which AWS operations you've unlocked.
              </Box>
            </SpaceBetween>
          </Container>

          <Container>
            <SpaceBetween size="s">
              <Box variant="h4">
                <Icon name="angle-right" /> Step 3: Use AWS Services
              </Box>
              <Box variant="p" padding={{ left: 'l' }}>
                Once you've completed the required training, use the service pages (like S3)
                to perform operations. If training is incomplete, you'll see a training gate
                with links to the relevant modules.
              </Box>
            </SpaceBetween>
          </Container>

          <Alert type="success">
            <Box variant="p">
              <strong>Ready to begin?</strong> Click "Get Started" to view available training modules
              and start your first course.
            </Box>
          </Alert>
        </SpaceBetween>
      )
    }
  ]

  const currentStepData = steps[currentStep]
  const isFirstStep = currentStep === 0
  const isLastStep = currentStep === steps.length - 1

  async function handleComplete() {
    setIsCompleting(true)
    try {
      await agentService.updateUserProfile(userId, {
        preferences: {
          has_completed_onboarding: true,
          show_training_reminders: true
        }
      } as any)
      onComplete()
      navigate('/training')
    } catch (error) {
      console.error('Failed to update onboarding status:', error)
      onComplete()
      navigate('/training')
    } finally {
      setIsCompleting(false)
    }
  }

  function handleNext() {
    if (isLastStep) {
      handleComplete()
    } else {
      setCurrentStep(prev => prev + 1)
    }
  }

  function handlePrevious() {
    if (!isFirstStep) {
      setCurrentStep(prev => prev - 1)
    }
  }

  return (
    <Modal
      visible={visible}
      onDismiss={onDismiss}
      size="large"
      header={currentStepData.title}
      footer={
        <Box float="right">
          <SpaceBetween direction="horizontal" size="xs">
            {!isFirstStep && (
              <Button onClick={handlePrevious}>
                Previous
              </Button>
            )}
            <Button variant="primary" onClick={handleNext} loading={isCompleting}>
              {isLastStep ? 'Get Started' : 'Next'}
            </Button>
          </SpaceBetween>
        </Box>
      }
    >
      <SpaceBetween size="l">
        {currentStepData.content}

        <Box textAlign="center" fontSize="body-s" color="text-body-secondary">
          Step {currentStep + 1} of {steps.length}
        </Box>
      </SpaceBetween>
    </Modal>
  )
}
