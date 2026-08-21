import { accessToken } from './auth'
import {
  isCapabilityDiscovery,
  isManagementErrorEnvelope,
  isOverview,
  isOperatorActionCatalog,
  type CapabilityDiscovery,
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
        message: 'Tessera could not read this management resource.',
        remedy: { kind: 'retry_later', label: 'Try again' },
        retry: 'same_request',
      },
    }
  }
}

export function getCapabilities(): Promise<ResourceState<CapabilityDiscovery>> {
  return get('/tessera/v1/capabilities', isCapabilityDiscovery)
}

export function getOverview(): Promise<ResourceState<Overview>> {
  return get('/tessera/v1/overview', isOverview)
}

export function getOperatorActions(): Promise<ResourceState<OperatorActionCatalog>> {
  return get('/tessera/v1/operator/actions', isOperatorActionCatalog)
}
