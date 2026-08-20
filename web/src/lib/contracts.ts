export type CapabilityStatus = 'unsupported' | 'preview' | 'operational' | 'degraded'
export type UIExposure = 'hidden' | 'disabled' | 'enabled'
export type CompatibilityState = 'compatible' | 'incompatible' | 'not_present' | 'unknown'

export interface ManagementErrorEnvelope {
  error: {
    type: string
    reason: string
    message: string
    remedy: { kind: string; label: string }
    retry: string
    diagnostic_ref?: string
    missing_entitlement?: string
    required_permission?: string
  }
}

export interface CapabilityDiscovery {
  schema_version: 1
  resource_revision: string
  bundle_manifest_digest?: string
  observed_at: string
  components: Array<{
    role: string
    version?: string
    api_major?: number
    state: CompatibilityState
    reason?: string
    observed_at: string
  }>
  capabilities: Array<{
    id: string
    status: CapabilityStatus
    exposure: UIExposure
    reason?: string
    required_components: string[]
    operation_kinds: string[]
    proof?: {
      conformance_id: string
      bundle_manifest_digest: string
      result: 'passed' | 'failed'
      verified_at: string
      evidence_digest: string
    }
  }>
}

export interface Overview {
  schema_version: 1
  service_id: 'tessera'
  resource_revision: string
  observed_at: string
  readiness: {
    status: 'ready' | 'degraded'
    issuer: string
    signing_keys: number
    flows: number
    policy_revision: string
    reasons: string[]
  }
  lenses: Array<{
    id: 'infrastructure' | 'ai' | 'customers'
    label: string
    value: number
    unit: string
    detail: string
    status: 'ready' | 'degraded'
  }>
  federation: {
    upstreams: unknown[]
    clients: unknown[]
  }
  activity: unknown[]
}

export interface OperatorActionCatalog {
  schema_version: 1
  resource_revision: string
  observed_at: string
  actions: Array<{
    id: string
    title: string
    consequence: string
    stage: 'read' | 'plan' | 'apply' | 'verify' | 'cancel'
    method: string
    href: string
    intent_schema: Record<string, unknown>
    required_permissions: string[]
    required_assurance?: string
    capability_id: string
    exposure: UIExposure
    reason?: string
    reversible: boolean
    seed_suggestions: Array<{ id: string; label: string; value?: string; source: string; explanation: string }>
  }>
}

export interface Environment {
  api?: string
  issuer?: string
  clientid?: string
}

export interface ResourceState<T> {
  status: 'loading' | 'ready' | 'authentication_required' | 'forbidden' | 'unavailable'
  data?: T
  error?: ManagementErrorEnvelope['error']
}

export function isCapabilityDiscovery(value: unknown): value is CapabilityDiscovery {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<CapabilityDiscovery>
  return candidate.schema_version === 1 && Array.isArray(candidate.components) && Array.isArray(candidate.capabilities)
}

export function isOverview(value: unknown): value is Overview {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<Overview>
  return candidate.schema_version === 1 && candidate.service_id === 'tessera' && Boolean(candidate.readiness)
}

export function isOperatorActionCatalog(value: unknown): value is OperatorActionCatalog {
  if (!value || typeof value !== 'object') return false
  const candidate = value as Partial<OperatorActionCatalog>
  return candidate.schema_version === 1 && Array.isArray(candidate.actions)
}
