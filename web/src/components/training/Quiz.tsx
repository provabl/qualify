import { useState, useMemo, useCallback } from 'react'
import SpaceBetween from '@cloudscape-design/components/space-between'
import Box from '@cloudscape-design/components/box'
import Button from '@cloudscape-design/components/button'
import RadioGroup from '@cloudscape-design/components/radio-group'
import Alert from '@cloudscape-design/components/alert'
import Container from '@cloudscape-design/components/container'
import Header from '@cloudscape-design/components/header'
import { agentService } from '@/services/agent'
import type { QuizQuestion, QuizAnswer, QuizResponse } from '@/types/api'

interface QuizProps {
  moduleId: string
  userId: string
  questions: QuizQuestion[]
  onComplete: (score: number, passed: boolean) => void
}

export default function Quiz({ moduleId, userId, questions, onComplete }: QuizProps) {
  const [selectedAnswers, setSelectedAnswers] = useState<Record<string, number>>({})
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [quizResult, setQuizResult] = useState<QuizResponse | null>(null)
  const [error, setError] = useState<string | null>(null)

  const allQuestionsAnswered = useMemo(() => {
    return questions.every(q => selectedAnswers[q.id] !== undefined)
  }, [questions, selectedAnswers])

  const handleAnswerChange = useCallback((questionId: string, value: string) => {
    setSelectedAnswers(prev => ({
      ...prev,
      [questionId]: parseInt(value, 10),
    }))
  }, [])

  const handleSubmit = useCallback(async () => {
    if (!allQuestionsAnswered) return

    setIsSubmitting(true)
    setError(null)

    try {
      const answers: QuizAnswer[] = Object.entries(selectedAnswers).map(([questionId, answer]) => ({
        question_id: questionId,
        selected_answer: answer,
      }))

      const result = await agentService.submitQuizAnswers(moduleId, userId, answers)
      setQuizResult(result)

      const passed = result.score >= 70
      onComplete(result.score, passed)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to submit quiz')
      console.error('Failed to submit quiz:', err)
    } finally {
      setIsSubmitting(false)
    }
  }, [allQuestionsAnswered, selectedAnswers, moduleId, userId, onComplete])

  const handleRetry = useCallback(() => {
    setSelectedAnswers({})
    setQuizResult(null)
    setError(null)
  }, [])

  if (quizResult) {
    const passed = quizResult.score >= 70

    return (
      <Container>
        <SpaceBetween size="l">
          <Alert
            type={passed ? 'success' : 'warning'}
            header={passed ? 'Quiz Passed!' : 'Quiz Not Passed'}
          >
            {passed
              ? `Congratulations! You scored ${quizResult.score}% (${quizResult.correct_answers}/${quizResult.total_questions} correct).`
              : `You scored ${quizResult.score}% (${quizResult.correct_answers}/${quizResult.total_questions} correct). You need 70% to pass.`}
          </Alert>

          <SpaceBetween size="m">
            <Header variant="h3">Quiz Results</Header>
            {quizResult.results.map((result, index) => {
              const question = questions.find(q => q.id === result.question_id)
              if (!question) return null

              return (
                <Container key={result.question_id}>
                  <SpaceBetween size="s">
                    <Box variant="strong">
                      Question {index + 1}: {question.question}
                    </Box>

                    <Box padding={{ left: 'm' }}>
                      {question.options.map((option, optIndex) => {
                        const isSelected = result.selected_answer === optIndex
                        const isCorrect = result.correct_answer === optIndex
                        let style = {}
                        let indicator = ''

                        if (isCorrect) {
                          style = { color: 'green', fontWeight: 'bold' }
                          indicator = ' ✓ Correct'
                        } else if (isSelected && !isCorrect) {
                          style = { color: 'red', fontWeight: 'bold' }
                          indicator = ' ✗ Your answer'
                        }

                        return (
                          <Box key={optIndex} padding={{ bottom: 'xs' }}>
                            <span style={style}>
                              {option}
                              {indicator}
                            </span>
                          </Box>
                        )
                      })}
                    </Box>

                    {result.explanation && (
                      <Box fontSize="body-s" color="text-body-secondary" padding={{ left: 'm' }}>
                        <em>{result.explanation}</em>
                      </Box>
                    )}
                  </SpaceBetween>
                </Container>
              )
            })}
          </SpaceBetween>

          {!passed && (
            <Button onClick={handleRetry}>Retry Quiz</Button>
          )}
        </SpaceBetween>
      </Container>
    )
  }

  return (
    <Container>
      <SpaceBetween size="l">
        <Alert type="info">
          Complete this quiz to demonstrate your understanding. You need 70% to pass.
        </Alert>

        {error && (
          <Alert
            type="error"
            dismissible
            onDismiss={() => setError(null)}
          >
            {error}
          </Alert>
        )}

        <SpaceBetween size="m">
          {questions.map((question, index) => (
            <Container key={question.id}>
              <SpaceBetween size="s">
                <Box variant="strong">
                  Question {index + 1}: {question.question}
                </Box>

                <RadioGroup
                  value={selectedAnswers[question.id]?.toString() ?? ''}
                  onChange={({ detail }) => handleAnswerChange(question.id, detail.value)}
                  items={question.options.map((option, optIndex) => ({
                    value: optIndex.toString(),
                    label: option,
                  }))}
                />
              </SpaceBetween>
            </Container>
          ))}
        </SpaceBetween>

        <Button
          variant="primary"
          onClick={handleSubmit}
          loading={isSubmitting}
          disabled={!allQuestionsAnswered}
        >
          Submit Quiz
        </Button>

        {!allQuestionsAnswered && (
          <Box fontSize="body-s" color="text-status-warning">
            Please answer all questions before submitting
          </Box>
        )}
      </SpaceBetween>
    </Container>
  )
}
