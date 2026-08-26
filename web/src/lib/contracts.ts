import { type } from 'arktype'

export type CapabilityStatus = 'unsupported' | 'preview' | 'operational' | 'degraded'
export type UIExposure = 'hidden' | 'disabled' | 'enabled'
export type CompatibilityState = 'compatible' | 'incompatible' | 'not_present' | 'unknown'

const compatibilityState = type("'compatible' | 'incompatible' | 'not_present' | 'unknown'")
const capabilityStatus = type("'unsupported' | 'preview' | 'operational' | 'degraded'")
const uiExposure = type("'hidden' | 'disabled' | 'enabled'")
const stringRecord = type({ '[string]': 'unknown' })

export const managementErrorEnvelopeSchema = type({
  error: {
    type: 'string',
    reason: 'string',
    message: 'string',
    remedy: { kind: 'string', label: 'string' },
    retry: 'string',
    'diagnostic_ref?': 'string',
    'missing_entitlement?': 'string',
    'required_permission?': 'string',
  },
})

const componentSchema = type({
  role: 'string',
  'version?': 'string',
  'api_major?': 'number.integer',
  state: compatibilityState,
  'reason?': 'string',
  observed_at: 'string',
})

const conformanceProofSchema = type({
  conformance_id: 'string',
  bundle_manifest_digest: 'string',
  result: "'passed' | 'failed'",
  verified_at: 'string',
  evidence_digest: 'string',
})

const capabilitySchema = type({
  id: 'string',
  status: capabilityStatus,
  exposure: uiExposure,
  'reason?': 'string',
  required_components: 'string[]',
  operation_kinds: 'string[]',
  'proof?': conformanceProofSchema,
})

export const capabilityDiscoverySchema = type({
  schema_version: '1',
  resource_revision: 'string',
  'bundle_manifest_digest?': 'string',
  observed_at: 'string',
  components: componentSchema.array(),
  capabilities: capabilitySchema.array(),
})

const readinessSchema = type({
  status: "'ready' | 'degraded'",
  issuer: 'string',
  signing_keys: 'number.integer >= 0',
  flows: 'number.integer >= 0',
  policy_revision: 'string',
  reasons: 'string[]',
})

const lensSchema = type({
  id: "'infrastructure' | 'ai' | 'customers'",
  label: 'string',
  value: 'number',
  unit: 'string',
  detail: 'string',
  status: "'ready' | 'degraded'",
})

export const overviewSchema = type({
  schema_version: '1',
  service_id: "'nomen'",
  resource_revision: 'string',
  observed_at: 'string',
  readiness: readinessSchema,
  lenses: lensSchema.array(),
  federation: {
    upstreams: 'unknown[]',
    clients: 'unknown[]',
  },
  activity: 'unknown[]',
})

const preflightCheckSchema = type({
  id: "'database' | 'issuer' | 'tls' | 'asymmetric_signing' | 'notification_delivery'",
  status: "'passed' | 'warning' | 'failed'",
  required: 'boolean',
  summary: 'string',
  'reason?': 'string',
  'remediation?': 'string',
  'diagnostic_ref?': 'string',
})

export const deploymentPreflightSchema = type({
  schema_version: '1',
  resource_revision: 'string',
  observed_at: 'string',
  status: "'ready' | 'action_required' | 'blocked'",
  issuer: 'string',
  checks: preflightCheckSchema.array(),
})

export const ownerEnrollmentSchema = type({
  schema_version: '1',
  resource_revision: 'string',
  observed_at: 'string',
  state: "'pending' | 'passkey_pending' | 'recovery_pending' | 'complete'",
  'ceremony_id?': 'string',
  'owner_id?': 'string',
  passkey_enrolled: 'boolean',
  recovery_confirmed: 'boolean',
  'expires_at?': 'string',
  revision: 'number.integer >= 0',
})

const webAuthnCredentialDescriptorSchema = type({
  type: "'public-key'",
  id: 'string',
  'transports?': 'string[]',
})

