import type {
  AgentHealthCheck,
  CreateBucketRequest,
  S3Bucket,
  PolicyDecision,
  ApiError,
  TrainingModule,
  StartModuleRequest,
  CompleteModuleRequest,
  QuizAnswer,
  QuizResponse,
  DashboardStats,
  ActivityItem,
  UserProfile
} from '@/types/api'

const AGENT_URL = 'http://127.0.0.1:8737'
const BACKEND_URL = 'http://127.0.0.1:8081'

class AgentService {
  private baseUrl: string
  private backendUrl: string

  constructor(baseUrl: string = AGENT_URL, backendUrl: string = BACKEND_URL) {
    this.baseUrl = baseUrl
    this.backendUrl = backendUrl
  }

  private async request<T>(
    endpoint: string,
    options?: RequestInit,
    useBackend: boolean = false
  ): Promise<T> {
    const baseUrl = useBackend ? this.backendUrl : this.baseUrl
    const url = `${baseUrl}${endpoint}`

    try {
      const response = await fetch(url, {
        ...options,
        headers: {
          'Content-Type': 'application/json',
          ...options?.headers,
        },
      })

      if (!response.ok) {
        const error: ApiError = await response.json()
        throw new Error(error.error || `HTTP ${response.status}: ${response.statusText}`)
      }

      return await response.json()
    } catch (error) {
      if (error instanceof Error) {
        throw error
      }
      throw new Error('An unknown error occurred')
    }
  }

  async healthCheck(): Promise<AgentHealthCheck> {
    return this.request<AgentHealthCheck>('/health')
  }

  async checkPing(): Promise<boolean> {
    try {
      await this.healthCheck()
      return true
    } catch {
      return false
    }
  }

  async createBucket(request: CreateBucketRequest): Promise<S3Bucket | PolicyDecision> {
    return this.request<S3Bucket | PolicyDecision>('/api/s3/buckets', {
      method: 'POST',
      body: JSON.stringify(request),
    })
  }

  async listCredentials(): Promise<string[]> {
    return this.request<string[]>('/api/credentials')
  }

  // Training module methods (call backend directly)
  async listTrainingModules(): Promise<TrainingModule[]> {
    const response = await this.request<{ modules: TrainingModule[] }>(
      '/api/training/modules',
      {},
      true
    )
    return response.modules
  }

  async getTrainingModule(moduleId: string): Promise<TrainingModule> {
    return this.request<TrainingModule>(
      `/api/training/modules/${moduleId}`,
      {},
      true
    )
  }

  async startTrainingModule(moduleId: string, userId: string): Promise<void> {
    const request: StartModuleRequest = { user_id: userId }
    await this.request<void>(
      `/api/training/modules/${moduleId}/start`,
      {
        method: 'POST',
        body: JSON.stringify(request),
      },
      true
    )
  }

  async completeTrainingModule(
    moduleId: string,
    userId: string,
    score?: number,
    timeSpentSeconds?: number
  ): Promise<void> {
    const request: CompleteModuleRequest = {
      user_id: userId,
      score,
      time_spent_seconds: timeSpentSeconds,
    }
    await this.request<void>(
      `/api/training/modules/${moduleId}/complete`,
      {
        method: 'POST',
        body: JSON.stringify(request),
      },
      true
    )
  }

  async submitQuizAnswers(
    moduleId: string,
    userId: string,
    answers: QuizAnswer[]
  ): Promise<QuizResponse> {
    return this.request<QuizResponse>(
      `/api/training/modules/${moduleId}/quiz/submit`,
      {
        method: 'POST',
        body: JSON.stringify({ user_id: userId, answers }),
      },
      true
    )
  }

  // Dashboard methods
  async getDashboardStats(userId: string): Promise<DashboardStats> {
    return this.request<DashboardStats>(
      `/api/dashboard/stats/${userId}`,
      {},
      true
    )
  }

  async getUserActivity(userId: string, limit?: number, offset?: number): Promise<ActivityItem[]> {
    const params = new URLSearchParams()
    if (limit !== undefined) params.append('limit', limit.toString())
    if (offset !== undefined) params.append('offset', offset.toString())

    const queryString = params.toString()
    const url = `/api/training/activity/${userId}${queryString ? `?${queryString}` : ''}`

    const response = await this.request<{ activities: ActivityItem[] }>(url, {}, true)
    return response.activities
  }

  async getUserProfile(userId: string): Promise<UserProfile> {
    return this.request<UserProfile>(
      `/api/users/${userId}/profile`,
      {},
      true
    )
  }

  async updateUserProfile(userId: string, updates: Partial<UserProfile>): Promise<UserProfile> {
    return this.request<UserProfile>(
      `/api/users/${userId}/profile`,
      {
        method: 'PUT',
        body: JSON.stringify(updates),
      },
      true
    )
  }
}

export const agentService = new AgentService()
export default agentService
