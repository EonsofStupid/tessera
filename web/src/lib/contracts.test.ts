import { describe, expect, it } from 'vitest'
import { isCapabilityDiscovery, isOverview } from './contracts'

describe('management contract guards', () => {
  const overview = {
    schema_version: 1,
    service_id: 'tessera',
    resource_revision: 'revision-1',
    observed_at: '2026-08-21T00:00:00Z',
    readiness: { status: 'ready', issuer: 'https://identity.example.test', signing_keys: 1, flows: 1, policy_revision: 'policy-1', reasons: [] },
    lenses: [],
    federation: { upstreams: [], clients: [] },
    activity: [],
  }

  it('accepts the standalone overview service identity', () => {
    expect(isOverview(overview)).toBe(true)
  })

  it('rejects an overview projected for another product identity', () => {
    expect(isOverview({ ...overview, service_id: 'host.tessera' })).toBe(false)
  })

  it('requires both component and capability fact arrays', () => {
    const discovery = { schema_version: 1, resource_revision: 'revision-1', observed_at: '2026-08-21T00:00:00Z', components: [], capabilities: [] }
    expect(isCapabilityDiscovery(discovery)).toBe(true)
    expect(isCapabilityDiscovery({ ...discovery, capabilities: undefined })).toBe(false)
  })

  it('rejects a deeply malformed capability instead of trusting array presence', () => {
    const discovery = {
      schema_version: 1,
      resource_revision: 'revision-1',
      observed_at: '2026-08-21T00:00:00Z',
      components: [],
      capabilities: [{ id: 'downstream_oidc', status: 'operational', exposure: 'enabled' }],
    }
    expect(isCapabilityDiscovery(discovery)).toBe(false)
  })
})
