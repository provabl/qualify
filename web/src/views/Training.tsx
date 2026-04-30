import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { CheckCircle2, Circle, Clock } from 'lucide-react'
import { agentService } from '@/services/agent'
import { cn } from '@/lib/utils'
import type { TrainingModule } from '@/types/api'

export default function Training() {
  const navigate = useNavigate()
  const [modules, setModules] = useState<TrainingModule[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    agentService.listTrainingModules()
      .then(setModules)
      .catch(e => setError(e instanceof Error ? e.message : 'Failed to load modules'))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return <div className="p-8 text-sm text-slate-500">Loading…</div>

  const StatusIcon = ({ status }: { status?: string }) => {
    if (status === 'completed')   return <CheckCircle2 className="h-4 w-4 text-green-500 flex-none" />
    if (status === 'in_progress') return <Clock className="h-4 w-4 text-blue-500 flex-none" />
    return <Circle className="h-4 w-4 text-slate-300 flex-none" />
  }

  const StatusBadge = ({ status }: { status?: string }) => {
    if (status === 'completed')   return <span className="text-xs font-medium text-green-600 bg-green-50 px-2 py-0.5 rounded-full">Completed</span>
    if (status === 'in_progress') return <span className="text-xs font-medium text-blue-600 bg-blue-50 px-2 py-0.5 rounded-full">In Progress</span>
    return <span className="text-xs font-medium text-slate-500 bg-slate-100 px-2 py-0.5 rounded-full">Not Started</span>
  }

  return (
    <div className="p-6 max-w-3xl">
      <h1 className="text-xl font-semibold text-slate-900 mb-6">Training Modules</h1>

      {error && (
        <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-600">{error}</div>
      )}

      {modules.length === 0 && !error ? (
        <p className="text-sm text-slate-400 py-8 text-center">No training modules available.</p>
      ) : (
        <div className="space-y-2">
          {modules.map(m => (
            <div
              key={m.id ?? m.name}
              className={cn(
                'flex items-center justify-between p-4 rounded-lg border bg-white cursor-pointer transition-colors',
                m.status === 'completed' ? 'border-green-100 hover:border-green-200' : 'border-slate-200 hover:border-slate-300'
              )}
              onClick={() => navigate(`/training/${m.name}`)}
            >
              <div className="flex items-center gap-3 min-w-0">
                <StatusIcon status={m.status} />
                <div className="min-w-0">
                  <p className="text-sm font-medium text-slate-900 truncate">{m.title}</p>
                  {m.description && <p className="text-xs text-slate-400 truncate mt-0.5">{m.description}</p>}
                </div>
              </div>
              <div className="flex items-center gap-3 flex-none ml-3">
                {m.estimated_minutes && <span className="text-xs text-slate-400">{m.estimated_minutes} min</span>}
                <StatusBadge status={m.status} />
                <button
                  className="text-sm text-brand-600 hover:text-brand-700 font-medium px-3 py-1 rounded-md hover:bg-brand-50 transition-colors"
                  onClick={e => { e.stopPropagation(); navigate(`/training/${m.name}`) }}
                >
                  {m.status === 'completed' ? 'Review' : 'Start Training'}
                </button>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
