import Container from '@cloudscape-design/components/container'
import Header from '@cloudscape-design/components/header'
import SpaceBetween from '@cloudscape-design/components/space-between'
import Box from '@cloudscape-design/components/box'

export default function Home() {
  return (
    <SpaceBetween size="l">
      <Header variant="h1">
        Welcome to Ark
      </Header>

      <Container>
        <SpaceBetween size="l">
          <Box variant="p">
            Ark is an AWS Research Kit designed for academic institutions, providing integrated AWS training and security tooling.
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
