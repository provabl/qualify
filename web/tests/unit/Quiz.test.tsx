import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import Quiz from '@/components/training/Quiz'
import { agentService } from '@/services/agent'
import type { QuizQuestion } from '@/types/api'

// Mock the agent service
vi.mock('@/services/agent', () => ({
  agentService: {
    submitQuizAnswers: vi.fn(),
  },
}))

const mockQuestions: QuizQuestion[] = [
  {
    id: 'q1',
    question: 'What is the maximum size of an S3 object?',
    options: ['5 GB', '5 TB', '50 TB', 'No limit'],
    correctAnswer: 1,
    explanation: 'The maximum object size in S3 is 5 TB.',
  },
  {
    id: 'q2',
    question: 'Which encryption type uses AWS KMS?',
    options: ['AES256', 'SSE-KMS', 'SSE-C', 'None'],
    correctAnswer: 1,
    explanation: 'SSE-KMS uses AWS Key Management Service.',
  },
]

describe('Quiz Component', () => {
  const mockOnComplete = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders all questions', () => {
    render(
      <Quiz
        moduleId="test-module"
        userId="test-user"
        questions={mockQuestions}
        onComplete={mockOnComplete}
      />
    )

    expect(screen.getByText(/Question 1:/)).toBeInTheDocument()
    expect(screen.getByText(/What is the maximum size of an S3 object/i)).toBeInTheDocument()
    expect(screen.getByText(/Question 2:/)).toBeInTheDocument()
    expect(screen.getByText(/Which encryption type uses AWS KMS/i)).toBeInTheDocument()
  })

  it('renders all answer options', () => {
    render(
      <Quiz
        moduleId="test-module"
        userId="test-user"
        questions={mockQuestions}
        onComplete={mockOnComplete}
      />
    )

    expect(screen.getByText('5 GB')).toBeInTheDocument()
    expect(screen.getByText('5 TB')).toBeInTheDocument()
    expect(screen.getByText('AES256')).toBeInTheDocument()
    expect(screen.getByText('SSE-KMS')).toBeInTheDocument()
  })

  it('disables submit button when not all questions answered', () => {
    render(
      <Quiz
        moduleId="test-module"
        userId="test-user"
        questions={mockQuestions}
        onComplete={mockOnComplete}
      />
    )

    const submitButton = screen.getByRole('button', { name: /submit quiz/i })
    expect(submitButton).toBeDisabled()
  })

  it('enables submit button when all questions answered', async () => {
    const user = userEvent.setup()

    render(
      <Quiz
        moduleId="test-module"
        userId="test-user"
        questions={mockQuestions}
        onComplete={mockOnComplete}
      />
    )

    // Answer both questions
    const q1Option = screen.getByLabelText('5 GB')
    const q2Option = screen.getByLabelText('AES256')

    await user.click(q1Option)
    await user.click(q2Option)

    const submitButton = screen.getByRole('button', { name: /submit quiz/i })
    expect(submitButton).not.toBeDisabled()
  })

  it('submits quiz and displays pass result', async () => {
    const user = userEvent.setup()

    const mockResponse = {
      score: 100,
      total_questions: 2,
      correct_answers: 2,
      results: [
        {
          question_id: 'q1',
          correct: true,
          selected_answer: 1,
          correct_answer: 1,
          explanation: 'The maximum object size in S3 is 5 TB.',
        },
        {
          question_id: 'q2',
          correct: true,
          selected_answer: 1,
          correct_answer: 1,
          explanation: 'SSE-KMS uses AWS Key Management Service.',
        },
      ],
    }

    vi.mocked(agentService.submitQuizAnswers).mockResolvedValue(mockResponse)

    render(
      <Quiz
        moduleId="test-module"
        userId="test-user"
        questions={mockQuestions}
        onComplete={mockOnComplete}
      />
    )

    // Answer both questions correctly
    await user.click(screen.getByLabelText('5 TB'))
    await user.click(screen.getByLabelText('SSE-KMS'))

    // Submit quiz
    await user.click(screen.getByRole('button', { name: /submit quiz/i }))

    // Wait for results
    await waitFor(() => {
      expect(screen.getByText('Quiz Passed!')).toBeInTheDocument()
    })

    expect(screen.getByText(/You scored 100%/)).toBeInTheDocument()
    expect(mockOnComplete).toHaveBeenCalledWith(100, true)
  })

  it('submits quiz and displays fail result', async () => {
    const user = userEvent.setup()

    const mockResponse = {
      score: 50,
      total_questions: 2,
      correct_answers: 1,
      results: [
        {
          question_id: 'q1',
          correct: false,
          selected_answer: 0,
          correct_answer: 1,
          explanation: 'The maximum object size in S3 is 5 TB.',
        },
        {
          question_id: 'q2',
          correct: true,
          selected_answer: 1,
          correct_answer: 1,
          explanation: 'SSE-KMS uses AWS Key Management Service.',
        },
      ],
    }

    vi.mocked(agentService.submitQuizAnswers).mockResolvedValue(mockResponse)

    render(
      <Quiz
        moduleId="test-module"
        userId="test-user"
        questions={mockQuestions}
        onComplete={mockOnComplete}
      />
    )

    // Answer questions (one wrong)
    await user.click(screen.getByLabelText('5 GB'))
    await user.click(screen.getByLabelText('SSE-KMS'))

    // Submit quiz
    await user.click(screen.getByRole('button', { name: /submit quiz/i }))

    // Wait for results
    await waitFor(() => {
      expect(screen.getByText('Quiz Not Passed')).toBeInTheDocument()
    })

    expect(screen.getByText(/You scored 50%/)).toBeInTheDocument()
    expect(screen.getByText(/You need 70% to pass/)).toBeInTheDocument()
    expect(mockOnComplete).toHaveBeenCalledWith(50, false)
  })

  it('displays retry button after failing', async () => {
    const user = userEvent.setup()

    const mockResponse = {
      score: 50,
      total_questions: 2,
      correct_answers: 1,
      results: [
        {
          question_id: 'q1',
          correct: false,
          selected_answer: 0,
          correct_answer: 1,
        },
        {
          question_id: 'q2',
          correct: true,
          selected_answer: 1,
          correct_answer: 1,
        },
      ],
    }

    vi.mocked(agentService.submitQuizAnswers).mockResolvedValue(mockResponse)

    render(
      <Quiz
        moduleId="test-module"
        userId="test-user"
        questions={mockQuestions}
        onComplete={mockOnComplete}
      />
    )

    // Answer and submit
    await user.click(screen.getByLabelText('5 GB'))
    await user.click(screen.getByLabelText('SSE-KMS'))
    await user.click(screen.getByRole('button', { name: /submit quiz/i }))

    // Wait for results and retry button
    await waitFor(() => {
      expect(screen.getByText('Retry Quiz')).toBeInTheDocument()
    })

    // Click retry
    await user.click(screen.getByText('Retry Quiz'))

    // Should show quiz form again
    expect(screen.getByRole('button', { name: /submit quiz/i })).toBeInTheDocument()
  })

  it('shows error message on submission failure', async () => {
    const user = userEvent.setup()

    vi.mocked(agentService.submitQuizAnswers).mockRejectedValue(new Error('Network error'))

    render(
      <Quiz
        moduleId="test-module"
        userId="test-user"
        questions={mockQuestions}
        onComplete={mockOnComplete}
      />
    )

    // Answer and submit
    await user.click(screen.getByLabelText('5 TB'))
    await user.click(screen.getByLabelText('SSE-KMS'))
    await user.click(screen.getByRole('button', { name: /submit quiz/i }))

    // Wait for error message
    await waitFor(() => {
      expect(screen.getByText('Network error')).toBeInTheDocument()
    })
  })

  it('shows validation message when not all questions answered', () => {
    render(
      <Quiz
        moduleId="test-module"
        userId="test-user"
        questions={mockQuestions}
        onComplete={mockOnComplete}
      />
    )

    expect(screen.getByText('Please answer all questions before submitting')).toBeInTheDocument()
  })

  it('displays explanations in results', async () => {
    const user = userEvent.setup()

    const mockResponse = {
      score: 100,
      total_questions: 2,
      correct_answers: 2,
      results: [
        {
          question_id: 'q1',
          correct: true,
          selected_answer: 1,
          correct_answer: 1,
          explanation: 'The maximum object size in S3 is 5 TB.',
        },
        {
          question_id: 'q2',
          correct: true,
          selected_answer: 1,
          correct_answer: 1,
          explanation: 'SSE-KMS uses AWS Key Management Service.',
        },
      ],
    }

    vi.mocked(agentService.submitQuizAnswers).mockResolvedValue(mockResponse)

    render(
      <Quiz
        moduleId="test-module"
        userId="test-user"
        questions={mockQuestions}
        onComplete={mockOnComplete}
      />
    )

    // Answer and submit
    await user.click(screen.getByLabelText('5 TB'))
    await user.click(screen.getByLabelText('SSE-KMS'))
    await user.click(screen.getByRole('button', { name: /submit quiz/i }))

    // Wait for results
    await waitFor(() => {
      expect(screen.getByText('The maximum object size in S3 is 5 TB.')).toBeInTheDocument()
      expect(screen.getByText('SSE-KMS uses AWS Key Management Service.')).toBeInTheDocument()
    })
  })
})
