import { useState, useMemo, useEffect } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import Container from '@cloudscape-design/components/container'
import Header from '@cloudscape-design/components/header'
import SpaceBetween from '@cloudscape-design/components/space-between'
import Box from '@cloudscape-design/components/box'
import Button from '@cloudscape-design/components/button'
import ProgressBar from '@cloudscape-design/components/progress-bar'
import Alert from '@cloudscape-design/components/alert'
import { agentService } from '@/services/agent'
import type { TrainingModule, TrainingSection } from '@/types/api'
import Quiz from '@/components/training/Quiz'

// Hardcoded user ID for now (TODO: Replace with actual auth)
const USER_ID = '00000000-0000-0000-0000-000000000001'

export default function TrainingModuleView() {
  const { moduleName } = useParams<{ moduleName: string }>()
  const navigate = useNavigate()

  const [module, setModule] = useState<TrainingModule | null>(null)
  const [currentSectionIndex, setCurrentSectionIndex] = useState(0)
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isStarting, setIsStarting] = useState(false)
  const [isCompleting, setIsCompleting] = useState(false)
  const [hasStarted, setHasStarted] = useState(false)
  const [quizScores, setQuizScores] = useState<Record<number, { score: number; passed: boolean }>>({})

  const currentSection = useMemo<TrainingSection | null>(() => {
    if (!module?.content?.sections) return null
    return module.content.sections[currentSectionIndex]
  }, [module, currentSectionIndex])

  const progress = useMemo(() => {
    if (!module?.content?.sections) return 0
    const total = module.content.sections.length
    return Math.round(((currentSectionIndex + 1) / total) * 100)
  }, [module, currentSectionIndex])

  const isFirstSection = useMemo(() => currentSectionIndex === 0, [currentSectionIndex])

  const isLastSection = useMemo(() => {
    if (!module?.content?.sections) return true
    return currentSectionIndex === module.content.sections.length - 1
  }, [module, currentSectionIndex])

  const currentSectionQuizPassed = useMemo(() => {
    if (currentSection?.type !== 'quiz') return true
    return quizScores[currentSectionIndex]?.passed ?? false
  }, [currentSection, currentSectionIndex, quizScores])

  const averageQuizScore = useMemo(() => {
    const scores = Object.values(quizScores)
    if (scores.length === 0) return undefined
    const sum = scores.reduce((acc, { score }) => acc + score, 0)
    return Math.round(sum / scores.length)
  }, [quizScores])

  useEffect(() => {
    loadModule()
  }, [moduleName])

  async function loadModule() {
    if (!moduleName) {
      setError('No module specified')
      setIsLoading(false)
      return
    }

    setIsLoading(true)
    setError(null)

    try {
      const loadedModule = await agentService.getTrainingModule(moduleName)
      setModule(loadedModule)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load training module')
      console.error('Failed to load module:', err)
    } finally {
      setIsLoading(false)
    }
  }

  async function startModule() {
    if (!module) return

    setIsStarting(true)
    try {
      await agentService.startTrainingModule(module.id, USER_ID)
      setHasStarted(true)
    } catch (err) {
      console.error('Failed to start module:', err)
      setError(err instanceof Error ? err.message : 'Failed to start module')
    } finally {
      setIsStarting(false)
    }
  }

  async function completeModule() {
    if (!module) return

    setIsCompleting(true)
    try {
      await agentService.completeTrainingModule(module.id, USER_ID, averageQuizScore)
      navigate('/training')
    } catch (err) {
      console.error('Failed to complete module:', err)
      setError(err instanceof Error ? err.message : 'Failed to complete module')
    } finally {
      setIsCompleting(false)
    }
  }

  function handleQuizComplete(score: number, passed: boolean) {
    setQuizScores(prev => ({
      ...prev,
      [currentSectionIndex]: { score, passed },
    }))
  }

  function previousSection() {
    if (!isFirstSection) {
      setCurrentSectionIndex(prev => prev - 1)
    }
  }

  function nextSection() {
    if (!isLastSection) {
      setCurrentSectionIndex(prev => prev + 1)
    }
  }

  function backToList() {
    navigate('/training')
  }

  return (
    <SpaceBetween size="l">
      <Header
        variant="h1"
        actions={
          <Button onClick={backToList}>Back to Training List</Button>
        }
      >
        {module?.title || 'Loading...'}
      </Header>

      {error && (
        <Alert
          type="error"
          dismissible
          onDismiss={() => setError(null)}
        >
          {error}
        </Alert>
      )}

      {isLoading && (
        <Container>
          <Box variant="p">Loading training module...</Box>
        </Container>
      )}

      {!isLoading && module && !hasStarted && (
        <Container>
          <SpaceBetween size="m">
            <Box variant="h2">{module.title}</Box>
            <Box variant="p">{module.description}</Box>
            <Box fontSize="body-s">
              <strong>Estimated time:</strong> {module.estimated_minutes} minutes
            </Box>
            {module.category && (
              <Box fontSize="body-s">
                <strong>Category:</strong> {module.category}
              </Box>
            )}
            {module.difficulty && (
              <Box fontSize="body-s">
                <strong>Difficulty:</strong> {module.difficulty}
              </Box>
            )}
            <Button
              variant="primary"
              loading={isStarting}
              onClick={startModule}
            >
              Start Training
            </Button>
          </SpaceBetween>
        </Container>
      )}

      {!isLoading && module && hasStarted && currentSection && (
        <Container>
          <SpaceBetween size="m">
            <ProgressBar
              value={progress}
              label={`Section ${currentSectionIndex + 1} of ${module.content?.sections?.length || 0}`}
            />

            <Box variant="h2">{currentSection.title}</Box>

            {currentSection.type === 'text' && (
              <div style={{ whiteSpace: 'pre-wrap' }}>
                {currentSection.content}
              </div>
            )}

            {currentSection.type === 'quiz' && currentSection.questions && module && (
              <Quiz
                moduleId={module.id}
                userId={USER_ID}
                questions={currentSection.questions}
                onComplete={handleQuizComplete}
              />
            )}

            <SpaceBetween direction="horizontal" size="xs">
              <Button
                onClick={previousSection}
                disabled={isFirstSection}
              >
                Previous
              </Button>
              {!isLastSection ? (
                <Button
                  variant="primary"
                  onClick={nextSection}
                  disabled={!currentSectionQuizPassed}
                >
                  Next
                </Button>
              ) : (
                <Button
                  variant="primary"
                  loading={isCompleting}
                  onClick={completeModule}
                  disabled={!currentSectionQuizPassed}
                >
                  Complete Training
                </Button>
              )}
            </SpaceBetween>
            {!currentSectionQuizPassed && currentSection?.type === 'quiz' && (
              <Alert type="info">
                Pass the quiz to continue
              </Alert>
            )}
          </SpaceBetween>
        </Container>
      )}
    </SpaceBetween>
  )
}
