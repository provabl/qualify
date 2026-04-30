import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { X, BookOpen, ShieldCheck, Zap, ChevronRight } from 'lucide-react'
import { agentService } from '@/services/agent'
import { cn } from '@/lib/utils'
import type { TrainingModule, UserProfile } from '@/types/api'

interface Props {
  visible: boolean
  userId: string
  trainingModules: TrainingModule[]
  onDismiss: () => void
  onComplete: () => void
}

const steps = [
  {
    icon: BookOpen,
    title: 'Complete Required Training',
    desc: 'Finish compliance training modules to unlock access to sensitive data environments.',
  },
  {
    icon: ShieldCheck,
    title: 'Get Your Access Tags',
    desc: 'qualify writes IAM tags to your role on completion. attest evaluates these in real time.',
  },
  {
    icon: Zap,
    title: 'Access Unlocked',
    desc: "Once your tags are set, you'll have access to the AWS resources your institution has approved.",
  },
]

export default function OnboardingWizard({ visible, userId, trainingModules, onDismiss, onComplete }: Props) {
  const navigate = useNavigate()
  const [step, setStep] = useState(0)
  const [completing, setCompleting] = useState(false)

  if (!visible) return null

  async function finish() {
    setCompleting(true)
    try {
      await agentService.updateUserProfile(userId, {
        preferences: { has_completed_onboarding: true, show_training_reminders: true },
      } satisfies Partial<UserProfile>)
    } catch {
      // non-fatal
    } finally {
      setCompleting(false)
      onComplete()
    }
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 backdrop-blur-sm">
      <div className="relative w-full max-w-lg mx-4 bg-white rounded-xl shadow-xl overflow-hidden">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-200">
          <h2 className="text-base font-semibold text-slate-900">Welcome to qualify</h2>
          <button onClick={onDismiss} className="text-slate-400 hover:text-slate-600 transition-colors">
            <X className="h-4 w-4" />
          </button>
        </div>

        {/* Step indicator */}
        <div className="flex gap-1.5 px-6 pt-5">
          {steps.map((_, i) => (
            <div key={i} className={cn('h-1 flex-1 rounded-full transition-colors', i <= step ? 'bg-brand-500' : 'bg-slate-200')} />
          ))}
        </div>

        {/* Step content */}
        <div className="px-6 py-5">
          {step < steps.length ? (
            <div className="flex gap-4">
              {(() => {
                const { icon: Icon, title, desc } = steps[step]
                return (
                  <>
                    <div className="h-10 w-10 rounded-lg bg-brand-50 flex items-center justify-center flex-none">
                      <Icon className="h-5 w-5 text-brand-600" />
                    </div>
                    <div>
                      <p className="font-medium text-slate-900">{title}</p>
                      <p className="text-sm text-slate-500 mt-1 leading-relaxed">{desc}</p>
                    </div>
                  </>
                )
              })()}
            </div>
          ) : (
            <div>
              <p className="font-medium text-slate-900 mb-3">Your required training modules</p>
              {trainingModules.length === 0 ? (
                <p className="text-sm text-slate-400">No modules assigned yet — check back after your SRE admin configures your environment.</p>
              ) : (
                <div className="space-y-1.5 max-h-48 overflow-y-auto">
                  {trainingModules.slice(0, 6).map(m => (
                    <div key={m.id ?? m.name} className="flex items-center justify-between p-2.5 rounded-lg border border-slate-200 text-sm">
                      <span className="text-slate-700">{m.title}</span>
                      {m.estimated_minutes && <span className="text-xs text-slate-400">{m.estimated_minutes} min</span>}
                    </div>
                  ))}
                </div>
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between px-6 py-4 border-t border-slate-200 bg-slate-50">
          <button onClick={onDismiss} className="text-sm text-slate-500 hover:text-slate-700">
            Skip for now
          </button>
          {step < steps.length ? (
            <button
              onClick={() => setStep(s => s + 1)}
              className="flex items-center gap-1.5 px-4 py-1.5 text-sm font-medium text-white bg-brand-600 hover:bg-brand-700 rounded-md transition-colors"
            >
              Next <ChevronRight className="h-4 w-4" />
            </button>
          ) : (
            <div className="flex gap-2">
              <button
                onClick={() => { finish(); navigate('/training') }}
                disabled={completing}
                className="px-4 py-1.5 text-sm font-medium text-white bg-brand-600 hover:bg-brand-700 rounded-md disabled:opacity-40 transition-colors"
              >
                Start Training
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
