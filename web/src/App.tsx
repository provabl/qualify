import { useState, useEffect } from 'react'
import { Routes, Route, useNavigate, useLocation } from 'react-router-dom'
import AppLayout from '@cloudscape-design/components/app-layout'
import SideNavigation, { SideNavigationProps } from '@cloudscape-design/components/side-navigation'
import TopNavigation from '@cloudscape-design/components/top-navigation'
import AgentStatus from '@/components/common/AgentStatus'
import OnboardingWizard from '@/components/onboarding/OnboardingWizard'
import Home from '@/views/Home'
import Dashboard from '@/views/Dashboard'
import S3 from '@/views/S3'
import Training from '@/views/Training'
import TrainingModule from '@/views/TrainingModule'
import { agentService } from '@/services/agent'
import type { TrainingModule as TrainingModuleType } from '@/types/api'

const USER_ID = '00000000-0000-0000-0000-000000000001'

const navigationItems: SideNavigationProps['items'] = [
  { type: 'link', text: 'Home', href: '/' },
  { type: 'link', text: 'Dashboard', href: '/dashboard' },
  { type: 'divider' },
  { type: 'link', text: 'S3', href: '/s3' },
  { type: 'link', text: 'Training', href: '/training' }
]

export default function App() {
  const navigate = useNavigate()
  const location = useLocation()
  const [navigationOpen, setNavigationOpen] = useState(true)
  const [showOnboarding, setShowOnboarding] = useState(false)
  const [trainingModules, setTrainingModules] = useState<TrainingModuleType[]>([])
  const [onboardingChecked, setOnboardingChecked] = useState(false)

  useEffect(() => {
    checkOnboardingStatus()
  }, [])

  async function checkOnboardingStatus() {
    try {
      const [profile, modules] = await Promise.all([
        agentService.getUserProfile(USER_ID),
        agentService.listTrainingModules()
      ])

      setTrainingModules(modules)

      if (!profile.preferences.has_completed_onboarding) {
        setShowOnboarding(true)
      }
    } catch (error) {
      console.error('Failed to check onboarding status:', error)
    } finally {
      setOnboardingChecked(true)
    }
  }

  function handleOnboardingComplete() {
    setShowOnboarding(false)
  }

  function handleOnboardingDismiss() {
    setShowOnboarding(false)
  }

  const handleNavigation = (event: any) => {
    if (event.detail && event.detail.href) {
      event.preventDefault()
      navigate(event.detail.href)
    }
  }

  return (
    <>
      <OnboardingWizard
        visible={showOnboarding && onboardingChecked}
        userId={USER_ID}
        trainingModules={trainingModules}
        onDismiss={handleOnboardingDismiss}
        onComplete={handleOnboardingComplete}
      />
      <div id="top-nav">
        <TopNavigation
          identity={{
            href: '/',
            title: 'qualify',
            logo: {
              src: '/vite.svg',
              alt: 'qualify'
            }
          }}
          utilities={[]}
          i18nStrings={{
            overflowMenuTitleText: 'More',
            overflowMenuTriggerText: 'More'
          }}
        />
      </div>
      <AppLayout
        navigationOpen={navigationOpen}
        onNavigationChange={(event) => setNavigationOpen(event.detail.open)}
        toolsHide={true}
        navigation={
          <>
            <SideNavigation
              items={navigationItems}
              activeHref={location.pathname}
              onFollow={handleNavigation}
            />
            <AgentStatus />
          </>
        }
        content={
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path="/dashboard" element={<Dashboard />} />
            <Route path="/s3" element={<S3 />} />
            <Route path="/training" element={<Training />} />
            <Route path="/training/:moduleName" element={<TrainingModule />} />
          </Routes>
        }
      />
    </>
  )
}
