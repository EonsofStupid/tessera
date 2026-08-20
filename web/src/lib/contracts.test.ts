import { describe, expect, it } from 'vitest'
import { isCapabilityDiscovery, isOverview } from './contracts'

describe('management contract guards', () => {
  it('accepts the standalone overview service identity', () => {
    expect(isOverview({ schema_version: 1, service_id: 'tessera', readiness: {} })).toBe(true)
  })

  it('rejects an overview projected for another product identity', () => {
    expect(isOverview({ schema_version: 1, service_id: 'host.tessera', readiness: {} })).toBe(false)
  })

  it('requires both component and capability fact arrays', () => {
    expect(isCapabilityDiscovery({ schema_version: 1, components: [], capabilities: [] })).toBe(true)
    expect(isCapabilityDiscovery({ schema_version: 1, components: [] })).toBe(false)
  })
})
