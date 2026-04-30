import { useState } from 'react'
import { CheckCircle2, XCircle } from 'lucide-react'
import { agentService } from '@/services/agent'
import { cn } from '@/lib/utils'
import type { QuizQuestion, QuizResponse } from '@/types/api'

interface Props {
  moduleId: string
  userId: string
  questions: QuizQuestion[]
  passingScore?: number
  onComplete: (score: number, passed: boolean) => void
}

export default function Quiz({ moduleId, userId, questions, passingScore = 70, onComplete }: Props) {
  const [answers, setAnswers] = useState<Record<number, number>>({})
  const [submitted, setSubmitted] = useState(false)
  const [result, setResult] = useState<QuizResponse | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const allAnswered = Object.keys(answers).length === questions.length

  async function submit() {
    if (!allAnswered) return
    setSubmitting(true)
    setError(null)
    try {
      const quizAnswers = questions.map((q, i) => ({
        question_id: q.id,
        selected_answer: answers[i] ?? 0,
      }))
      const response = await agentService.submitQuizAnswers(moduleId, userId, quizAnswers)
      setResult(response)
      setSubmitted(true)
      const passed = response.score >= passingScore
      onComplete(response.score, passed)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to submit quiz')
    } finally {
      setSubmitting(false)
    }
  }

  if (submitted && result) {
    const passed = result.score >= passingScore
    return (
      <div className="space-y-4">
        <div className={cn('flex items-center gap-3 p-4 rounded-lg border', passed ? 'bg-green-50 border-green-200' : 'bg-red-50 border-red-200')}>
          {passed
            ? <CheckCircle2 className="h-5 w-5 text-green-600 flex-none" />
            : <XCircle className="h-5 w-5 text-red-600 flex-none" />}
          <div>
            <p className={cn('font-medium', passed ? 'text-green-800' : 'text-red-800')}>
              {passed ? 'Quiz Passed!' : 'Quiz Not Passed'}
            </p>
            <p className={cn('text-sm mt-0.5', passed ? 'text-green-600' : 'text-red-600')}>
              You scored {result.score}%
              {!passed && ` — You need ${passingScore}% to pass.`}
            </p>
          </div>
        </div>
        {!passed && (
          <button onClick={() => { setSubmitted(false); setAnswers({}) }} className="text-sm text-brand-600 hover:text-brand-700 font-medium">
            Retry Quiz
          </button>
        )}
        <div className="space-y-3 mt-2">
          {result.results?.map((r, qi) => {
            const q = questions.find(q => q.id === r.question_id) ?? questions[qi]
            if (!q) return null
            return (
              <div key={r.question_id} className={cn('p-3 rounded-lg border text-sm', r.correct ? 'border-green-100 bg-green-50' : 'border-red-100 bg-red-50')}>
                <p className="font-medium text-slate-800 mb-2">Question {qi + 1}: {q.question}</p>
                {q.options.map((opt, oi) => (
                  <div key={oi} className={cn('px-2 py-1 rounded',
                    oi === r.correct_answer && 'text-green-700 font-medium',
                    oi === r.selected_answer && !r.correct && 'text-red-700 line-through',
                  )}>
                    {oi === r.correct_answer ? '✓' : oi === r.selected_answer ? '✗' : ' '} {opt}
                  </div>
                ))}
                {r.explanation && <p className="mt-2 text-xs text-slate-500 italic">{r.explanation}</p>}
              </div>
            )
          })}
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-5">
      <p className="text-sm text-slate-500">{questions.length} questions — {passingScore}% to pass</p>
      {questions.map((q, qi) => (
        <div key={q.id} className="space-y-2">
          <p className="text-sm font-medium text-slate-900">Question {qi + 1}: {q.question}</p>
          <div className="space-y-1.5" role="radiogroup">
            {q.options.map((opt, oi) => (
              <label
                key={oi}
                className={cn(
                  'flex items-center gap-3 p-2.5 rounded-lg border cursor-pointer transition-colors text-sm',
                  answers[qi] === oi
                    ? 'border-brand-300 bg-brand-50 text-brand-900'
                    : 'border-slate-200 hover:border-slate-300 text-slate-700'
                )}
              >
                <input
                  type="radio"
                  name={`q-${qi}`}
                  value={oi}
                  aria-label={opt}
                  checked={answers[qi] === oi}
                  onChange={() => setAnswers(prev => ({ ...prev, [qi]: oi }))}
                  className="accent-brand-600"
                />
                {opt}
              </label>
            ))}
          </div>
        </div>
      ))}

      {!allAnswered && (
        <p className="text-sm text-slate-400">Please answer all questions before submitting</p>
      )}

      {error && <p className="text-sm text-red-500">{error}</p>}

      <button
        onClick={submit}
        disabled={submitting || !allAnswered}
        className="px-4 py-2 bg-brand-600 text-white text-sm font-medium rounded-md hover:bg-brand-700 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
      >
        {submitting ? 'Submitting…' : 'Submit Quiz'}
      </button>
    </div>
  )
}
