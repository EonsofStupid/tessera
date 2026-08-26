import { expect, test } from '@playwright/test'

const runtimeOrigin = (process.env.NOMEN_E2E_BASE_URL ?? 'https://127.0.0.1:8089').replace(/\/$/, '')

test.describe('production container baseline', () => {
  test('serves readiness and OIDC discovery from the same runtime', async ({ request }) => {
    const ready = await request.get('/debug/ready')
    expect(ready.status()).toBe(200)
    expect(await ready.json()).toBe('ok')

    const discovery = await request.get('/.well-known/openid-configuration')
    expect(discovery.status()).toBe(200)
    expect(await discovery.json()).toMatchObject({
      issuer: runtimeOrigin,
      authorization_endpoint: expect.stringContaining('/oauth/v2/authorize'),
      token_endpoint: expect.stringContaining('/oauth/v2/token'),
      jwks_uri: expect.stringContaining('/oauth/v2/keys'),
      code_challenge_methods_supported: ['S256'],
    })
  })

  test('opens a public Nomen landing page from the runtime root', async ({ page }) => {
    await page.goto('/')

    await expect(page).toHaveTitle(/Nomen/)
    await expect(page).toHaveURL(/\/ui\/console\/$/)
    await expect(page.getByRole('heading', { name: /Your identities.*Your keys.*Your boundary/ })).toBeVisible()
    await expect(page.getByRole('navigation', { name: 'Public navigation' })).toBeVisible()
    await expect(page.getByText('PostgreSQL authority')).toBeVisible()
    await expect(page.getByText('Nomen Vault')).toBeVisible()
    await expect(page.getByText('Nomen Mesh')).toBeVisible()
    await expect(page.getByText(/Redis is only on paid Nomen\.sh hosting|Public edition/)).toBeVisible()
    await expect(page.getByText(/Public edition|Hosted demo|Enterprise/)).toBeVisible()

    const productText = (await page.locator('body').innerText()).toLowerCase()
    expect(productText).not.toContain('zitadel')
    expect(productText).not.toContain('authentik')
  })

  test('keeps tenant facts behind the operator workspace', async ({ page }) => {
    await page.goto('/ui/console/overview')

    await expect(page.getByRole('heading', { name: 'Your identity system, at a glance.' })).toBeVisible()
    await expect(page.getByRole('navigation', { name: 'Primary navigation' })).toBeVisible()
    await expect(page.getByText('Connect your Nomen session')).toBeVisible()
  })

  test('supports keyboard search and product-area navigation', async ({ page }) => {
    await page.goto('/ui/console/overview')

    const search = page.getByRole('textbox', { name: 'Search Nomen' })
		await expect(search).toBeVisible()
		await page.keyboard.press('Control+K')
    await expect(search).toBeFocused()
    await search.fill('feder')
    await page.getByRole('option', { name: /Federation/ }).click()

    await expect(page).toHaveURL(/\/ui\/console\/federation$/)
    await expect(page.getByRole('heading', { name: 'Bring every trusted identity source together' })).toBeVisible()
  })

  test('uses an accessible navigation drawer on a phone viewport', async ({ page }) => {
    await page.setViewportSize({ width: 390, height: 844 })
    await page.goto('/ui/console/overview')

    const openNavigation = page.getByRole('button', { name: 'Open navigation' })
    await expect(openNavigation).toBeVisible()
    await openNavigation.click()
    await expect(openNavigation).toHaveAttribute('aria-expanded', 'true')
    await expect(page.getByRole('link', { name: 'Applications' })).toBeVisible()
    await page.getByRole('link', { name: 'Applications' }).click()
    await expect(page).toHaveURL(/\/ui\/console\/applications$/)
    await expect(page.getByRole('heading', { name: 'Connect applications without guesswork' })).toBeVisible()

    const hasHorizontalOverflow = await page.evaluate(() => document.documentElement.scrollWidth > window.innerWidth)
    expect(hasHorizontalOverflow).toBe(false)
  })

  test('starts Authorization Code with PKCE from the public page', async ({ page }) => {
    let authorizationURL = ''
    page.on('request', request => {
      if (new URL(request.url()).pathname === '/oauth/v2/authorize') authorizationURL = request.url()
    })
    await page.goto('/ui/console/')

    const signIn = page.getByRole('button', { name: 'Sign in to your deployment' })
    await expect(signIn).toBeEnabled()
    await signIn.click()
    await expect.poll(() => authorizationURL).not.toBe('')
    await expect(page).toHaveURL(/\/ui\/(?:v2\/)?login\/login/)

    const parsed = new URL(authorizationURL)
    expect(parsed.pathname).toBe('/oauth/v2/authorize')
    expect(parsed.searchParams.get('response_type')).toBe('code')
    expect(parsed.searchParams.get('code_challenge_method')).toBe('S256')
    expect(parsed.searchParams.get('code_challenge')).toBeTruthy()
    expect(parsed.searchParams.get('redirect_uri')).toMatch(/\/ui\/console\/auth\/callback$/)
  })
})
