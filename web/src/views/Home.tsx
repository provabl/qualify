import Container from '@cloudscape-design/components/container'
import Header from '@cloudscape-design/components/header'
import SpaceBetween from '@cloudscape-design/components/space-between'
import Box from '@cloudscape-design/components/box'

export default function Home() {
  return (
    <SpaceBetween size="l">
      <Header variant="h1">
        Welcome to qualify
      </Header>

      <Container>
        <SpaceBetween size="l">
          <Box variant="p">
            qualify provides compliance training and per-researcher access gating for AWS Secure Research Environments.
          </Box>

          <Box variant="h2">Key Features</Box>
          <ul>
            <li>Training-gated AWS operations</li>
            <li>Comprehensive audit logging</li>
            <li>S3 bucket management</li>
            <li>Interactive training modules</li>
          </ul>

          <Box variant="p">
            Navigate using the sidebar to explore different features.
          </Box>
        </SpaceBetween>
      </Container>
    </SpaceBetween>
  )
}
