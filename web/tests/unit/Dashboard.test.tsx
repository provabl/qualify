import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import { BrowserRouter } from 'react-router-dom'
import Dashboard from '@/views/Dashboard'
import { agentService } from '@/services/agent'
import type { DashboardStats } from '@/types/api'

// Mock the agent service
vi.mock('@/services/agent', () => ({
  agentService: {
    getDashboardStats: vi.fn(),
  },
}))

// Mock react-router-dom navigate
const mockNavigate = vi.fn()
vi.mock('react-router-dom', async () => {
  const actual = await vi.importActual('react-router-dom')
  return {
    ...actual,
    useNavigate: () => mockNavigate,
  }
})

const mockDashboardStats: DashboardStats = {
  user_id: 'test-user',
  training_summary: {
    total_modules: 4,
    completed: 2,
    in_progress: 1,
    not_started: 1,
    completion_percentage: 50,
    average_score: 85,
  },
  recent_activity: [
    {
      id: 'act-1',
      type: 'module_completed',
      module_name: 'S3 Basics',
      module_id: 'mod-1',
      timestamp: new Date().toISOString(),
      score: 90,
    },
    {
      id: 'act-2',
      type: 'quiz_passed',
      module_name: 'S3 Security',
      module_id: 'mod-2',
      timestamp: new Date(Date.now() - 3600000).toISOString(), // 1 hour ago
      score: 80,
    },
  ],
  available_operations: {
    unlocked: ['s3:CreateBucket', 's3:ListBuckets'],
    locked: ['ec2:RunInstance', 'iam:CreateUser'],
  },
}

const mockEmptyDashboard: DashboardStats = {
  user_id: 'test-user',
  training_summary: {
    total_modules: 4,
    completed: 0,
    in_progress: 0,
    not_started: 4,
    completion_percentage: 0,
  },
  recent_activity: [],
  available_operations: {
    unlocked: [],
    locked: ['s3:CreateBucket', 's3:ListBuckets', 'ec2:RunInstance', 'iam:CreateUser'],
  },
}

function renderDashboard() {
  return render(
    <BrowserRouter>
      <Dashboard />
    </BrowserRouter>
  )
}

