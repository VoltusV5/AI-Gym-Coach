import MockAdapter from 'axios-mock-adapter'
import api from './api'
import { getMockPlanGenerateResponse } from '@/mocks/planGenerate.mock'

// По умолчанию mock отвечает 404 на любой URL без handler — ломает новые эндпоинты.
// passthrough: неизвестные запросы уходят на реальный backend (или сеть).
const mock = new MockAdapter(api, { delayResponse: 500, onNoMatch: 'passthrough' })

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
let mockNutritionEntries = []
let mockNutritionFavorites = []
let mockNutritionGoal = null
let mockNutritionDishes = [
  { title: 'Куриная грудка', protein_g: 22.5, fat_g: 1.9, carbs_g: 0 },
  { title: 'Овсянка', protein_g: 13.5, fat_g: 5.9, carbs_g: 68.7 },
  { title: 'Творог 5%', protein_g: 17, fat_g: 5, carbs_g: 3 },
  { title: 'Банан', protein_g: 0.74, fat_g: 0.29, carbs_g: 23 }
]
mockNutritionDishes = mockNutritionDishes.map((x, idx) => ({ id: idx + 1, ...x, base_grams: 100, calories: Number((x.protein_g * 4 + x.fat_g * 9 + x.carbs_g * 4).toFixed(1)) }))
let mockWaterByDay = {}
let mockWeightByDay = {}

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

mock.onGet('/api/v1/nutrition/entries').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const params = config.params || {}
  const day = String(params.day || new Date().toISOString().slice(0, 10))
  const items = [...mockNutritionEntries]
    .filter((x) => String(x.consumed_at || '').slice(0, 10) === day)
    .sort((a, b) => b.id - a.id)
  return [200, { items, day }]
})

mock.onPost('/api/v1/nutrition/entries').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  try {
    const body = config.data ? JSON.parse(config.data) : {}
    const now = new Date().toISOString()
    const dayPart = body.day ? String(body.day).slice(0, 10) : now.slice(0, 10)
    const consumed_at = body.consumed_at || `${dayPart}T12:00:00.000Z`
    const item = {
      id: Date.now(),
      dish_id: body.dish_id || null,
      grams: Number(body.grams || 100),
      meal_type: body.meal_type || 'snack',
      title: String(body.title || 'Без названия'),
      protein_g: Number(body.protein_g || 0),
      fat_g: Number(body.fat_g || 0),
      carbs_g: Number(body.carbs_g || 0),
      calories: Number(body.calories || (Number(body.protein_g || 0) * 4 + Number(body.fat_g || 0) * 9 + Number(body.carbs_g || 0) * 4)),
      consumed_at,
      created_at: now
    }
    mockNutritionEntries.unshift(item)
    return [200, item]
  } catch (_) {
    return [400, { message: 'Invalid JSON' }]
  }
})

mock.onPatch(/\/api\/v1\/nutrition\/entries\/\d+$/).reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const id = Number(String(config.url || '').split('/').pop())
  const idx = mockNutritionEntries.findIndex((x) => x.id === id)
  if (idx < 0) return [404, { message: 'Not found' }]
  const body = config.data ? JSON.parse(config.data) : {}
  mockNutritionEntries[idx] = { ...mockNutritionEntries[idx], ...body, grams: Number(body.grams || mockNutritionEntries[idx].grams || 100) }
  return [200, { ok: true }]
})

mock.onDelete(/\/api\/v1\/nutrition\/entries\/\d+$/).reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const id = Number(String(config.url || '').split('/').pop())
  mockNutritionEntries = mockNutritionEntries.filter((x) => x.id !== id)
  return [200, { ok: true }]
})

mock.onGet('/api/v1/nutrition/favorites').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  return [200, { items: [...mockNutritionFavorites] }]
})

mock.onPost('/api/v1/nutrition/favorites').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const body = config.data ? JSON.parse(config.data) : {}
  const item = {
    id: Date.now(),
    title: String(body.title || 'Без названия'),
    protein_g: Number(body.protein_g || 0),
    fat_g: Number(body.fat_g || 0),
    carbs_g: Number(body.carbs_g || 0),
    unit_type: body.unit_type || 'gram'
  }
  mockNutritionFavorites.unshift(item)
  return [200, { id: item.id, ok: true }]
})

mock.onDelete(/\/api\/v1\/nutrition\/favorites\/\d+$/).reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const id = Number(String(config.url || '').split('/').pop())
  mockNutritionFavorites = mockNutritionFavorites.filter((x) => x.id !== id)
  return [200, { ok: true }]
})

