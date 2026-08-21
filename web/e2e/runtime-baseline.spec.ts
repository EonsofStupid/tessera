import { expect, test } from '@playwright/test'

test.describe('production container baseline', () => {
  test('serves readiness and OIDC discovery from the same runtime', async ({ request }) => {
    const ready = await request.get('/debug/ready')
    expect(ready.status()).toBe(200)
    expect(await ready.json()).toBe('ok')

    const discovery = await request.get('/.well-known/openid-configuration')
    expect(discovery.status()).toBe(200)
    expect(await discovery.json()).toMatchObject({
      authorization_endpoint: expect.stringContaining('/oauth/v2/authorize'),
      token_endpoint: expect.stringContaining('/oauth/v2/token'),
      jwks_uri: expect.stringContaining('/oauth/v2/keys'),
      code_challenge_methods_supported: ['S256'],
    })
  })

  test('serves the standalone Tessera shell from the production binary', async ({ page }) => {
    await page.goto('/ui/console/')

    await expect(page).toHaveTitle(/Tessera/)
    await expect(page.getByRole('heading', { name: 'Good evening, operator.' })).toBeVisible()
    await expect(page.getByRole('navigation', { name: 'Primary navigation' })).toBeVisible()
    await expect(page.getByText('Connect your Tessera session')).toBeVisible()

    const productText = (await page.locator('body').innerText()).toLowerCase()
    expect(productText).not.toContain('zitadel')
    expect(productText).not.toContain('authentik')
  })
})
