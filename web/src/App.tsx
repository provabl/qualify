import { useState, useEffect } from 'react'
import { Routes, Route, NavLink, useLocation } from 'react-router-dom'
import { LayoutDashboard, FolderOpen, BookOpen, Home as HomeIcon, GraduationCap } from 'lucide-react'
import AgentStatus from '@/components/common/AgentStatus'
import OnboardingWizard from '@/components/onboarding/OnboardingWizard'
import Home from '@/views/Home'
import Dashboard from '@/views/Dashboard'
import S3 from '@/views/S3'
import Training from '@/views/Training'
import TrainingModule from '@/views/TrainingModule'
import { agentService } from '@/services/agent'
import { cn } from '@/lib/utils'
import type { TrainingModule as TrainingModuleType } from '@/types/api'

const navItems = [
  { href: '/',          label: 'Home',     icon: HomeIcon },
  { href: '/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { href: '/s3',        label: 'S3',        icon: FolderOpen },
  { href: '/training',  label: 'Training',  icon: BookOpen },
]

export default function App() {
  const location = useLocation()
  const [userId, setUserId] = useState<string>('')
  const [showOnboarding, setShowOnboarding] = useState(false)
  const [trainingModules, setTrainingModules] = useState<TrainingModuleType[]>([])
  const [onboardingChecked, setOnboardingChecked] = useState(false)

  useEffect(() => {
    initAuth()
  }, [])

  async function initAuth() {
    try {
      // Acquire a token if we don't have one.
      // In dev mode the backend provides /api/auth/dev-token; in production
      // this would be replaced with an OIDC redirect flow.
      if (!agentService.getToken()) {
        const res = await fetch(
          `${import.meta.env.VITE_BACKEND_URL ?? 'http://127.0.0.1:8081'}/api/auth/dev-token`
        )
        if (res.ok) {
          const data = await res.json()
          agentService.setToken(data.token)
          setUserId(data.user_id)
        }
      } else {
        // Resolve current user from existing token.
        const me = await agentService.getMe()
        setUserId(me.user_id)
      }

      // Load onboarding state and training modules.
      const [profile, modules] = await Promise.all([
        agentService.getUserProfile(),
        agentService.listTrainingModules(),
      ])
      setTrainingModules(modules)
      if (!profile.preferences.has_completed_onboarding) setShowOnboarding(true)
    } catch {
      // Backend not running — continue without auth/onboarding.
    } finally {
      setOnboardingChecked(true)
    }
  }

  return (
    <div className="flex h-screen overflow-hidden bg-slate-50">
      {/* Sidebar */}
      <aside className="w-56 flex-none bg-white border-r border-slate-200 flex flex-col">
        {/* Brand */}
        <div className="flex items-center gap-2 px-4 py-4 border-b border-slate-200">
          <GraduationCap className="h-5 w-5 text-brand-600" />
          <span className="font-semibold text-slate-900">qualify</span>
        </div>

        {/* Navigation */}
        <nav className="flex-1 px-2 py-3 space-y-0.5">
          {navItems.map(({ href, label, icon: Icon }) => {
            const active = href === '/'
              ? location.pathname === '/'
              : location.pathname.startsWith(href)
            return (
              <NavLink
                key={href}
                to={href}
                className={cn(
                  'flex items-center gap-2.5 px-3 py-2 rounded-md text-sm font-medium transition-colors',
                  active
                    ? 'bg-brand-50 text-brand-700'
                    : 'text-slate-600 hover:bg-slate-100 hover:text-slate-900'
                )}
              >
                <Icon className="h-4 w-4 flex-none" />
                {label}
              </NavLink>
            )
          })}
        </nav>

        {/* Agent status + user at bottom */}
        <div className="border-t border-slate-200 p-3 space-y-1">
          <AgentStatus />
          {userId && (
            <p className="text-xs text-slate-400 truncate" title={userId}>
              {userId.substring(0, 8)}…
            </p>
          )}
        </div>
      </aside>

      {/* Main content */}
      <main className="flex-1 overflow-auto">
        <Routes>
          <Route path="/"                      element={<Home />} />
          <Route path="/dashboard"             element={<Dashboard userId={userId} />} />
          <Route path="/s3"                    element={<S3 />} />
          <Route path="/training"              element={<Training />} />
          <Route path="/training/:moduleName"  element={<TrainingModule />} />
        </Routes>
      </main>

      {onboardingChecked && (
        <OnboardingWizard
          visible={showOnboarding}
          userId={userId}
          trainingModules={trainingModules}
          onDismiss={() => setShowOnboarding(false)}
          onComplete={() => setShowOnboarding(false)}
        />
      )}
    </div>
  )
}
