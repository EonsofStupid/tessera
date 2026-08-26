import type { OwnerEnrollmentBegin } from './contracts'

function decodeBase64URL(value: string): ArrayBuffer {
  const normalized = value.replaceAll('-', '+').replaceAll('_', '/')
  const padded = normalized.padEnd(Math.ceil(normalized.length / 4) * 4, '=')
  const binary = atob(padded)
  return Uint8Array.from(binary, character => character.charCodeAt(0)).buffer
}

function encodeBase64URL(value: ArrayBuffer): string {
  const bytes = new Uint8Array(value)
  let binary = ''
  for (const byte of bytes) binary += String.fromCharCode(byte)
  return btoa(binary).replaceAll('+', '-').replaceAll('/', '_').replace(/=+$/u, '')
}

export async function createOwnerPasskey(options: OwnerEnrollmentBegin['publicKey']): Promise<unknown> {
  if (!window.PublicKeyCredential || !navigator.credentials) throw new Error('This browser does not support WebAuthn passkeys.')
  const publicKey: PublicKeyCredentialCreationOptions = {
    ...options,
    challenge: decodeBase64URL(options.challenge),
    user: { ...options.user, id: decodeBase64URL(options.user.id) },
    excludeCredentials: options.excludeCredentials?.map(credential => ({
      ...credential,
      id: decodeBase64URL(credential.id),
      transports: credential.transports as AuthenticatorTransport[] | undefined,
    })),
    authenticatorSelection: options.authenticatorSelection as AuthenticatorSelectionCriteria | undefined,
    attestation: options.attestation as AttestationConveyancePreference | undefined,
    extensions: options.extensions as AuthenticationExtensionsClientInputs | undefined,
  }
  const created = await navigator.credentials.create({ publicKey })
  if (!(created instanceof PublicKeyCredential) || !(created.response instanceof AuthenticatorAttestationResponse)) throw new Error('The authenticator did not return a registration credential.')
  return {
    id: created.id,
    rawId: encodeBase64URL(created.rawId),
    type: created.type,
    authenticatorAttachment: created.authenticatorAttachment,
    clientExtensionResults: created.getClientExtensionResults(),
    response: {
      clientDataJSON: encodeBase64URL(created.response.clientDataJSON),
      attestationObject: encodeBase64URL(created.response.attestationObject),
      transports: created.response.getTransports(),
    },
  }
}
