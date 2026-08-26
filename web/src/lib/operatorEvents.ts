import { accessToken } from './auth'

const sessionKey = 'nomen.ui.operator_session'
const sequenceKey = 'nomen.ui.operator_sequence'

function sessionID(): string {
  let id = sessionStorage.getItem(sessionKey)
  if (!id) {
    id = crypto.randomUUID()
    sessionStorage.setItem(sessionKey, id)
  }
  return id
}

function nextSequence(): number {
  const sequence = Number.parseInt(sessionStorage.getItem(sequenceKey) ?? '0', 10) + 1
  sessionStorage.setItem(sequenceKey, String(sequence))
  return sequence
}

export type SemanticEvent = {
  routeId: string
  eventType: 'route_opened' | 'control_activated' | 'suggestion_accepted' | 'guide_advanced' | 'action_result'
  controlId?: string
  actionId?: string
  outcome?: 'observed' | 'accepted' | 'refused' | 'failed'
  resourceRevision?: string
}

export function emitOperatorEvent(event: SemanticEvent): void {
  const token = accessToken()
  if (!token) return
  const body = {
    events: [{
      schema_version: 1,
      event_id: crypto.randomUUID(),
      session_id: sessionID(),
      sequence: nextSequence(),
      occurred_at: new Date().toISOString(),
      route_id: event.routeId,
      control_id: event.controlId,
      event_type: event.eventType,
      action_id: event.actionId,
      resource_revision: event.resourceRevision,
      outcome: event.outcome ?? 'observed',
    }],
  }
  void fetch('/nomen/v1/operator/events', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    credentials: 'same-origin',
    body: JSON.stringify(body),
    keepalive: true,
  })
}
