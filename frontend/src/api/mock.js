import MockAdapter from 'axios-mock-adapter'
import api from './api'
import { getMockPlanGenerateResponse } from '@/mocks/planGenerate.mock'

const mock = new MockAdapter(api, { delayResponse: 500 })

function createEmptyMockProfile() {
  return {
    age: null,
    gender: null,
    height_cm: null,
    weight_kg: null,
    activity_level: null,
    injuries_notes: null,
    goal: null,
    fitness_level: null,
    training_days_map: null
  }
}

// Временное хранилище для профиля в памяти (мок) — одно на «сессию» гостя
let mockProfile = createEmptyMockProfile()

mock.onPost('/api/v1/auth/guest').reply(() => {
  // Новый гостевой вход = чистый профиль (иначе после «Заново пройти тест» опрос снова не стартует)
  mockProfile = createEmptyMockProfile()
  return [200, { token: `mock-jwt-${Date.now()}` }]
})

mock.onGet('/api/v1/profile').reply((config) => {
  // Проверка токена
  if (!config.headers.Authorization) {
    return [401, { message: 'Unauthorized' }]
  }
  return [200, { ...mockProfile }]
})

mock.onPatch('/api/v1/profile').reply((config) => {
  if (!config.headers.Authorization) {
    return [401, { message: 'Unauthorized' }]
  }

  try {
    const data = JSON.parse(config.data)
    // Обновляем мок-профиль только присланными полями
    mockProfile = { ...mockProfile, ...data }
    console.log('Mock Profile updated:', mockProfile)
    return [200, { ...mockProfile }]
  } catch (err) {
    return [400, { message: 'Invalid JSON' }]
  }
})

// Генерация плана (контракт ТЗ: день A — слоты с массивами вариаций)
mock.onPost('/api/v1/plans/generate').reply((config) => {
  if (!config.headers.Authorization) {
    return [401, { message: 'Unauthorized' }]
  }

  return new Promise((resolve) => {
    setTimeout(() => {
      resolve([200, getMockPlanGenerateResponse()])
    }, 2000)
  })
})

mock.onPost('/api/v1/workouts/complete').reply((config) => {
  if (!config.headers.Authorization) {
    return [401, { message: 'Unauthorized' }]
  }
  try {
    const body = config.data ? JSON.parse(config.data) : {}
    console.log('[mock] workouts/complete', body)
  } catch (_) {
    /* ignore */
  }
  return [200, { ok: true, saved_id: `mock-${Date.now()}` }]
})

mock.onPost('/api/v1/auth/register').reply((config) => {
  try {
    const body = config.data ? JSON.parse(config.data) : {}
    const token = `mock-jwt-${Date.now()}`
    return [200, { token, user: { id: 1, email: body.email || '', name: body.name || '' } }]
  } catch (_) {
    return [400, { message: 'Invalid JSON' }]
  }
})

mock.onPost('/api/v1/auth/login').reply((config) => {
  try {
    const body = config.data ? JSON.parse(config.data) : {}
    const token = `mock-jwt-${Date.now()}`
    return [200, { token, user: { id: 1, email: body.email || '', name: 'Mock User' } }]
  } catch (_) {
    return [400, { message: 'Invalid JSON' }]
  }
})

mock.onPost('/api/v1/auth/change-password').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  return [200, { ok: true }]
})

let mockNotes = []

mock.onGet('/api/v1/notes').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const items = [...mockNotes].sort(
    (a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime()
  )
  return [200, { items }]
})

mock.onPost('/api/v1/notes').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  try {
    const body = config.data ? JSON.parse(config.data) : {}
    const now = new Date().toISOString()
    const note = {
      id: `mock-note-${Date.now()}`,
      title: String(body.title ?? ''),
      body: String(body.body ?? ''),
      created_at: now,
      updated_at: now,
      deleted_at: null
    }
    mockNotes.unshift(note)
    return [200, note]
  } catch (_) {
    return [400, { message: 'Invalid JSON' }]
  }
})

mock.onPatch(/\/api\/v1\/notes\/[^/]+$/).reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  try {
    const id = String(config.url || '').split('/').pop()
    const body = config.data ? JSON.parse(config.data) : {}
    const idx = mockNotes.findIndex((n) => String(n.id) === id)
    if (idx < 0) return [404, { message: 'Not found' }]
    mockNotes[idx] = {
      ...mockNotes[idx],
      title: body.title != null ? String(body.title) : mockNotes[idx].title,
      body: body.body != null ? String(body.body) : mockNotes[idx].body,
      updated_at: new Date().toISOString()
    }
    return [200, mockNotes[idx]]
  } catch (_) {
    return [400, { message: 'Invalid JSON' }]
  }
})

mock.onDelete(/\/api\/v1\/notes\/[^/]+$/).reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const id = String(config.url || '').split('/').pop()
  const idx = mockNotes.findIndex((n) => String(n.id) === id)
  if (idx < 0) return [404, { message: 'Not found' }]
  mockNotes[idx] = { ...mockNotes[idx], deleted_at: new Date().toISOString() }
  mockNotes = mockNotes.filter((n) => !n.deleted_at)
  return [200, { ok: true }]
})

console.log('Mock API initialized')

export default mock
