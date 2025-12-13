import { createContext, useContext, useReducer, useCallback, useEffect, useRef } from 'react'
import { agentService } from '@/services/agent'

// State types
interface AgentState {
  isConnected: boolean
  isChecking: boolean
  lastCheck: Date | null
  version: string | null
  error: string | null
}

// Action types
type AgentAction =
  | { type: 'CHECK_START' }
  | { type: 'CHECK_SUCCESS'; version?: string }
  | { type: 'CHECK_ERROR'; error: string }
  | { type: 'RESET' }

// Context types
interface AgentContextType extends AgentState {
  status: 'checking' | 'connected' | 'error' | 'disconnected'
  statusColor: 'green' | 'blue' | 'red' | 'grey'
  checkConnection: () => Promise<void>
  startPeriodicCheck: (intervalMs: number) => void
  reset: () => void
}

// Initial state
const initialState: AgentState = {
  isConnected: false,
  isChecking: false,
  lastCheck: null,
  version: null,
  error: null,
}

// Reducer
function agentReducer(state: AgentState, action: AgentAction): AgentState {
  switch (action.type) {
    case 'CHECK_START':
      return {
        ...state,
        isChecking: true,
        error: null,
      }
    case 'CHECK_SUCCESS':
      return {
        ...state,
        isConnected: true,
        isChecking: false,
        lastCheck: new Date(),
        version: action.version || state.version,
        error: null,
      }
    case 'CHECK_ERROR':
      return {
        ...state,
        isConnected: false,
        isChecking: false,
        lastCheck: new Date(),
        error: action.error,
      }
    case 'RESET':
      return initialState
    default:
      return state
  }
}

// Helper functions for computed values
function getStatus(state: AgentState): AgentContextType['status'] {
  if (state.isChecking) return 'checking'
  if (state.isConnected) return 'connected'
  if (state.error) return 'error'
  return 'disconnected'
}

function getStatusColor(status: AgentContextType['status']): AgentContextType['statusColor'] {
  switch (status) {
    case 'connected':
      return 'green'
    case 'checking':
      return 'blue'
    case 'error':
      return 'red'
    case 'disconnected':
      return 'grey'
  }
}

// Create context
const AgentContext = createContext<AgentContextType | undefined>(undefined)

// Provider component
export function AgentProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(agentReducer, initialState)
  const intervalRef = useRef<number | null>(null)

  const checkConnection = useCallback(async () => {
    dispatch({ type: 'CHECK_START' })
    try {
      const result = await agentService.healthCheck()
      dispatch({ type: 'CHECK_SUCCESS', version: result.version })
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Connection failed'
      dispatch({ type: 'CHECK_ERROR', error: errorMessage })
    }
  }, [])

  const startPeriodicCheck = useCallback((intervalMs: number) => {
    // Clear existing interval if any
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
    }

    // Check immediately
    checkConnection()

    // Start periodic checks
    intervalRef.current = setInterval(() => {
      checkConnection()
    }, intervalMs)
  }, [checkConnection])

  const reset = useCallback(() => {
    if (intervalRef.current) {
      clearInterval(intervalRef.current)
      intervalRef.current = null
    }
    dispatch({ type: 'RESET' })
  }, [])

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      if (intervalRef.current) {
        clearInterval(intervalRef.current)
      }
    }
  }, [])

  const status = getStatus(state)
  const statusColor = getStatusColor(status)

  const value: AgentContextType = {
    ...state,
    status,
    statusColor,
    checkConnection,
    startPeriodicCheck,
    reset,
  }

  return <AgentContext.Provider value={value}>{children}</AgentContext.Provider>
}

// Custom hook for using the agent context
export function useAgent() {
  const context = useContext(AgentContext)
  if (context === undefined) {
    throw new Error('useAgent must be used within an AgentProvider')
  }
  return context
}