mock.onGet('/api/v1/nutrition/goals').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  if (!mockNutritionGoal) return [204]
  return [200, mockNutritionGoal]
})

mock.onPost('/api/v1/nutrition/goals/recalculate').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const body = config.data ? JSON.parse(config.data) : {}
  const calories = 2300
  const delta = Number(body.target_delta_kg || 0)
  mockNutritionGoal = {
    protein_g: 150,
    fat_g: 70,
    carbs_g: 250,
    calories: Math.max(1200, calories + delta * 220),
    updated_at: new Date().toISOString()
  }
  return [200, mockNutritionGoal]
})

mock.onGet('/api/v1/nutrition/stats').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const sum = mockNutritionEntries.reduce(
    (acc, x) => {
      acc.protein += Number(x.protein_g || 0)
      acc.fat += Number(x.fat_g || 0)
      acc.carbs += Number(x.carbs_g || 0)
      acc.calories += Number(x.calories || 0)
      return acc
    },
    { protein: 0, fat: 0, carbs: 0, calories: 0 }
  )
  return [
    200,
    {
      day: { protein_g: sum.protein, fat_g: sum.fat, carbs_g: sum.carbs },
      week_avg: { protein_g: sum.protein / 7, fat_g: sum.fat / 7, carbs_g: sum.carbs / 7, calories: sum.calories / 7 },
      month_avg: { protein_g: sum.protein / 30, fat_g: sum.fat / 30, carbs_g: sum.carbs / 30, calories: sum.calories / 30 }
    }
  ]
})

mock.onGet('/api/v1/nutrition/dishes/search').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const params = config.params || {}
  const search = String(params.q || '').trim().toLowerCase()
  const limit = Math.max(1, Number(params.limit || 30))
  let items = [...mockNutritionDishes]
  if (search) items = items.filter((x) => String(x.title || '').toLowerCase().includes(search))
  return [200, { items: items.slice(0, limit) }]
})

mock.onGet('/api/v1/nutrition/dishes/mine').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  return [200, { items: [...mockNutritionDishes] }]
})

mock.onPatch(/\/api\/v1\/nutrition\/dishes\/\d+$/).reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const id = Number(String(config.url || '').match(/\/(\d+)(?:\?|$)/)?.[1] || 0)
  const raw = config.data
  const body =
    raw && typeof raw === 'object' ? raw : raw ? JSON.parse(String(raw)) : {}
  const idx = mockNutritionDishes.findIndex((x) => Number(x.id) === id)
  if (idx < 0) return [404, { message: 'Not found' }]
  const title = String(body.title ?? mockNutritionDishes[idx].title).trim()
  const protein_g = Number(body.protein_g ?? mockNutritionDishes[idx].protein_g)
  const fat_g = Number(body.fat_g ?? mockNutritionDishes[idx].fat_g)
  const carbs_g = Number(body.carbs_g ?? mockNutritionDishes[idx].carbs_g)
  const base_grams = Math.max(1, Number(body.base_grams ?? mockNutritionDishes[idx].base_grams ?? 100))
  let calories = Number(body.calories)
  if (!Number.isFinite(calories) || calories <= 0) {
    calories = protein_g * 4 + fat_g * 9 + carbs_g * 4
  }
  mockNutritionDishes[idx] = {
    ...mockNutritionDishes[idx],
    title,
    protein_g,
    fat_g,
    carbs_g,
    calories,
    base_grams
  }
  return [200, mockNutritionDishes[idx]]
})

mock.onDelete(/\/api\/v1\/nutrition\/dishes\/\d+$/).reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const id = Number(String(config.url || '').match(/\/(\d+)(?:\?|$)/)?.[1] || 0)
  const before = mockNutritionDishes.length
  mockNutritionDishes = mockNutritionDishes.filter((x) => Number(x.id) !== id)
  if (mockNutritionDishes.length === before) return [404, { message: 'Not found' }]
  return [204]
})

mock.onPost('/api/v1/nutrition/dishes').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const raw = config.data
  const body =
    raw && typeof raw === 'object' ? raw : raw ? JSON.parse(String(raw)) : {}
  const base_grams = Number(body.base_grams || 100)
  const title = String(body.title || '').trim()
  if (!title) return [400, { message: 'Title is required' }]
  const protein_g = Number(body.protein_g || 0)
  const fat_g = Number(body.fat_g || 0)
  const carbs_g = Number(body.carbs_g || 0)
  const calories = Number(body.calories || (protein_g * 4 + fat_g * 9 + carbs_g * 4))
  const existing = mockNutritionDishes.find((x) => String(x.title).toLowerCase() === title.toLowerCase())
  if (existing) {
    existing.protein_g = protein_g
    existing.fat_g = fat_g
    existing.carbs_g = carbs_g
    existing.calories = calories
    existing.base_grams = base_grams
    return [200, existing]
  }
  const item = { id: Date.now(), title, protein_g, fat_g, carbs_g, calories, base_grams }
  mockNutritionDishes.unshift(item)
  return [200, item]
})

