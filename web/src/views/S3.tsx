import { useState, useMemo, useEffect } from 'react'
import Container from '@cloudscape-design/components/container'
import Header from '@cloudscape-design/components/header'
import SpaceBetween from '@cloudscape-design/components/space-between'
import Box from '@cloudscape-design/components/box'
import Button from '@cloudscape-design/components/button'
import Form from '@cloudscape-design/components/form'
import FormField from '@cloudscape-design/components/form-field'
import Input from '@cloudscape-design/components/input'
import Select, { SelectProps } from '@cloudscape-design/components/select'
import Toggle from '@cloudscape-design/components/toggle'
import Flashbar from '@cloudscape-design/components/flashbar'
import Alert from '@cloudscape-design/components/alert'
import { agentService } from '@/services/agent'
import type { CreateBucketRequest, S3Bucket, PolicyDecision, TrainingModule } from '@/types/api'

// Region options
const regionOptions = [
  { label: 'US East (N. Virginia)', value: 'us-east-1' },
  { label: 'US West (Oregon)', value: 'us-west-2' },
  { label: 'EU (Ireland)', value: 'eu-west-1' },
  { label: 'Asia Pacific (Singapore)', value: 'ap-southeast-1' }
]

// Encryption options
const encryptionOptions = [
  { label: 'AES256', value: 'AES256' },
  { label: 'AWS KMS', value: 'aws:kms' }
]

