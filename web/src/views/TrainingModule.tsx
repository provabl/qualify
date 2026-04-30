import { useState, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import Quiz from '@/components/training/Quiz'
import { agentService } from '@/services/agent'
import type { TrainingModule as TM, TrainingContent, TrainingSection, QuizQuestion } from '@/types/api'

export default function TrainingModule() {
  const { moduleName } = useParams<{ moduleName: string }>()
  const navigate = useNavigate()
  const [module, setModule] = useState<TM | null>(null)
  const [sections, setSections] = useState<TrainingSection[]>([])
  const [quiz, setQuiz] = useState<QuizQuestion[]>([])
  const [passingScore, setPassingScore] = useState(80)
  const [sectionIndex, setSectionIndex] = useState(0)
  const [showQuiz, setShowQuiz] = useState(false)
  const [completed, setCompleted] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!moduleName) return
    agentService.getTrainingModule(moduleName)
      .then(m => {
        setModule(m)
        if (m.content) {
          const c: TrainingContent = typeof m.content === 'string'
            ? JSON.parse(m.content) as TrainingContent
            : m.content
          setSections(c.sections ?? [])
          setQuiz(c.quiz ?? [])
          setPassingScore(c.passing_score ?? 80)
        }
      })
      .catch(e => setError(e instanceof Error ? e.message : 'Failed to load module'))
      .finally(() => setLoading(false))
  }, [moduleName])

  if (loading) return <div className="p-8 text-sm text-slate-500">Loading…</div>
  if (error)   return <div className="p-8 text-sm text-red-500">{error}</div>
  if (!module) return null

  const progress = showQuiz ? 100 : sections.length > 0 ? Math.round(((sectionIndex) / sections.length) * 100) : 0
  const currentSection = sections[sectionIndex]

  return (
    <div className="p-6 max-w-2xl">
      {/* Back link */}
      <button onClick={() => navigate('/training')} className="flex items-center gap-1 text-sm text-slate-500 hover:text-slate-700 mb-4">
        <ChevronLeft className="h-4 w-4" /> Back to Training
      </button>

      <h1 className="text-xl font-semibold text-slate-900 mb-1">{module.title}</h1>

      {/* Progress bar */}
      <div className="mb-6">
        <div className="flex justify-between text-xs text-slate-400 mb-1">
          <span>{showQuiz ? 'Quiz' : `Section ${sectionIndex + 1} of ${sections.length}`}</span>
          <span>{progress}%</span>
        </div>
        <div className="h-1.5 rounded-full bg-slate-100 overflow-hidden">
          <div className="h-full rounded-full bg-brand-500 transition-all" style={{ width: `${progress}%` }} />
        </div>
      </div>

      {completed ? (
        <div className="rounded-lg border border-green-200 bg-green-50 p-6 text-center">
          <p className="text-lg font-semibold text-green-800">Module Complete 🎉</p>
          <p className="text-sm text-green-600 mt-1">Your access tags have been updated.</p>
          <button onClick={() => navigate('/training')} className="mt-4 text-sm text-brand-600 hover:text-brand-700 font-medium">
            Back to Training →
          </button>
        </div>
      ) : showQuiz ? (
        <div className="rounded-lg border border-slate-200 bg-white p-6">
          <h2 className="text-base font-semibold text-slate-900 mb-4">Knowledge Check</h2>
          <Quiz
            moduleId={moduleName ?? ''}
            userId="00000000-0000-0000-0000-000000000001"
            questions={quiz}
            passingScore={passingScore}
            onComplete={(_, passed) => { if (passed) setCompleted(true) }}
          />
        </div>
      ) : currentSection ? (
        <div className="rounded-lg border border-slate-200 bg-white p-6">
          <h2 className="text-base font-semibold text-slate-900 mb-3">{currentSection.title}</h2>
          <div className="prose prose-sm prose-slate max-w-none text-slate-600 whitespace-pre-wrap leading-relaxed">
            {currentSection.content}
          </div>
          <div className="flex items-center justify-between mt-6 pt-4 border-t border-slate-100">
            <button
              onClick={() => setSectionIndex(i => i - 1)}
              disabled={sectionIndex === 0}
              className="flex items-center gap-1 text-sm text-slate-500 hover:text-slate-700 disabled:opacity-30"
            >
              <ChevronLeft className="h-4 w-4" /> Previous
            </button>
            {sectionIndex < sections.length - 1 ? (
              <button
                onClick={() => setSectionIndex(i => i + 1)}
                className="flex items-center gap-1 text-sm font-medium text-brand-600 hover:text-brand-700"
              >
                Next <ChevronRight className="h-4 w-4" />
              </button>
            ) : (
              <button
                onClick={() => quiz.length > 0 ? setShowQuiz(true) : setCompleted(true)}
                className="flex items-center gap-1 text-sm font-medium text-brand-600 hover:text-brand-700"
              >
                {quiz.length > 0 ? 'Start Quiz' : 'Complete'} <ChevronRight className="h-4 w-4" />
              </button>
            )}
          </div>
        </div>
      ) : (
        <p className="text-sm text-slate-400">No content available for this module.</p>
      )}
    </div>
  )
}
