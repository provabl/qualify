import { GraduationCap } from 'lucide-react'

export default function Home() {
  return (
    <div className="p-8 max-w-2xl">
      <div className="flex items-center gap-3 mb-3">
        <GraduationCap className="h-7 w-7 text-brand-600" />
        <h1 className="text-2xl font-semibold text-slate-900">Welcome to qualify</h1>
      </div>
      <p className="text-slate-600 leading-relaxed">
        qualify provides compliance training and per-researcher access gating for AWS
        Secure Research Environments. Complete required training modules to unlock
        access to sensitive data environments.
      </p>
      <div className="mt-6 p-4 bg-brand-50 border border-brand-100 rounded-lg text-sm text-brand-700">
        Use the CLI for the full training experience:{' '}
        <code className="font-mono bg-white px-1.5 py-0.5 rounded border border-brand-200">
          qualify train required
        </code>
      </div>
    </div>
  )
}
