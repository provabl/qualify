/**
 * Mock data generators for testing
 */

import type {
  DashboardStats,
  QuizQuestion,
  QuizResponse,
  QuizResult,
  ActivityItem,
  TrainingModule,
} from '@/types/api'

/**
 * Generate mock quiz questions
 */
export function createMockQuizQuestions(count: number = 3): QuizQuestion[] {
  const questions: QuizQuestion[] = []

  for (let i = 0; i < count; i++) {
    questions.push({
      id: `q${i + 1}`,
      question: `Sample question ${i + 1}?`,
      options: [
        `Option A for question ${i + 1}`,
        `Option B for question ${i + 1}`,
        `Option C for question ${i + 1}`,
        `Option D for question ${i + 1}`,
      ],
      correctAnswer: 1, // Option B is correct
      explanation: `Explanation for question ${i + 1}`,
    })
  }

  return questions
}

/**
 * Generate mock quiz response
 */
export function createMockQuizResponse(
  options: {
    score?: number
    totalQuestions?: number
    correctAnswers?: number
    passed?: boolean
  } = {}
): QuizResponse {
  const {
    score = 80,
    totalQuestions = 3,
    correctAnswers = 2,
    passed = true,
  } = options

  const results: QuizResult[] = []

  for (let i = 0; i < totalQuestions; i++) {
    results.push({
      question_id: `q${i + 1}`,
      correct: i < correctAnswers,
      selected_answer: i < correctAnswers ? 1 : 0,
      correct_answer: 1,
      explanation: `Explanation for question ${i + 1}`,
    })
  }

  return {
    score,
    total_questions: totalQuestions,
    correct_answers: correctAnswers,
    results,
  }
}

/**
 * Generate mock activity items
 */
export function createMockActivityItems(count: number = 5): ActivityItem[] {
  const activityTypes = [
    'module_started',
    'module_completed',
    'quiz_passed',
    'quiz_failed',
    'operation_blocked',
  ]

  const moduleNames = ['S3 Basics', 'S3 Security', 'IAM Basics', 'EC2 Basics']

  const activities: ActivityItem[] = []

  for (let i = 0; i < count; i++) {
    const type = activityTypes[i % activityTypes.length]
    const hasScore = type === 'quiz_passed' || type === 'quiz_failed'
    const hasModule = type !== 'operation_blocked'

    const activity: ActivityItem = {
      id: `act-${i + 1}`,
      type,
      timestamp: new Date(Date.now() - i * 3600000).toISOString(), // i hours ago
    }

    if (hasModule) {
      activity.module_name = moduleNames[i % moduleNames.length]
      activity.module_id = `mod-${i + 1}`
    }

    if (hasScore) {
      activity.score = type === 'quiz_passed' ? 80 + i * 5 : 60 - i * 5
    }

    activities.push(activity)
  }

  return activities
}

/**
 * Generate mock dashboard stats
 */
export function createMockDashboardStats(
  options: {
    userId?: string
    completed?: number
    inProgress?: number
    notStarted?: number
    averageScore?: number
    activityCount?: number
    unlockedOps?: string[]
    lockedOps?: string[]
  } = {}
): DashboardStats {
  const {
    userId = 'test-user',
    completed = 2,
    inProgress = 1,
    notStarted = 1,
    averageScore = 85,
    activityCount = 5,
    unlockedOps = ['s3:CreateBucket', 's3:ListBuckets'],
    lockedOps = ['ec2:RunInstance', 'iam:CreateUser'],
  } = options

  const totalModules = completed + inProgress + notStarted
  const completionPercentage = Math.round((completed / totalModules) * 100)

  return {
    user_id: userId,
    training_summary: {
      total_modules: totalModules,
      completed,
      in_progress: inProgress,
      not_started: notStarted,
      completion_percentage: completionPercentage,
      average_score: completed > 0 ? averageScore : undefined,
    },
    recent_activity: createMockActivityItems(activityCount),
    available_operations: {
      unlocked: unlockedOps,
      locked: lockedOps,
    },
  }
}

/**
 * Generate mock training module
 */
export function createMockTrainingModule(
  options: {
    id?: string
    name?: string
    title?: string
    status?: string
    completedAt?: string
    score?: number
  } = {}
): TrainingModule {
  const {
    id = 'mod-1',
    name = 's3-basics',
    title = 'S3 Basics',
    status = 'not_started',
    completedAt,
    score,
  } = options

  return {
    id,
    name,
    title,
    description: `Learn the basics of ${title}`,
    status,
    completed_at: completedAt,
    score,
    content: {
      introduction: `Introduction to ${title}`,
      sections: [
        {
          title: 'Overview',
          content: 'Overview content',
        },
        {
          title: 'Key Concepts',
          content: 'Key concepts content',
        },
      ],
      quiz: createMockQuizQuestions(3),
    },
  }
}

/**
 * Generate multiple mock training modules
 */
export function createMockTrainingModules(count: number = 4): TrainingModule[] {
  const modules: TrainingModule[] = []
  const statuses = ['completed', 'in_progress', 'not_started', 'not_started']
  const names = ['s3-basics', 's3-security', 'iam-basics', 'ec2-basics']
  const titles = ['S3 Basics', 'S3 Security', 'IAM Basics', 'EC2 Basics']

  for (let i = 0; i < count; i++) {
    modules.push(
      createMockTrainingModule({
        id: `mod-${i + 1}`,
        name: names[i],
        title: titles[i],
        status: statuses[i],
        completedAt: statuses[i] === 'completed' ? new Date().toISOString() : undefined,
        score: statuses[i] === 'completed' ? 85 + i * 5 : undefined,
      })
    )
  }

  return modules
}

/**
 * Create empty dashboard stats for new user
 */
export function createEmptyDashboardStats(userId: string = 'test-user'): DashboardStats {
  return createMockDashboardStats({
    userId,
    completed: 0,
    inProgress: 0,
    notStarted: 4,
    averageScore: undefined,
    activityCount: 0,
    unlockedOps: [],
    lockedOps: ['s3:CreateBucket', 's3:ListBuckets', 'ec2:RunInstance', 'iam:CreateUser'],
  })
}
