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

mock.onPost('/auth/guest').reply(() => {
  // Новый гостевой вход = чистый профиль (иначе после «Заново пройти тест» опрос снова не стартует)
  mockProfile = createEmptyMockProfile()
  return [200, { token: `mock-jwt-${Date.now()}` }]
})

mock.onGet('/profile').reply((config) => {
  // Проверка токена
  if (!config.headers.Authorization) {
    return [401, { message: 'Unauthorized' }]
  }
  return [200, { ...mockProfile }]
})

mock.onPost('/profile').reply((config) => {
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

console.log('Mock API initialized')

export default mock