// Validation function
function validateBucketName(name: string): string | null {
  if (name.length < 3 || name.length > 63) {
    return 'Bucket name must be between 3 and 63 characters'
  }

  if (name[0] === '-' || name[name.length - 1] === '-') {
    return 'Bucket name cannot start or end with a hyphen'
  }

  if (name[0] === '.' || name[name.length - 1] === '.') {
    return 'Bucket name cannot start or end with a period'
  }

  for (let i = 0; i < name.length; i++) {
    const c = name[i]
    if (!((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c === '-' || c === '.')) {
      return 'Bucket name can only contain lowercase letters, numbers, hyphens, and periods'
    }

    if (c === '.' && i > 0 && name[i - 1] === '.') {
      return 'Bucket name cannot contain consecutive periods'
    }
  }

  // Check if it looks like an IP address
  const parts = name.split('.')
  if (parts.length === 4) {
    let allNumeric = true
    for (const part of parts) {
      if (part.length === 0 || part.length > 3) {
        allNumeric = false
        break
      }
      for (const c of part) {
        if (c < '0' || c > '9') {
          allNumeric = false
          break
        }
      }
    }
    if (allNumeric) {
      return 'Bucket name cannot be formatted as an IP address'
    }
  }

  return null
}

export default function S3() {
  // Form state
  const [bucketName, setBucketName] = useState('')
  const [region, setRegion] = useState<SelectProps.Option>({ label: 'US East (N. Virginia)', value: 'us-east-1' })
  const [encryptionType, setEncryptionType] = useState<SelectProps.Option>({ label: 'AES256', value: 'AES256' })
  const [kmsKeyId, setKmsKeyId] = useState('')
  const [versioningEnabled, setVersioningEnabled] = useState(false)
  const [selectedProfile, setSelectedProfile] = useState<SelectProps.Option>({ label: 'default', value: 'default' })

  // UI state
  const [isCreating, setIsCreating] = useState(false)
  const [showForm, setShowForm] = useState(false)
  const [profiles, setProfiles] = useState<Array<{ label: string; value: string }>>([])
  const [flashbarItems, setFlashbarItems] = useState<any[]>([])
  const [trainingRequired, setTrainingRequired] = useState<TrainingModule[] | null>(null)

  // Computed properties
  const bucketNameError = useMemo(() => {
    if (!bucketName) return null
    return validateBucketName(bucketName)
  }, [bucketName])

  const isFormValid = useMemo(() => {
    if (!bucketName) return false
    if (bucketNameError) return false
    if (encryptionType.value === 'aws:kms' && !kmsKeyId) return false
    return true
  }, [bucketName, bucketNameError, encryptionType.value, kmsKeyId])

  // Load credential profiles
  useEffect(() => {
    loadProfiles()
  }, [])

  async function loadProfiles() {
    try {
      const profileList = await agentService.listCredentials()
      const mappedProfiles = profileList.map(p => ({ label: p, value: p }))
      setProfiles(mappedProfiles.length > 0 ? mappedProfiles : [{ label: 'default', value: 'default' }])
    } catch (error) {
      console.error('Failed to load profiles:', error)
      setProfiles([{ label: 'default', value: 'default' }])
    }
  }

  // Handle bucket creation
  async function handleCreateBucket() {
    if (!isFormValid) return

    setIsCreating(true)
    setTrainingRequired(null)
    setFlashbarItems([])

    try {
      const request: CreateBucketRequest = {
        bucket_name: bucketName,
        region: region.value ?? 'us-east-1',
        encryption: {
          type: (encryptionType.value as 'AES256' | 'aws:kms') ?? 'AES256',
          kms_key_id: encryptionType.value === 'aws:kms' ? kmsKeyId : undefined
        },
        versioning_enabled: versioningEnabled,
        profile: selectedProfile.value ?? 'default'
      }

      const response = await agentService.createBucket(request)

      // Check if response is PolicyDecision (training block)
      if ('action' in response && response.action === 'block') {
        handleTrainingBlock(response as PolicyDecision)
      } else {
        handleSuccess(response as S3Bucket)
      }
    } catch (error) {
      handleError(error as Error)
    } finally {
      setIsCreating(false)
    }
  }

  // Handle successful bucket creation
  function handleSuccess(bucket: S3Bucket) {
    const createdDate = bucket.created_at ? new Date(bucket.created_at).toLocaleString() : ''

    setFlashbarItems([{
      type: 'success',
      dismissible: true,
      dismissLabel: 'Dismiss',
      content: `Bucket Created Successfully: ${bucket.bucket_name} in ${bucket.region}` +
               (bucket.location ? ` (${bucket.location})` : '') +
               (createdDate ? ` at ${createdDate}` : ''),
      onDismiss: () => setFlashbarItems([])
    }])

    // Reset form
    setShowForm(false)
    setBucketName('')
    setKmsKeyId('')
    setVersioningEnabled(false)
    setEncryptionType({ label: 'AES256', value: 'AES256' })
  }

  // Handle training gate block
  function handleTrainingBlock(decision: PolicyDecision) {
    setTrainingRequired(decision.required_modules || [])
  }

  // Handle errors
  function handleError(error: Error) {
    setFlashbarItems([{
      type: 'error',
      dismissible: true,
      dismissLabel: 'Dismiss',
      content: `Bucket Creation Failed: ${error.message}`,
      onDismiss: () => setFlashbarItems([])
    }])
  }

  // Toggle form visibility
  function toggleForm() {
    setShowForm(prev => !prev)
    setTrainingRequired(null)
  }

  return (
    <SpaceBetween size="l">
      <Flashbar items={flashbarItems} />

      <Header
        variant="h1"
        actions={
          <Button variant="primary" onClick={toggleForm}>
            {showForm ? 'Cancel' : 'Create Bucket'}
          </Button>
        }
      >
        S3 Buckets
      </Header>

      {/* Training Required Alert */}
      {trainingRequired && trainingRequired.length > 0 && (
        <Alert
          type="warning"
          dismissible
          onDismiss={() => setTrainingRequired(null)}
        >
          <SpaceBetween size="m">
            <Box variant="strong">Training Required</Box>
            <Box>Complete the following modules before creating S3 buckets:</Box>
            <SpaceBetween size="s">
              {trainingRequired.map((module, index) => (
                <div key={module.id}>
                  <Box>{index + 1}. {module.title} ({module.estimated_minutes} minutes)</Box>
                  <Box fontSize="body-s" color="text-body-secondary">
                    Start training: http://localhost:8080/training/{module.name}
                  </Box>
                </div>
              ))}
            </SpaceBetween>
            <Box>After completing training, try again.</Box>
          </SpaceBetween>
        </Alert>
      )}

      {/* Bucket Creation Form */}
      {showForm && (
        <Container>
          <Form
            actions={
              <SpaceBetween direction="horizontal" size="xs">
                <Button variant="link" onClick={toggleForm}>Cancel</Button>
                <Button
                  variant="primary"
                  disabled={!isFormValid || isCreating}
                  loading={isCreating}
                  onClick={handleCreateBucket}
                >
                  Create Bucket
                </Button>
              </SpaceBetween>
            }
          >
            <SpaceBetween size="l">
              <FormField
                label="Bucket name"
                errorText={bucketNameError || undefined}
                description="Must be globally unique. 3-63 characters, lowercase letters, numbers, hyphens, and periods only."
              >
                <Input
                  value={bucketName}
                  onChange={(event) => setBucketName(event.detail.value)}
                  placeholder="my-research-bucket"
                  disabled={isCreating}
                />
              </FormField>

              <FormField
                label="AWS Region"
                description="The AWS region where the bucket will be created"
              >
                <Select
                  selectedOption={region}
                  onChange={(event) => setRegion(event.detail.selectedOption)}
                  options={regionOptions}
                  disabled={isCreating}
                />
              </FormField>

              <FormField
                label="Encryption"
                description="Server-side encryption for objects stored in the bucket"
              >
                <Select
                  selectedOption={encryptionType}
                  onChange={(event) => setEncryptionType(event.detail.selectedOption)}
                  options={encryptionOptions}
                  disabled={isCreating}
                />
              </FormField>

              {encryptionType.value === 'aws:kms' && (
                <FormField
                  label="KMS Key ID"
                  description="The AWS KMS key ID for encryption"
                >
                  <Input
                    value={kmsKeyId}
                    onChange={(event) => setKmsKeyId(event.detail.value)}
                    placeholder="arn:aws:kms:us-east-1:123456789012:key/..."
                    disabled={isCreating}
                  />
                </FormField>
              )}

              <FormField
                label="Versioning"
                description="Enable versioning to keep multiple versions of objects"
              >
                <Toggle
                  checked={versioningEnabled}
                  onChange={(event) => setVersioningEnabled(event.detail.checked)}
                  disabled={isCreating}
                >
                  {versioningEnabled ? 'Enabled' : 'Disabled'}
                </Toggle>
              </FormField>

              <FormField
                label="Credential Profile"
                description="AWS credential profile to use for this operation"
              >
                <Select
                  selectedOption={selectedProfile}
                  onChange={(event) => setSelectedProfile(event.detail.selectedOption)}
                  options={profiles}
                  disabled={isCreating}
                />
              </FormField>
            </SpaceBetween>
          </Form>
        </Container>
      )}

      {/* Empty State */}
      {!showForm && (
        <Container>
          <SpaceBetween size="m">
            <Box variant="h2">S3 Bucket Management</Box>
            <Box variant="p">
              Create and manage S3 buckets with integrated training gates and audit logging.
            </Box>
            <Box variant="p">
              <strong>Note:</strong> S3 bucket creation requires completion of the S3 Basics training module.
            </Box>
            <Box variant="p" color="text-status-info">
              Click "Create Bucket" above to get started.
            </Box>
          </SpaceBetween>
        </Container>
      )}
    </SpaceBetween>
  )
}
