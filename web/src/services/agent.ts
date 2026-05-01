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

const AGENT_URL = import.meta.env.VITE_AGENT_URL ?? 'http://127.0.0.1:8737'
const BACKEND_URL = import.meta.env.VITE_BACKEND_URL ?? 'http://127.0.0.1:8081'

class AgentService {
  private baseUrl: string
  private backendUrl: string

  constructor(baseUrl: string = AGENT_URL, backendUrl: string = BACKEND_URL) {
    this.baseUrl = baseUrl
    this.backendUrl = backendUrl
  }

  /** Returns the stored JWT from sessionStorage, or null if not authenticated. */
  getToken(): string | null {
    return sessionStorage.getItem('qualify_token')
  }

  /** Stores a JWT in sessionStorage. */
  setToken(token: string): void {
    sessionStorage.setItem('qualify_token', token)
  }

  /** Clears the stored JWT. */
  clearToken(): void {
    sessionStorage.removeItem('qualify_token')
  }

  private async request<T>(
    endpoint: string,
    options?: RequestInit,
    useBackend: boolean = false
  ): Promise<T> {
    const baseUrl = useBackend ? this.backendUrl : this.baseUrl
    const url = `${baseUrl}${endpoint}`

    const token = this.getToken()
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    }
    if (token) headers['Authorization'] = `Bearer ${token}`
    if (options?.headers) {
      Object.entries(options.headers as Record<string, string>).forEach(([k, v]) => {
        headers[k] = v
      })
    }

    try {
      const response = await fetch(url, {
        ...options,
        headers,
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

  // Auth methods

  /** Returns the authenticated user's identity from the backend. */
  async getMe(): Promise<{ user_id: string; email: string; institution: string; role: string }> {
    return this.request('/api/auth/me', {}, true)
  }

  // Dashboard methods — routes now use context user (no user_id in URL)

  async getDashboardStats(_userId?: string): Promise<DashboardStats> {
    return this.request<DashboardStats>('/api/dashboard/stats', {}, true)
  }

  async getUserActivity(_userId?: string, limit?: number, offset?: number): Promise<ActivityItem[]> {
    const params = new URLSearchParams()
    if (limit !== undefined) params.append('limit', limit.toString())
    if (offset !== undefined) params.append('offset', offset.toString())
    const qs = params.toString()
    const response = await this.request<{ activities: ActivityItem[] }>(
      `/api/training/activity${qs ? `?${qs}` : ''}`, {}, true
    )
    return response.activities
  }

  async getUserProfile(_userId?: string): Promise<UserProfile> {
    return this.request<UserProfile>('/api/users/me', {}, true)
  }

  async updateUserProfile(_userId: string, updates: Partial<UserProfile>): Promise<UserProfile> {
    return this.request<UserProfile>('/api/users/me', {
      method: 'PUT',
      body: JSON.stringify(updates),
    }, true)
  }
}

export const agentService = new AgentService()
export default agentService
