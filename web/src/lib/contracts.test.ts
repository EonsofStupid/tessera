import { describe, expect, it } from 'vitest'
import { isCapabilityDiscovery, isDeploymentPreflight, isOverview, isOwnerEnrollment } from './contracts'

describe('management contract guards', () => {
  const overview = {
    schema_version: 1,
    service_id: 'nomen',
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
    expect(isOverview({ ...overview, service_id: 'host.nomen' })).toBe(false)
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

  it('accepts only the versioned deployment preflight shape', () => {
    const preflight = {
      schema_version: 1,
      resource_revision: 'sha256:revision',
      observed_at: '2026-08-22T00:00:00Z',
      status: 'action_required',
      issuer: 'https://identity.example.test',
      checks: [{ id: 'notification_delivery', status: 'warning', required: false, summary: 'Email is not configured.', reason: 'notification_provider_missing', remediation: 'Configure email.' }],
    }
    expect(isDeploymentPreflight(preflight)).toBe(true)
    expect(isDeploymentPreflight({ ...preflight, checks: [{ ...preflight.checks[0], status: 'probably' }] })).toBe(false)
    expect(isDeploymentPreflight({ ...preflight, schema_version: 2 })).toBe(false)
  })

  it('rejects invented owner-enrollment completion flags', () => {
    const enrollment = { schema_version: 1, resource_revision: 'sha256:revision', observed_at: '2026-08-22T00:00:00Z', state: 'pending', passkey_enrolled: false, recovery_confirmed: false, revision: 0 }
    expect(isOwnerEnrollment(enrollment)).toBe(true)
    expect(isOwnerEnrollment({ ...enrollment, recovery_confirmed: 'yes' })).toBe(false)
    expect(isOwnerEnrollment({ ...enrollment, state: 'password_complete' })).toBe(false)
  })
})
