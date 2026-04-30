import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { RefreshCw } from 'lucide-react'
import { agentService } from '@/services/agent'
import { cn } from '@/lib/utils'
import type { DashboardStats, ActivityItem } from '@/types/api'

const USER_ID = '00000000-0000-0000-0000-000000000001'

const activityLabel: Record<string, string> = {
  module_started:    'Started Module',
  module_completed:  'Completed Module',
  quiz_passed:       'Passed Quiz',
  quiz_failed:       'Failed Quiz',
  operation_blocked: 'Operation Blocked',
}

const activityColor: Record<string, string> = {
  module_completed:  'bg-green-100 text-green-700',
  quiz_passed:       'bg-green-100 text-green-700',
  module_started:    'bg-blue-100 text-blue-700',
  quiz_failed:       'bg-red-100 text-red-700',
  operation_blocked: 'bg-red-100 text-red-700',
}

function formatTs(ts: string) {
  const diff = (Date.now() - new Date(ts).getTime()) / 1000
  if (diff < 60)  return 'Just now'
  if (diff < 3600) return `${Math.floor(diff / 60)}m ago`
  if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`
  return `${Math.floor(diff / 86400)}d ago`
}

export default function Dashboard() {
  const navigate = useNavigate()
  const [stats, setStats] = useState<DashboardStats | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => { load() }, [])

  async function load() {
    setLoading(true)
    setError(null)
    try {
      setStats(await agentService.getDashboardStats(USER_ID))
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Failed to load dashboard')
    } finally {
      setLoading(false)
    }
  }

  if (loading) return (
    <div className="p-8 text-sm text-slate-500">Loading dashboard...</div>
  )

  if (error) return (
    <div className="p-8 max-w-lg">
      <div className="rounded-lg border border-red-200 bg-red-50 p-4">
        <p className="font-medium text-red-700">Could not connect to qualify backend</p>
        <p className="mt-1 text-sm text-red-600">{error}</p>
        <p className="mt-2 text-xs text-red-500">Use the CLI in the meantime:{' '}
          <code className="font-mono">qualify train status</code>
        </p>
        <button onClick={load} className="mt-3 flex items-center gap-1.5 text-sm text-red-700 hover:text-red-800 font-medium">
          <RefreshCw className="h-3.5 w-3.5" /> Retry
        </button>
      </div>
    </div>
  )

  if (!stats) return null

  const { training_summary: ts, recent_activity, available_operations } = stats

  return (
    <div className="p-6 space-y-6 max-w-4xl">
      <div className="flex items-center justify-between">
        <h1 className="text-xl font-semibold text-slate-900">Dashboard</h1>
        <button onClick={() => navigate('/training')} className="text-sm text-brand-600 hover:text-brand-700 font-medium">
          View All Training
        </button>
      </div>

      {/* Training summary */}
      <div className="rounded-lg border border-slate-200 bg-white p-5">
        <h2 className="text-sm font-semibold text-slate-700 mb-4">Training Progress</h2>
        <div className="grid grid-cols-4 gap-4 mb-4">
          {[
            { label: 'Total Modules', value: ts.total_modules, color: 'text-slate-900' },
            { label: 'Completed',     value: ts.completed,     color: 'text-green-600' },
            { label: 'In Progress',   value: ts.in_progress,   color: 'text-blue-600' },
            { label: 'Not Started',   value: ts.not_started,   color: 'text-slate-500' },
          ].map(({ label, value, color }) => (
            <div key={label}>
              <p className="text-xs text-slate-500">{label}</p>
              <p className={cn('text-2xl font-semibold mt-0.5', color)}>{value}</p>
            </div>
          ))}
        </div>
        <div>
          <div className="flex justify-between text-xs text-slate-500 mb-1">
            <span>Overall Completion</span>
            <span>{ts.completion_percentage}%
              {ts.average_score !== undefined && ` · Average quiz score: ${ts.average_score}%`}
            </span>
          </div>
          <div className="h-2 rounded-full bg-slate-100 overflow-hidden">
            <div className="h-full rounded-full bg-brand-500 transition-all" style={{ width: `${ts.completion_percentage}%` }} />
          </div>
        </div>
      </div>

      <div className="grid grid-cols-2 gap-6">
        {/* Recent Activity */}
        <div className="rounded-lg border border-slate-200 bg-white p-5">
          <h2 className="text-sm font-semibold text-slate-700 mb-3">Recent Activity</h2>
          {recent_activity.length === 0 ? (
            <p className="text-sm text-slate-400">No recent activity to display. Start a training module to see your progress here.</p>
          ) : (
            <div className="space-y-2.5">
              {recent_activity.map((a: ActivityItem) => (
                <div key={a.id} className="flex items-start justify-between gap-2">
                  <div className="flex items-center gap-2 min-w-0">
                    <span className={cn('text-xs font-medium px-2 py-0.5 rounded-full flex-none', activityColor[a.type] ?? 'bg-slate-100 text-slate-600')}>
                      {activityLabel[a.type] ?? a.type}
                    </span>
                    {a.module_name && <span className="text-xs text-slate-500 truncate">{a.module_name}</span>}
                  </div>
                  <div className="flex items-center gap-1.5 flex-none">
                    {a.score !== undefined && (
                      <span className={cn('text-xs font-medium', a.score >= 70 ? 'text-green-600' : 'text-red-600')}>{a.score}%</span>
                    )}
                    <span className="text-xs text-slate-400">{formatTs(a.timestamp)}</span>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* AWS Operations */}
        <div className="rounded-lg border border-slate-200 bg-white p-5">
          <h2 className="text-sm font-semibold text-slate-700 mb-3">AWS Operations</h2>
          {available_operations.unlocked.length > 0 && (
            <div className="mb-3">
              <p className="text-xs font-medium text-slate-500 mb-1.5">Unlocked Operations</p>
              <div className="space-y-1">
                {available_operations.unlocked.map((op: string) => (
                  <div key={op} className="flex items-center gap-1.5 text-xs text-green-700">
                    <span className="text-green-500">✓</span> {op}
                  </div>
                ))}
              </div>
            </div>
          )}
          {available_operations.locked.length > 0 && (
            <div>
              <p className="text-xs font-medium text-slate-500 mb-1.5">Locked Operations</p>
          <p className="text-xs text-slate-400 mb-1.5">Complete required training to unlock these operations</p>
              <div className="space-y-1">
                {available_operations.locked.map((op: string) => (
                  <div key={op} className="flex items-center gap-1.5 text-xs text-slate-400">
                    <span>🔒</span> {op}
                  </div>
                ))}
              </div>
            </div>
          )}
          {available_operations.unlocked.length === 0 && available_operations.locked.length === 0 && (
            <p className="text-sm text-slate-400">
              {ts.completed === 0
                ? 'Complete training modules to unlock AWS operations. Run: qualify train required'
                : 'Operations are determined by your active SRE environments. Contact your SRE admin.'}
            </p>
          )}
        </div>
      </div>
    </div>
  )
}
