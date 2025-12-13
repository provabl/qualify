import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import Container from '@cloudscape-design/components/container'
import Header from '@cloudscape-design/components/header'
import SpaceBetween from '@cloudscape-design/components/space-between'
import Box from '@cloudscape-design/components/box'
import Cards from '@cloudscape-design/components/cards'
import Button from '@cloudscape-design/components/button'
import StatusIndicator from '@cloudscape-design/components/status-indicator'
import { agentService } from '@/services/agent'
import type { TrainingModule } from '@/types/api'

function getStatusType(status?: string) {
  switch (status) {
    case 'completed':
      return 'success'
    case 'in_progress':
      return 'in-progress'
    case 'not_started':
    default:
      return 'pending'
  }
}

function getStatusText(status?: string) {
  switch (status) {
    case 'completed':
      return 'Completed'
    case 'in_progress':
      return 'In Progress'
    case 'not_started':
    default:
      return 'Not Started'
  }
}

export default function Training() {
  const navigate = useNavigate()
  const [trainingModules, setTrainingModules] = useState<TrainingModule[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadModules()
  }, [])

  async function loadModules() {
    setIsLoading(true)
    setError(null)

    try {
      const modules = await agentService.listTrainingModules()
      setTrainingModules(modules)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load training modules')
      console.error('Failed to load training modules:', err)
    } finally {
      setIsLoading(false)
    }
  }

  function startTraining(moduleName: string) {
    navigate(`/training/${moduleName}`)
  }

  return (
    <SpaceBetween size="l">
      <Header variant="h1">
        Training Modules
      </Header>

      <Container>
        <SpaceBetween size="m">
          <Box variant="p">
            Complete training modules to unlock AWS operations. Some operations require specific training completion.
          </Box>

          {error && (
            <Box variant="p" color="text-status-error">
              Error: {error}
            </Box>
          )}

          {isLoading && (
            <Box variant="p">
              Loading training modules...
            </Box>
          )}

          {!isLoading && !error && (
            <Cards
              items={trainingModules}
              cardsPerRow={[{ cards: 1 }, { minWidth: 500, cards: 2 }]}
              cardDefinition={{
                header: (item) => (
                  <SpaceBetween direction="horizontal" size="xs">
                    <Box variant="h2">{item.title}</Box>
                    <StatusIndicator type={getStatusType(item.status)}>
                      {getStatusText(item.status)}
                    </StatusIndicator>
                  </SpaceBetween>
                ),
                sections: [
                  {
                    id: 'description',
                    content: (item) => (
                      <SpaceBetween size="s">
                        <Box variant="p">{item.description}</Box>
                        <Box fontSize="body-s">
                          <strong>Estimated time:</strong> {item.estimated_minutes} minutes
                        </Box>
                        {item.category && (
                          <Box fontSize="body-s">
                            <strong>Category:</strong> {item.category}
                          </Box>
                        )}
                        {item.difficulty && (
                          <Box fontSize="body-s">
                            <strong>Difficulty:</strong> {item.difficulty}
                          </Box>
                        )}
                        <Button
                          onClick={() => startTraining(item.name)}
                          variant={item.status === 'completed' ? 'normal' : 'primary'}
                        >
                          {item.status === 'completed' ? 'Review' : 'Start Training'}
                        </Button>
                      </SpaceBetween>
                    ),
                  },
                ],
              }}
            />
          )}
        </SpaceBetween>
      </Container>
    </SpaceBetween>
  )
}
