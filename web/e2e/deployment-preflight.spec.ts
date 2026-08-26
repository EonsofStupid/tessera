import { expect, test } from '@playwright/test'

const accessToken = process.env.NOMEN_E2E_ACCESS_TOKEN
const expectedChecks = ['database', 'issuer', 'tls', 'asymmetric_signing', 'notification_delivery']

test.describe('deployment preflight', () => {
  test('keeps deployment facts behind typed authorization', async ({ request }) => {
    for (const path of ['/nomen/v1/deployment/preflight', '/nomen/v1/deployment/owner-enrollment']) {
      const response = await request.get(path)
      expect(response.status()).toBe(401)
      const body = await response.json()
      expect(body).toMatchObject({ error: { type: 'authentication_required', reason: 'authentication_required' } })
      expect(JSON.stringify(body)).not.toContain('checks')
      expect(JSON.stringify(body)).not.toContain('challenge_digest')
    }
  })

  test('renders a protected read-only deployment workbench', async ({ page }) => {
    await page.goto('/ui/console/deployment')

    await expect(page.getByRole('heading', { name: 'Know what is safe before changing anything.' })).toBeVisible()
    await expect(page.getByText('Sign in to inspect deployment facts')).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Protected preflight' })).toBeVisible()
    await expect(page.getByText('No deployment fact is inferred from process health or placeholder content.')).toBeVisible()
  })

  test('returns five canonical live checks to an authorized operator', async ({ request }) => {
    test.skip(!accessToken, 'NOMEN_E2E_ACCESS_TOKEN is required for protected preflight evidence')
    const response = await request.get('/nomen/v1/deployment/preflight', { headers: { Authorization: `Bearer ${accessToken}` } })
    expect(response.status()).toBe(200)
    const body = await response.json()
    expect(body.schema_version).toBe(1)
    expect(body.checks.map((check: { id: string }) => check.id)).toEqual(expectedChecks)
    expect(['ready', 'action_required', 'blocked']).toContain(body.status)
    for (const check of body.checks) {
      expect(['passed', 'warning', 'failed']).toContain(check.status)
      expect(check.summary).toBeTruthy()
    }
  })
})
