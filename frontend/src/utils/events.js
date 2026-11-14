import api from '@/api/api'
export function sendProgramEvent(type, programId, surface, requestId = '') {
  const event = {
    event_id: crypto.randomUUID(),
    type,
    program_id: programId,
    occurred_at: new Date().toISOString(),
    context: { surface, ...(requestId ? { recommendation_request_id: requestId } : {}) },
  }
  void api.post('/api/v2/events', { events: [event] }).catch(() => undefined)
}