export const ownerEnrollmentBeginSchema = type({
  enrollment: ownerEnrollmentSchema,
  publicKey: {
    rp: { id: 'string', name: 'string' },
    user: { id: 'string', name: 'string', displayName: 'string' },
    challenge: 'string',
    pubKeyCredParams: type({ type: "'public-key'", alg: 'number.integer' }).array(),
    'timeout?': 'number.integer > 0',
    'excludeCredentials?': webAuthnCredentialDescriptorSchema.array(),
    'authenticatorSelection?': 'unknown',
    'attestation?': 'string',
    'extensions?': 'unknown',
  },
})

export const ownerEnrollmentCompleteSchema = type({
  enrollment: ownerEnrollmentSchema,
  recovery_artifact: 'string',
})

const seedSuggestionSchema = type({
  id: 'string',
  label: 'string',
  'value?': 'string',
  source: 'string',
  explanation: 'string',
})

const operatorActionSchema = type({
  id: 'string',
  title: 'string',
  consequence: 'string',
  stage: "'read' | 'plan' | 'apply' | 'verify' | 'cancel'",
  method: 'string',
  href: 'string',
  intent_schema: stringRecord,
  required_permissions: 'string[]',
  'required_assurance?': 'string',
  capability_id: 'string',
  exposure: uiExposure,
  'reason?': 'string',
  reversible: 'boolean',
  seed_suggestions: seedSuggestionSchema.array(),
})

export const operatorActionCatalogSchema = type({
  schema_version: '1',
  resource_revision: 'string',
  observed_at: 'string',
  actions: operatorActionSchema.array(),
})

export const environmentSchema = type({
  'api?': 'string',
  'issuer?': 'string',
  'clientid?': 'string',
  'edition?': "'public' | 'enterprise'",
  'demo_caps?': 'boolean',
  'version?': 'string',
})

export const tokenResponseSchema = type({ access_token: 'string' })

export type ManagementErrorEnvelope = typeof managementErrorEnvelopeSchema.infer
export type CapabilityDiscovery = typeof capabilityDiscoverySchema.infer
export type Overview = typeof overviewSchema.infer
export type DeploymentPreflight = typeof deploymentPreflightSchema.infer
export type OwnerEnrollment = typeof ownerEnrollmentSchema.infer
export type OwnerEnrollmentBegin = typeof ownerEnrollmentBeginSchema.infer
export type OwnerEnrollmentComplete = typeof ownerEnrollmentCompleteSchema.infer
export type OperatorActionCatalog = typeof operatorActionCatalogSchema.infer
export type Environment = typeof environmentSchema.infer

export interface ResourceState<T> {
  status: 'loading' | 'ready' | 'authentication_required' | 'forbidden' | 'unavailable'
  data?: T
  error?: ManagementErrorEnvelope['error']
}

function accepts<T>(schema: (value: unknown) => T | type.errors, value: unknown): value is T {
  return !(schema(value) instanceof type.errors)
}

export function isManagementErrorEnvelope(value: unknown): value is ManagementErrorEnvelope {
  return accepts(managementErrorEnvelopeSchema, value)
}

export function isCapabilityDiscovery(value: unknown): value is CapabilityDiscovery {
  return accepts(capabilityDiscoverySchema, value)
}

export function isOverview(value: unknown): value is Overview {
  return accepts(overviewSchema, value)
}

export function isDeploymentPreflight(value: unknown): value is DeploymentPreflight {
  return accepts(deploymentPreflightSchema, value)
}

export function isOwnerEnrollment(value: unknown): value is OwnerEnrollment {
  return accepts(ownerEnrollmentSchema, value)
}

export function isOwnerEnrollmentBegin(value: unknown): value is OwnerEnrollmentBegin {
  return accepts(ownerEnrollmentBeginSchema, value)
}

export function isOwnerEnrollmentComplete(value: unknown): value is OwnerEnrollmentComplete {
  return accepts(ownerEnrollmentCompleteSchema, value)
}

export function isOperatorActionCatalog(value: unknown): value is OperatorActionCatalog {
  return accepts(operatorActionCatalogSchema, value)
}

export function parseEnvironment(value: unknown): Environment {
  const result = environmentSchema(value)
  if (result instanceof type.errors) throw new Error(`environment_invalid: ${result.summary}`)
  return result
}

export function parseTokenResponse(value: unknown): typeof tokenResponseSchema.infer {
  const result = tokenResponseSchema(value)
  if (result instanceof type.errors) throw new Error(`token_response_invalid: ${result.summary}`)
  return result
}
