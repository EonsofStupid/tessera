import { readFile } from 'node:fs/promises'
import { expect, test } from '@playwright/test'

const bootstrapAuthority = process.env.NOMEN_E2E_BOOTSTRAP_AUTHORITY

test.describe('first-owner WebAuthn ceremony', () => {
  test('enrolls a verified passkey, confirms recovery, and consumes bootstrap', async ({ page, context, request }) => {
    test.skip(!bootstrapAuthority, 'NOMEN_E2E_BOOTSTRAP_AUTHORITY is required for the one-time bootstrap journey')

    const cdp = await context.newCDPSession(page)
    await cdp.send('WebAuthn.enable')
    await cdp.send('WebAuthn.addVirtualAuthenticator', {
      options: {
        protocol: 'ctap2',
        transport: 'internal',
        hasResidentKey: true,
        hasUserVerification: true,
        isUserVerified: true,
      },
    })

    await page.goto('/ui/console/deployment')
    await page.locator('[data-control-id="control.bootstrap_authority"]').fill(bootstrapAuthority!)
    await page.locator('[data-control-id="control.owner_username"]').fill('jesse@angryvibes.test')
    await page.locator('[data-control-id="control.owner_display_name"]').fill('Jesse Hall')
    await page.locator('[data-control-id="control.owner_enroll"]').click()

    await expect(page.getByRole('heading', { name: 'Export and verify recovery' })).toBeVisible({ timeout: 15_000 })
    const downloadPromise = page.waitForEvent('download')
    await page.locator('[data-control-id="control.recovery_download"]').click()
    const download = await downloadPromise
    const artifact = (await readFile(await download.path()!, 'utf8')).trim()
    expect(artifact).toMatch(/^nomen-recovery-v1\.[A-Za-z0-9_-]{43}$/u)

    await page.locator('[data-control-id="control.recovery_confirm_value"]').fill(artifact)
    await page.locator('[data-control-id="control.recovery_confirm"]').click()
    await expect(page.getByRole('heading', { name: 'Owner recovery is confirmed' })).toBeVisible()
    await expect(page.locator('[data-control-id="control.bootstrap_authority"]')).toHaveCount(0)

    const replay = await request.post('/nomen/v1/deployment/owner-enrollment:begin', {
      headers: {
        Authorization: `Bootstrap ${bootstrapAuthority}`,
        'Content-Type': 'application/json',
        'Idempotency-Key': crypto.randomUUID(),
      },
      data: { owner_id: crypto.randomUUID(), username: 'replay@angryvibes.test', display_name: 'Replay' },
    })
    expect(replay.status()).toBe(401)
    expect(await replay.json()).toMatchObject({ error: { type: 'authentication_required', reason: 'bootstrap_authority_required' } })
  })
})