describe('Dashboard Component', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('displays loading state initially', () => {
    vi.mocked(agentService.getDashboardStats).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    )

    renderDashboard()
    expect(screen.getByText('Loading dashboard...')).toBeInTheDocument()
  })

  it('displays dashboard stats after loading', async () => {
    vi.mocked(agentService.getDashboardStats).mockResolvedValue(mockDashboardStats)

    renderDashboard()

    await waitFor(() => {
      expect(screen.getByText('Dashboard')).toBeInTheDocument()
    })

    // Check training summary metrics
    expect(screen.getByText('Total Modules')).toBeInTheDocument()
    expect(screen.getByText('Completed')).toBeInTheDocument()
    expect(screen.getByText('In Progress')).toBeInTheDocument()
    expect(screen.getByText('Not Started')).toBeInTheDocument()
  })

  it('displays completion percentage', async () => {
    vi.mocked(agentService.getDashboardStats).mockResolvedValue(mockDashboardStats)

    renderDashboard()

    await waitFor(() => {
      expect(screen.getByText('Overall Completion')).toBeInTheDocument()
    })

    // Progress bar should show 50%
    expect(screen.getByText(/Average quiz score: 85%/i)).toBeInTheDocument()
  })

  it('displays recent activity', async () => {
    vi.mocked(agentService.getDashboardStats).mockResolvedValue(mockDashboardStats)

    renderDashboard()

    await waitFor(() => {
      expect(screen.getByText('Recent Activity')).toBeInTheDocument()
    })

    expect(screen.getByText('S3 Basics')).toBeInTheDocument()
    expect(screen.getByText('S3 Security')).toBeInTheDocument()
    expect(screen.getByText('Completed Module')).toBeInTheDocument()
    expect(screen.getByText('Passed Quiz')).toBeInTheDocument()
  })

  it('displays empty state for no activity', async () => {
    vi.mocked(agentService.getDashboardStats).mockResolvedValue(mockEmptyDashboard)

    renderDashboard()

    await waitFor(() => {
      expect(screen.getByText(/No recent activity to display/i)).toBeInTheDocument()
    })

    expect(screen.getByText(/Start a training module to see your progress here/i)).toBeInTheDocument()
  })

  it('displays unlocked operations', async () => {
    vi.mocked(agentService.getDashboardStats).mockResolvedValue(mockDashboardStats)

    renderDashboard()

    await waitFor(() => {
      expect(screen.getByText('Unlocked Operations')).toBeInTheDocument()
    })

    expect(screen.getByText('s3:CreateBucket')).toBeInTheDocument()
    expect(screen.getByText('s3:ListBuckets')).toBeInTheDocument()
  })

  it('displays locked operations', async () => {
    vi.mocked(agentService.getDashboardStats).mockResolvedValue(mockDashboardStats)

    renderDashboard()

    await waitFor(() => {
      expect(screen.getByText('Locked Operations')).toBeInTheDocument()
    })

    expect(screen.getByText('ec2:RunInstance')).toBeInTheDocument()
    expect(screen.getByText('iam:CreateUser')).toBeInTheDocument()
    expect(screen.getByText(/Complete required training to unlock/i)).toBeInTheDocument()
  })

  it('displays empty state when no operations configured', async () => {
    const emptyOps = {
      ...mockEmptyDashboard,
      available_operations: {
        unlocked: [],
        locked: [],
      },
    }

    vi.mocked(agentService.getDashboardStats).mockResolvedValue(emptyOps)

    renderDashboard()

    await waitFor(() => {
      expect(screen.getByText('AWS Operations')).toBeInTheDocument()
    })

    expect(screen.getByText('No operations configured')).toBeInTheDocument()
  })

  it('navigates to training when View All Training clicked', async () => {
    vi.mocked(agentService.getDashboardStats).mockResolvedValue(mockDashboardStats)

    renderDashboard()

    await waitFor(() => {
      expect(screen.getByText('View All Training')).toBeInTheDocument()
    })

    const button = screen.getByText('View All Training')
    button.click()

    expect(mockNavigate).toHaveBeenCalledWith('/training')
  })

  it('displays error state on API failure', async () => {
    vi.mocked(agentService.getDashboardStats).mockRejectedValue(new Error('Network error'))

    renderDashboard()

    await waitFor(() => {
      expect(screen.getByText('Failed to load dashboard')).toBeInTheDocument()
    })

    expect(screen.getByText('Network error')).toBeInTheDocument()
    expect(screen.getByText('Retry')).toBeInTheDocument()
  })

  it('retries loading on retry button click', async () => {
    vi.mocked(agentService.getDashboardStats)
      .mockRejectedValueOnce(new Error('Network error'))
      .mockResolvedValueOnce(mockDashboardStats)

    renderDashboard()

    // Wait for error
    await waitFor(() => {
      expect(screen.getByText('Retry')).toBeInTheDocument()
    })

    // Click retry
    const retryButton = screen.getByText('Retry')
    retryButton.click()

    // Should show success
    await waitFor(() => {
      expect(screen.getByText('Overall Completion')).toBeInTheDocument()
    })
  })

  it('formats activity timestamps correctly', async () => {
    vi.mocked(agentService.getDashboardStats).mockResolvedValue(mockDashboardStats)

    renderDashboard()

    await waitFor(() => {
      expect(screen.getByText('Recent Activity')).toBeInTheDocument()
    })

    // Should show relative time like "1h ago" or "Just now"
    const timeElements = screen.getAllByText(/ago|Just now/i)
    expect(timeElements.length).toBeGreaterThan(0)
  })

  it('displays score badges for activities with scores', async () => {
    vi.mocked(agentService.getDashboardStats).mockResolvedValue(mockDashboardStats)

    renderDashboard()

    await waitFor(() => {
      expect(screen.getByText('Recent Activity')).toBeInTheDocument()
    })

    expect(screen.getByText('90%')).toBeInTheDocument()
    expect(screen.getByText('80%')).toBeInTheDocument()
  })

  it('shows 0% completion for new users', async () => {
    vi.mocked(agentService.getDashboardStats).mockResolvedValue(mockEmptyDashboard)

    renderDashboard()

    await waitFor(() => {
      expect(screen.getByText('Overall Completion')).toBeInTheDocument()
    })

    // Should not show average score if none exist
    expect(screen.queryByText(/Average quiz score/i)).not.toBeInTheDocument()
  })

  it('displays all activity type badges correctly', async () => {
    const statsWithVariedActivity: DashboardStats = {
      ...mockDashboardStats,
      recent_activity: [
        {
          id: 'act-1',
          type: 'module_started',
          module_name: 'EC2 Basics',
          timestamp: new Date().toISOString(),
        },
        {
          id: 'act-2',
          type: 'quiz_failed',
          module_name: 'IAM Basics',
          timestamp: new Date().toISOString(),
          score: 60,
        },
        {
          id: 'act-3',
          type: 'operation_blocked',
          timestamp: new Date().toISOString(),
        },
      ],
    }

    vi.mocked(agentService.getDashboardStats).mockResolvedValue(statsWithVariedActivity)

    renderDashboard()

    await waitFor(() => {
      expect(screen.getByText('Started Module')).toBeInTheDocument()
    })

    expect(screen.getByText('Failed Quiz')).toBeInTheDocument()
    expect(screen.getByText('Operation Blocked')).toBeInTheDocument()
    expect(screen.getByText('60%')).toBeInTheDocument()
  })
})
