import { useNavigate } from 'react-router-dom'
import Container from '@cloudscape-design/components/container'
import SpaceBetween from '@cloudscape-design/components/space-between'
import Box from '@cloudscape-design/components/box'
import Button from '@cloudscape-design/components/button'
import Badge from '@cloudscape-design/components/badge'
import ColumnLayout from '@cloudscape-design/components/column-layout'
import Icon from '@cloudscape-design/components/icon'
import type { TrainingModule } from '@/types/api'

interface TrainingGateProps {
  requiredModules: TrainingModule[]
  operationName?: string
  onDismiss?: () => void
}

export default function TrainingGate({ requiredModules, operationName = 'this operation', onDismiss }: TrainingGateProps) {
  const navigate = useNavigate()

  if (requiredModules.length === 0) return null

  const totalEstimatedTime = requiredModules.reduce((sum, module) => sum + module.estimated_minutes, 0)

  function handleStartTraining(moduleName: string) {
    navigate(`/training/${moduleName}`)
  }

  function handleBrowseAll() {
    navigate('/training')
  }

  return (
    <Container>
      <SpaceBetween size="m">
        <Box textAlign="center" padding={{ vertical: 'm' }}>
          <SpaceBetween size="s">
            <Icon name="status-warning" size="big" variant="warning" />
            <Box variant="h2">Training Required</Box>
            <Box variant="p" color="text-body-secondary">
              Complete the following training modules to unlock {operationName}
            </Box>
          </SpaceBetween>
        </Box>

        <ColumnLayout columns={requiredModules.length === 1 ? 1 : 2}>
          {requiredModules.map((module) => (
            <Container key={module.id}>
              <SpaceBetween size="s">
                <SpaceBetween size="xxs">
                  <Box variant="h3">{module.title}</Box>
                  {module.category && (
                    <Badge color="blue">{module.category}</Badge>
                  )}
                </SpaceBetween>

                {module.description && (
                  <Box variant="p" fontSize="body-s">
                    {module.description}
                  </Box>
                )}

                <SpaceBetween direction="horizontal" size="xs">
                  <Box fontSize="body-s" color="text-body-secondary">
                    <Icon name="status-in-progress" /> {module.estimated_minutes} minutes
                  </Box>
                  {module.difficulty && (
                    <Box fontSize="body-s" color="text-body-secondary">
                      <Icon name="zoom-to-fit" /> {module.difficulty}
                    </Box>
                  )}
                </SpaceBetween>

                <Button
                  variant="primary"
                  onClick={() => handleStartTraining(module.name)}
                  iconName="arrow-right"
                  iconAlign="right"
                >
                  Start Training
                </Button>
              </SpaceBetween>
            </Container>
          ))}
        </ColumnLayout>

        <Box textAlign="center" padding={{ top: 's' }}>
          <SpaceBetween size="s">
            <Box variant="p" fontSize="body-s" color="text-body-secondary">
              Total estimated time: {totalEstimatedTime} minutes
            </Box>
            <SpaceBetween direction="horizontal" size="xs">
              <Button onClick={handleBrowseAll}>Browse All Training</Button>
              {onDismiss && (
                <Button onClick={onDismiss}>Dismiss</Button>
              )}
            </SpaceBetween>
          </SpaceBetween>
        </Box>
      </SpaceBetween>
    </Container>
  )
}
