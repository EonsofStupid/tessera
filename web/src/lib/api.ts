import { accessToken } from './auth'
import {
  isCapabilityDiscovery,
  isDeploymentPreflight,
  isOwnerEnrollment,
  isOwnerEnrollmentBegin,
  isOwnerEnrollmentComplete,
  isManagementErrorEnvelope,
  isOverview,
  isOperatorActionCatalog,
  type CapabilityDiscovery,
  type DeploymentPreflight,
  type OwnerEnrollment,
  type OwnerEnrollmentBegin,
  type OwnerEnrollmentComplete,
  type ManagementErrorEnvelope,
  type Overview,
  type OperatorActionCatalog,
  type ResourceState,
} from './contracts'

async function get<T>(path: string, validate: (value: unknown) => value is T): Promise<ResourceState<T>> {
  const token = accessToken()
  if (!token) return { status: 'authentication_required' }

  try {
    const response = await fetch(path, {
      headers: { Authorization: `Bearer ${token}` },
      credentials: 'same-origin',
      signal: AbortSignal.timeout(8_000),
    })
    const body: unknown = await response.json()
    if (response.ok && validate(body)) return { status: 'ready', data: body }
    const envelope: ManagementErrorEnvelope | undefined = isManagementErrorEnvelope(body) ? body : undefined
    if (response.status === 401) return { status: 'authentication_required', error: envelope?.error }
    if (response.status === 403) return { status: 'forbidden', error: envelope?.error }
    return { status: 'unavailable', error: envelope?.error }
  } catch {
    return {
      status: 'unavailable',
      error: {
        type: 'service_unavailable',
        reason: 'management_api_unreachable',
        message: 'Nomen could not read this management resource.',
        remedy: { kind: 'retry_later', label: 'Try again' },
        retry: 'same_request',
      },
    }
  }
}

export function getCapabilities(): Promise<ResourceState<CapabilityDiscovery>> {
  return get('/nomen/v1/capabilities', isCapabilityDiscovery)
}

export function getOverview(): Promise<ResourceState<Overview>> {
  return get('/nomen/v1/overview', isOverview)
}

export function getDeploymentPreflight(): Promise<ResourceState<DeploymentPreflight>> {
  return get('/nomen/v1/deployment/preflight', isDeploymentPreflight)
}

export function getOwnerEnrollment(): Promise<ResourceState<OwnerEnrollment>> {
  return get('/nomen/v1/deployment/owner-enrollment', isOwnerEnrollment)
}

export async function getOwnerEnrollmentWithBootstrap(authority: string): Promise<OwnerEnrollment> {
	const response = await fetch('/nomen/v1/deployment/owner-enrollment', {
		headers: { Authorization: `Bootstrap ${authority}` },
		credentials: 'same-origin',
		signal: AbortSignal.timeout(8_000),
	})
	const responseText = await response.text()
	let value: unknown
	try { value = JSON.parse(responseText) } catch { throw new Error('Nomen could not read owner-enrollment state at this deployment origin.') }
	if (response.ok && isOwnerEnrollment(value)) return value
	if (isManagementErrorEnvelope(value)) throw new Error(value.error.message)
	throw new Error('Nomen returned an invalid owner-enrollment response.')
}

async function bootstrapPost<T>(path: string, authority: string, body: unknown, validate: (value: unknown) => value is T, extraHeaders: Record<string, string> = {}): Promise<T> {
  const response = await fetch(path, {
    method: 'POST',
    headers: {
      Authorization: `Bootstrap ${authority}`,
      'Content-Type': 'application/json',
      ...extraHeaders,
    },
    credentials: 'same-origin',
    body: JSON.stringify(body),
    signal: AbortSignal.timeout(30_000),
  })
  const responseText = await response.text()
  let value: unknown
  try {
    value = JSON.parse(responseText)
  } catch {
    throw new Error('Nomen could not process the owner-enrollment request at this deployment origin.')
  }
  if (response.ok && validate(value)) return value
  if (isManagementErrorEnvelope(value)) throw new Error(value.error.message)
  throw new Error('Nomen returned an invalid owner-enrollment response.')
}

export function beginOwnerEnrollment(authority: string, idempotencyKey: string, request: { owner_id: string; username: string; display_name: string }): Promise<OwnerEnrollmentBegin> {
  return bootstrapPost('/nomen/v1/deployment/owner-enrollment:begin', authority, request, isOwnerEnrollmentBegin, { 'Idempotency-Key': idempotencyKey })
}

export function completeOwnerEnrollment(authority: string, ceremonyID: string, credential: unknown): Promise<OwnerEnrollmentComplete> {
  return bootstrapPost('/nomen/v1/deployment/owner-enrollment:complete', authority, { ceremony_id: ceremonyID, credential }, isOwnerEnrollmentComplete)
}

export function confirmOwnerRecovery(authority: string, ceremonyID: string, recoveryArtifact: string): Promise<OwnerEnrollment> {
  return bootstrapPost('/nomen/v1/deployment/owner-enrollment/recovery:confirm', authority, { ceremony_id: ceremonyID, recovery_artifact: recoveryArtifact }, isOwnerEnrollment)
}

export function getOperatorActions(): Promise<ResourceState<OperatorActionCatalog>> {
  return get('/nomen/v1/operator/actions', isOperatorActionCatalog)
}
