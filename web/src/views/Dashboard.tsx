import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import Container from '@cloudscape-design/components/container'
import Header from '@cloudscape-design/components/header'
import SpaceBetween from '@cloudscape-design/components/space-between'
import Box from '@cloudscape-design/components/box'
import ColumnLayout from '@cloudscape-design/components/column-layout'
import Alert from '@cloudscape-design/components/alert'
import Button from '@cloudscape-design/components/button'
import Badge from '@cloudscape-design/components/badge'
import ProgressBar from '@cloudscape-design/components/progress-bar'
import { agentService } from '@/services/agent'
import type { DashboardStats, ActivityItem } from '@/types/api'

const USER_ID = '00000000-0000-0000-0000-000000000001'

export default function Dashboard() {
  const navigate = useNavigate()
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    loadDashboard()
  }, [])

  async function loadDashboard() {
    setIsLoading(true)
    setError(null)

    try {
      const data = await agentService.getDashboardStats(USER_ID)
      setStats(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load dashboard')
      console.error('Failed to load dashboard:', err)
    } finally {
      setIsLoading(false)
    }
  }

  function getActivityTypeLabel(type: string): string {
    const labels: Record<string, string> = {
      module_started: 'Started Module',
      module_completed: 'Completed Module',
      quiz_passed: 'Passed Quiz',
      quiz_failed: 'Failed Quiz',
      operation_blocked: 'Operation Blocked',
    }
    return labels[type] || type
  }

  function getActivityTypeBadge(type: string): 'green' | 'blue' | 'red' | 'grey' | undefined {
    if (type === 'module_completed' || type === 'quiz_passed') return 'green'
    if (type === 'module_started') return 'blue'
    if (type === 'quiz_failed' || type === 'operation_blocked') return 'red'
    return 'grey'
  }

  function formatTimestamp(timestamp: string): string {
    const date = new Date(timestamp)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)
    const diffDays = Math.floor(diffMs / 86400000)

    if (diffMins < 1) return 'Just now'
    if (diffMins < 60) return `${diffMins}m ago`
    if (diffHours < 24) return `${diffHours}h ago`
    if (diffDays < 7) return `${diffDays}d ago`
    return date.toLocaleDateString()
  }

  if (isLoading) {
    return (
      <Container>
        <Box variant="p">Loading dashboard...</Box>
      </Container>
    )
  }

  if (error) {
    return (
      <SpaceBetween size="l">
        <Header variant="h1">Dashboard</Header>
        <Alert
          type="error"
          header="Failed to load dashboard"
          action={<Button onClick={loadDashboard}>Retry</Button>}
        >
          {error}
        </Alert>
      </SpaceBetween>
    )
  }

  if (!stats) {
    return (
      <Container>
        <Box variant="p">No data available</Box>
      </Container>
    )
  }

  return (
    <SpaceBetween size="l">
      <Header variant="h1">Dashboard</Header>

      <Container
        header={
          <Header
            variant="h2"
            actions={
              <Button onClick={() => navigate('/training')}>View All Training</Button>
            }
          >
            Training Progress
          </Header>
        }
      >
        <SpaceBetween size="m">
          <ColumnLayout columns={4} variant="text-grid">
            <div>
              <Box variant="awsui-key-label">Total Modules</Box>
              <Box variant="awsui-value-large">{stats.training_summary.total_modules}</Box>
            </div>
            <div>
              <Box variant="awsui-key-label">Completed</Box>
              <Box variant="awsui-value-large" color="text-status-success">
                {stats.training_summary.completed}
              </Box>
            </div>
            <div>
              <Box variant="awsui-key-label">In Progress</Box>
              <Box variant="awsui-value-large" color="text-status-info">
                {stats.training_summary.in_progress}
              </Box>
            </div>
            <div>
              <Box variant="awsui-key-label">Not Started</Box>
              <Box variant="awsui-value-large">{stats.training_summary.not_started}</Box>
            </div>
          </ColumnLayout>

          <ProgressBar
            value={stats.training_summary.completion_percentage}
            label="Overall Completion"
            description={
              stats.training_summary.average_score !== undefined
                ? `Average quiz score: ${stats.training_summary.average_score}%`
                : undefined
            }
          />
        </SpaceBetween>
      </Container>

      <ColumnLayout columns={2}>
        <Container
          header={<Header variant="h2">Recent Activity</Header>}
        >
          {stats.recent_activity.length === 0 ? (
            <Box variant="p" color="text-body-secondary">
              No recent activity to display. Start a training module to see your progress here.
            </Box>
          ) : (
            <SpaceBetween size="s">
              {stats.recent_activity.map((activity: ActivityItem) => (
                <div key={activity.id}>
                  <SpaceBetween size="xxs">
                    <Box>
                      <SpaceBetween direction="horizontal" size="xs">
                        <Badge color={getActivityTypeBadge(activity.type)}>
                          {getActivityTypeLabel(activity.type)}
                        </Badge>
                        {activity.score !== undefined && (
                          <Badge color={activity.score >= 70 ? 'green' : 'red'}>
                            {activity.score}%
                          </Badge>
                        )}
                      </SpaceBetween>
                    </Box>
                    {activity.module_name && (
                      <Box variant="p" fontSize="body-s">
                        {activity.module_name}
                      </Box>
                    )}
                    <Box fontSize="body-s" color="text-body-secondary">
                      {formatTimestamp(activity.timestamp)}
                    </Box>
                  </SpaceBetween>
                </div>
              ))}
            </SpaceBetween>
          )}
        </Container>

        <Container
          header={<Header variant="h2">AWS Operations</Header>}
        >
          <SpaceBetween size="m">
            {stats.available_operations.unlocked.length > 0 && (
              <div>
                <Box variant="h3" fontSize="heading-s" padding={{ bottom: 'xs' }}>
                  Unlocked Operations
                </Box>
                <SpaceBetween size="xs">
                  {stats.available_operations.unlocked.map(op => (
                    <Box key={op} padding={{ left: 's' }}>
                      <SpaceBetween direction="horizontal" size="xs">
                        <Badge color="green">✓</Badge>
                        <Box fontSize="body-s">{op}</Box>
                      </SpaceBetween>
                    </Box>
                  ))}
                </SpaceBetween>
              </div>
            )}

            {stats.available_operations.locked.length > 0 && (
              <div>
                <Box variant="h3" fontSize="heading-s" padding={{ bottom: 'xs' }}>
                  Locked Operations
                </Box>
                <Box variant="p" fontSize="body-s" color="text-body-secondary" padding={{ bottom: 's' }}>
                  Complete required training to unlock these operations
                </Box>
                <SpaceBetween size="xs">
                  {stats.available_operations.locked.map(op => (
                    <Box key={op} padding={{ left: 's' }}>
                      <SpaceBetween direction="horizontal" size="xs">
                        <Badge color="grey">🔒</Badge>
                        <Box fontSize="body-s" color="text-body-secondary">
                          {op}
                        </Box>
                      </SpaceBetween>
                    </Box>
                  ))}
                </SpaceBetween>
              </div>
            )}

            {stats.available_operations.unlocked.length === 0 &&
             stats.available_operations.locked.length === 0 && (
              <Box variant="p" color="text-body-secondary">
                No operations configured
              </Box>
            )}
          </SpaceBetween>
        </Container>
      </ColumnLayout>
    </SpaceBetween>
  )
}
