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
  service_id: "'tessera'",
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
})

export const tokenResponseSchema = type({ access_token: 'string' })

export type ManagementErrorEnvelope = typeof managementErrorEnvelopeSchema.infer
export type CapabilityDiscovery = typeof capabilityDiscoverySchema.infer
export type Overview = typeof overviewSchema.infer
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
