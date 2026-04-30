import { useState, useEffect } from 'react'
import { FolderOpen, Plus, AlertCircle } from 'lucide-react'
import TrainingGate from '@/components/training/TrainingGate'
import { agentService } from '@/services/agent'
import { cn } from '@/lib/utils'
import type { S3Bucket, TrainingModule, PolicyDecision } from '@/types/api'

function isPolicyDecision(v: unknown): v is PolicyDecision {
  return typeof v === 'object' && v !== null && 'action' in v
}

export default function S3() {
  const [buckets, setBuckets] = useState<S3Bucket[]>([])
  const [error] = useState<string | null>(null)
  const [requiredModules, setRequiredModules] = useState<TrainingModule[]>([])
  const [showCreate, setShowCreate] = useState(false)
  const [bucketName, setBucketName] = useState('')
  const [region, setRegion] = useState('us-east-1')
  const [creating, setCreating] = useState(false)
  const [createMsg, setCreateMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null)

  // S3 list is agent-provided; skip if agent not running
  useEffect(() => { checkTrainingGate() }, [])

  async function checkTrainingGate() {
    // A training gate check is triggered by attempting a minimal operation.
    // If training is required, agentService returns a PolicyDecision.
    try {
      const result = await agentService.createBucket({
        bucket_name: '__gate_check__',
        region: 'us-east-1',
        encryption: { type: 'AES256' },
        versioning_enabled: false,
        profile: 'default',
      })
      if (isPolicyDecision(result) && result.required_modules) {
        setRequiredModules(result.required_modules as TrainingModule[])
      }
    } catch {
      // agent not running — show form, let backend handle
    }
  }

  async function createBucket() {
    if (!bucketName.trim()) return
    setCreating(true)
    setCreateMsg(null)
    try {
      const result = await agentService.createBucket({
        bucket_name: bucketName.trim(),
        region,
        encryption: { type: 'AES256' },
        versioning_enabled: false,
        profile: 'default',
      })
      if (isPolicyDecision(result)) {
        setCreateMsg({ type: 'error', text: result.reason ?? 'Operation blocked by training gate.' })
        if (result.required_modules) setRequiredModules(result.required_modules as TrainingModule[])
      } else {
        setCreateMsg({ type: 'success', text: `Bucket "${bucketName}" created successfully.` })
        setBucketName('')
        setShowCreate(false)
        setBuckets(prev => [...prev, result as S3Bucket])
      }
    } catch (e) {
      setCreateMsg({ type: 'error', text: e instanceof Error ? e.message : 'Failed to create bucket' })
    } finally {
      setCreating(false)
    }
  }

  if (requiredModules.length > 0) {
    return (
      <div className="p-6 max-w-2xl">
        <h1 className="text-xl font-semibold text-slate-900 mb-4">S3 Buckets</h1>
        <TrainingGate requiredModules={requiredModules} operationName="S3 operations" />
      </div>
    )
  }

  return (
    <div className="p-6 max-w-3xl">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-semibold text-slate-900">S3 Buckets</h1>
        <button
          onClick={() => setShowCreate(s => !s)}
          className="flex items-center gap-1.5 px-3 py-1.5 text-sm font-medium text-white bg-brand-600 hover:bg-brand-700 rounded-md transition-colors"
        >
          <Plus className="h-4 w-4" /> Create Bucket
        </button>
      </div>

      {error && (
        <div className="mb-4 flex items-start gap-2 p-3 bg-red-50 border border-red-200 rounded-lg text-sm text-red-600">
          <AlertCircle className="h-4 w-4 flex-none mt-0.5" /> {error}
        </div>
      )}

      {createMsg && (
        <div className={cn('mb-4 p-3 rounded-lg border text-sm', createMsg.type === 'success' ? 'bg-green-50 border-green-200 text-green-700' : 'bg-red-50 border-red-200 text-red-600')}>
          {createMsg.text}
        </div>
      )}

      {showCreate && (
        <div className="mb-4 p-4 rounded-lg border border-slate-200 bg-white space-y-3">
          <h2 className="text-sm font-semibold text-slate-700">New bucket</h2>
          <div>
            <label className="block text-xs font-medium text-slate-600 mb-1">Bucket name</label>
            <input
              type="text"
              value={bucketName}
              onChange={e => setBucketName(e.target.value)}
              placeholder="my-research-data"
              className="w-full px-3 py-1.5 text-sm border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-brand-500 focus:border-transparent"
            />
            <p className="text-xs text-slate-400 mt-1">3–63 characters, lowercase letters, numbers, and hyphens.</p>
          </div>
          <div>
            <label className="block text-xs font-medium text-slate-600 mb-1">Region</label>
            <select
              value={region}
              onChange={e => setRegion(e.target.value)}
              className="w-full px-3 py-1.5 text-sm border border-slate-300 rounded-md focus:outline-none focus:ring-2 focus:ring-brand-500"
            >
              {['us-east-1','us-east-2','us-west-1','us-west-2','eu-west-1','eu-central-1','ap-southeast-1'].map(r => (
                <option key={r} value={r}>{r}</option>
              ))}
            </select>
          </div>
          <div className="flex gap-2 pt-1">
            <button onClick={createBucket} disabled={creating || !bucketName.trim()}
              className="px-4 py-1.5 text-sm font-medium text-white bg-brand-600 hover:bg-brand-700 rounded-md disabled:opacity-40 disabled:cursor-not-allowed transition-colors">
              {creating ? 'Creating…' : 'Create'}
            </button>
            <button onClick={() => setShowCreate(false)} className="px-4 py-1.5 text-sm text-slate-600 hover:text-slate-800">Cancel</button>
          </div>
        </div>
      )}

      {buckets.length === 0 ? (
        <div className="text-center py-12 text-slate-400">
          <FolderOpen className="h-8 w-8 mx-auto mb-2 opacity-40" />
          <p className="text-sm">No buckets yet. Create your first bucket above.</p>
        </div>
      ) : (
        <div className="space-y-1">
          {buckets.map(b => (
            <div key={b.bucket_name} className="flex items-center justify-between p-3 rounded-lg border border-slate-200 bg-white hover:border-slate-300 transition-colors">
              <div className="flex items-center gap-2.5">
                <FolderOpen className="h-4 w-4 text-slate-400 flex-none" />
                <div>
                  <p className="text-sm font-medium text-slate-900">{b.bucket_name}</p>
                  {b.region && <p className="text-xs text-slate-400">{b.region}</p>}
                </div>
              </div>
              {b.created_at && <span className="text-xs text-slate-400">{new Date(b.created_at).toLocaleDateString()}</span>}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
