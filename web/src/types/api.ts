// Agent API types

export interface AgentStatus {
  status: 'running' | 'stopped'
  version?: string
  uptime?: number
}

export interface AgentHealthCheck {
  status: 'ok'
  version: string
}

export interface S3Bucket {
  bucket_name: string
  region: string
  location?: string
  created_at?: string
}

export interface CreateBucketRequest {
  bucket_name: string
  region: string
  encryption: {
    type: 'AES256' | 'aws:kms'
    kms_key_id?: string
  }
  versioning_enabled: boolean
  profile: string
}

export interface TrainingModule {
  id: string
  name: string
  title: string
  description?: string
  category?: string
  difficulty?: string
  estimated_minutes: number
  status?: 'not_started' | 'in_progress' | 'completed'
  content?: TrainingContent
  prerequisites?: string[]
}

export interface TrainingContent {
  sections: TrainingSection[]
}

export interface TrainingSection {
  id: string
  title: string
  type: 'text' | 'quiz'
  content?: string
  questions?: QuizQuestion[]
}

export interface QuizQuestion {
  id: string
  question: string
  options: string[]
  correctAnswer: number
  explanation?: string
}

export interface QuizAnswer {
  question_id: string
  selected_answer: number
}

export interface QuizSubmission {
  user_id: string
  answers: QuizAnswer[]
}

export interface QuizResult {
  question_id: string
  correct: boolean
  selected_answer: number
  correct_answer: number
  explanation?: string
}

export interface QuizResponse {
  score: number
  total_questions: number
  correct_answers: number
  results: QuizResult[]
}

export interface StartModuleRequest {
  user_id: string
}

export interface CompleteModuleRequest {
  user_id: string
  score?: number
  time_spent_seconds?: number
}

export interface PolicyDecision {
  action: 'allow' | 'block'
  reason?: string
  required_modules?: TrainingModule[]
  message: string
}

export interface AuditLogEntry {
  id: string
  user_id: string
  action: string
  resource_type: string
  resource_id: string
  status: 'success' | 'failure' | 'blocked'
  details: Record<string, any>
  ip_address?: string
  user_agent?: string
  created_at: string
}

export interface ApiError {
  error: string
  details?: string
}

export interface ActivityItem {
  id: string
  type: string
  module_name?: string
  module_id?: string
  timestamp: string
  score?: number
  metadata?: Record<string, any>
}

export interface TrainingSummary {
  total_modules: number
  completed: number
  in_progress: number
  not_started: number
  completion_percentage: number
  average_score?: number
}

export interface AvailableOps {
  unlocked: string[]
  locked: string[]
}

export interface DashboardStats {
  user_id: string
  training_summary: TrainingSummary
  recent_activity: ActivityItem[]
  available_operations: AvailableOps
}

export interface UserPreferences {
  has_completed_onboarding: boolean
  show_training_reminders: boolean
  default_aws_region?: string
}

export interface UserProfile {
  user_id: string
  email: string
  name: string
  institution: string
  role: string
  preferences: UserPreferences
  created_at: string
}
