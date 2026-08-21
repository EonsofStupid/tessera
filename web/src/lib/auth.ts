import { parseEnvironment, parseTokenResponse, type Environment } from './contracts'

const tokenKey = 'tessera.ui.access_token'
const verifierKey = 'tessera.ui.pkce_verifier'
const stateKey = 'tessera.ui.oauth_state'

function encode(bytes: Uint8Array): string {
  return btoa(String.fromCharCode(...bytes)).replaceAll('+', '-').replaceAll('/', '_').replaceAll('=', '')
}

async function digest(value: string): Promise<string> {
  return encode(new Uint8Array(await crypto.subtle.digest('SHA-256', new TextEncoder().encode(value))))
}

function randomValue(length = 48): string {
  const bytes = new Uint8Array(length)
  crypto.getRandomValues(bytes)
  return encode(bytes)
}

export function accessToken(): string | null {
  return sessionStorage.getItem(tokenKey)
}

export function clearSession(): void {
  sessionStorage.removeItem(tokenKey)
  sessionStorage.removeItem(verifierKey)
  sessionStorage.removeItem(stateKey)
}

export async function loadEnvironment(): Promise<Environment> {
  const response = await fetch('/ui/console/assets/environment.json', { credentials: 'same-origin' })
  if (!response.ok) throw new Error('environment_unavailable')
  return parseEnvironment(await response.json())
}

function callbackURL(): string {
  return `${window.location.origin}/ui/console/auth/callback`
}

export async function beginSignIn(environment: Environment): Promise<void> {
  if (!environment.issuer || !environment.clientid) throw new Error('sign_in_not_configured')
  const verifier = randomValue()
  const state = randomValue(24)
  sessionStorage.setItem(verifierKey, verifier)
  sessionStorage.setItem(stateKey, state)

  const params = new URLSearchParams({
    client_id: environment.clientid,
    redirect_uri: callbackURL(),
    response_type: 'code',
    scope: 'openid profile email',
    code_challenge: await digest(verifier),
    code_challenge_method: 'S256',
    state,
  })
  window.location.assign(`${environment.issuer.replace(/\/$/, '')}/oauth/v2/authorize?${params}`)
}

export async function finishSignIn(environment: Environment): Promise<boolean> {
  const params = new URLSearchParams(window.location.search)
  const code = params.get('code')
  const state = params.get('state')
  const expectedState = sessionStorage.getItem(stateKey)
  const verifier = sessionStorage.getItem(verifierKey)
  if (!code) return false
  if (!environment.issuer || !environment.clientid || !state || state !== expectedState || !verifier) {
    clearSession()
    throw new Error('sign_in_state_invalid')
  }

  const body = new URLSearchParams({
    grant_type: 'authorization_code',
    client_id: environment.clientid,
    redirect_uri: callbackURL(),
    code,
    code_verifier: verifier,
  })
  const response = await fetch(`${environment.issuer.replace(/\/$/, '')}/oauth/v2/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body,
  })
  if (!response.ok) {
    clearSession()
    throw new Error('token_exchange_failed')
  }
  const result = parseTokenResponse(await response.json())
  sessionStorage.setItem(tokenKey, result.access_token)
  sessionStorage.removeItem(verifierKey)
  sessionStorage.removeItem(stateKey)
  window.history.replaceState({}, '', '/ui/console/')
  return true
}
