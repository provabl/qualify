import { useAgent } from '@/contexts/AgentContext'
import { cn } from '@/lib/utils'

export default function AgentStatus() {
  const agent = useAgent()

  const colors: Record<string, string> = {
    connected:    'bg-green-500',
    disconnected: 'bg-slate-400',
    checking:     'bg-amber-400 animate-pulse',
  }

  const labels: Record<string, string> = {
    connected:    'Agent connected',
    disconnected: 'Agent offline',
    checking:     'Checking…',
  }

  return (
    <div className="flex items-center gap-2 text-xs text-slate-500">
      <span className={cn('inline-block h-2 w-2 rounded-full flex-none', colors[agent.status] ?? 'bg-slate-400')} />
      <span>{labels[agent.status] ?? 'Agent unknown'}</span>
    </div>
  )
}
