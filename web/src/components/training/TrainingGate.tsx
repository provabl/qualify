import { useNavigate } from 'react-router-dom'
import { Lock, BookOpen } from 'lucide-react'
import type { TrainingModule } from '@/types/api'

interface Props {
  requiredModules: TrainingModule[]
  operationName: string
}

export default function TrainingGate({ requiredModules, operationName }: Props) {
  const navigate = useNavigate()

  return (
    <div className="rounded-lg border border-amber-200 bg-amber-50 p-5">
      <div className="flex items-start gap-3">
        <Lock className="h-5 w-5 text-amber-600 flex-none mt-0.5" />
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium text-amber-800">Training required for {operationName}</p>
          <p className="text-sm text-amber-600 mt-1">
            Complete the following modules to unlock this operation:
          </p>
          <div className="mt-3 space-y-1.5">
            {requiredModules.map(m => (
              <div key={m.id ?? m.name} className="flex items-center justify-between gap-3 p-2 bg-white rounded border border-amber-100">
                <div className="flex items-center gap-2 min-w-0">
                  <BookOpen className="h-3.5 w-3.5 text-amber-500 flex-none" />
                  <span className="text-sm text-slate-700 truncate">{m.title}</span>
                  {m.estimated_minutes && (
                    <span className="text-xs text-slate-400 flex-none">{m.estimated_minutes} min</span>
                  )}
                </div>
                <button
                  onClick={() => navigate(`/training/${m.name}`)}
                  className="text-xs text-brand-600 hover:text-brand-700 font-medium flex-none"
                >
                  Start →
                </button>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
