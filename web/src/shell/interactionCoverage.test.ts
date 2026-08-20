import { describe, expect, it } from 'vitest'
import source from './TesseraApp.tsx?raw'

describe('semantic interaction coverage', () => {
  it('gives every interactive JSX control a stable semantic id', () => {
    const controls = source.match(/<(?:button|input|Link)\b[^>]*>/g) ?? []
    expect(controls.length).toBeGreaterThan(0)
    for (const control of controls) {
      expect(control, control).toContain('data-control-id=')
    }
  })

  it('contains only Tessera product language', () => {
    expect(source.toLowerCase()).not.toContain('zitadel')
    expect(source.toLowerCase()).not.toContain('authentik')
  })
})