mock.onPost('/api/v1/nutrition/water').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const body = config.data ? JSON.parse(config.data) : {}
  const day = String(body.day || new Date().toISOString().slice(0, 10))
  mockWaterByDay[day] = Number(body.amount_ml || 0)
  return [200, { ok: true }]
})

mock.onPost('/api/v1/nutrition/weight').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const body = config.data ? JSON.parse(config.data) : {}
  const day = String(body.day || new Date().toISOString().slice(0, 10))
  mockWeightByDay[day] = Number(body.weight_kg || 0)
  return [200, { ok: true }]
})

mock.onGet('/api/v1/nutrition/dashboard').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const params = config.params || {}
  const day = String(params.day || new Date().toISOString().slice(0, 10))
  const dayItems = mockNutritionEntries.filter((x) => String(x.consumed_at || '').slice(0, 10) === day)
  const totals = dayItems.reduce((acc, x) => {
    acc.protein_g += Number(x.protein_g || 0)
    acc.fat_g += Number(x.fat_g || 0)
    acc.carbs_g += Number(x.carbs_g || 0)
    acc.calories += Number(x.calories || 0)
    return acc
  }, { protein_g: 0, fat_g: 0, carbs_g: 0, calories: 0 })
  const goal = Number(mockNutritionGoal?.calories || 0)
  const gp = Number(mockNutritionGoal?.protein_g || 0)
  const gf = Number(mockNutritionGoal?.fat_g || 0)
  const gc = Number(mockNutritionGoal?.carbs_g || 0)
  const weightDays = Object.keys(mockWeightByDay).sort().reverse()
  const lastWeightDay = weightDays[0] || null
  const lastW = lastWeightDay ? Number(mockWeightByDay[lastWeightDay] || 70) : 70
  const wGoal = Math.min(4500, Math.max(1500, Math.round(lastW * 33)))
  const today = new Date(`${day}T00:00:00Z`)
  const last = lastWeightDay ? new Date(`${lastWeightDay}T00:00:00Z`) : null
  const needWeightReminder = !last || (today - last) / (1000 * 60 * 60 * 24) >= 3
  return [200, {
    day,
    today: totals,
    goal: {
      calories: goal,
      protein_g: gp,
      fat_g: gf,
      carbs_g: gc,
      remaining_calories: goal - totals.calories
    },
    water: {
      amount_ml: Number(mockWaterByDay[day] || 0),
      goal_ml: wGoal,
      goal_liters: Math.round((wGoal / 1000) * 100) / 100
    },
    weight: {
      last_weight_kg: lastWeightDay ? mockWeightByDay[lastWeightDay] : null,
      last_weight_day: lastWeightDay,
      need_weight_reminder: needWeightReminder
    }
  }]
})

mock.onGet('/api/v1/nutrition/analytics').reply((config) => {
  if (!config.headers.Authorization) return [401, { message: 'Unauthorized' }]
  const days = Number(config.params?.days || 30)
  const cutoff = new Date(Date.now() - days * 24 * 60 * 60 * 1000)
  const byDay = {}
  for (const item of mockNutritionEntries) {
    const day = String(item.consumed_at || '').slice(0, 10)
    const d = new Date(`${day}T00:00:00Z`)
    if (d < cutoff) continue
    if (!byDay[day]) byDay[day] = { day, protein_g: 0, fat_g: 0, carbs_g: 0, calories: 0 }
    byDay[day].protein_g += Number(item.protein_g || 0)
    byDay[day].fat_g += Number(item.fat_g || 0)
    byDay[day].carbs_g += Number(item.carbs_g || 0)
    byDay[day].calories += Number(item.calories || 0)
  }
  const food = Object.values(byDay).sort((a, b) => String(b.day).localeCompare(String(a.day)))
  const weight = Object.entries(mockWeightByDay).map(([day, weight_kg]) => ({ day, weight_kg })).sort((a, b) => String(b.day).localeCompare(String(a.day)))
  const water = Object.entries(mockWaterByDay).map(([day, amount_ml]) => ({ day, amount_ml })).sort((a, b) => String(b.day).localeCompare(String(a.day)))
  return [200, { days, food, weight, water }]
})

console.log('Mock API initialized')

export default mock
