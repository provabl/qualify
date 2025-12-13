import { useEffect } from 'react'
import { useAgent } from '@/contexts/AgentContext'
import StatusIndicator from '@cloudscape-design/components/status-indicator'
import Box from '@cloudscape-design/components/box'

function getStatusType(status: string) {
  switch (status) {
    case 'connected':
      return 'success'
    case 'checking':
      return 'loading'
    case 'error':
    case 'disconnected':
      return 'error'
    default:
      return 'stopped'
  }
}

function getStatusText(status: string, error: string | null) {
  switch (status) {
    case 'connected':
      return 'Agent Connected'
    case 'checking':
      return 'Checking Agent...'
    case 'error':
      return `Agent Error: ${error}`
    case 'disconnected':
      return 'Agent Disconnected'
    default:
      return 'Unknown Status'
  }
}

export default function AgentStatus() {
  const agent = useAgent()

  useEffect(() => {
    // Start periodic agent connectivity checks (every 30 seconds)
    agent.startPeriodicCheck(30000)
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <Box padding={{ vertical: 'xs', horizontal: 's' }}>
      <StatusIndicator type={getStatusType(agent.status)}>
        {getStatusText(agent.status, agent.error)}
        {agent.version && agent.isConnected && (
          <span> (v{agent.version})</span>
        )}
      </StatusIndicator>
    </Box>
  )
}
