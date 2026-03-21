import MockAdapter from 'axios-mock-adapter'
import api from './api'

const mock = new MockAdapter(api, { delayResponse: 500 })

// Временное хранилище для профиля в памяти (мок)
let mockProfile = {
  age: null,
  gender: null,
  height_cm: null,
  weight_kg: null,
  activity_level: null,
  injuries_notes: null,
  goal: null,
  fitness_level: null,
  training_days_map: null,
}

mock.onPost('/auth/guest').reply(200, {
  token: 'mock-jwt-token-abc-123',
})

mock.onGet('/profile').reply((config) => {
  // Проверка токена
  if (!config.headers.Authorization) {
    return [401, { message: 'Unauthorized' }]
  }
  return [200, { ...mockProfile }]
})

mock.onPatch('/profile').reply((config) => {
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

// Генерация плана
mock.onPost('/api/v1/plans/generate').reply((config) => {
  if (!config.headers.Authorization) {
    return [401, { message: 'Unauthorized' }]
  }

  // Имитируем долгий процесс
  return new Promise((resolve) => {
    setTimeout(() => {
      resolve([200, { success: true }])
    }, 2000)
  })
})

console.log('Mock API initialized')

export default mock
