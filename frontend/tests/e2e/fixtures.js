export const ids = {
  program: '11111111-1111-4111-8111-111111111111',
  second: '22222222-2222-4222-8222-222222222222',
  plan: '33333333-3333-4333-8333-333333333333',
  day: '44444444-4444-4444-8444-444444444444',
  workout: '55555555-5555-4555-8555-555555555555',
  verification: '66666666-6666-4666-8666-666666666666',
}
export const equipment = [{ id: 'dumbbell', code: 'dumbbell', name: 'Гантели' }]
export const exercise = { id: 3, name: 'Жим гантелей', muscle_group: 'Грудь', equipment_ids: ['dumbbell'], is_active: true }
export const profile = {
  version: 4, age: 28, height_cm: 171, weight_kg: 72.5, gender: 'Мужской',
  activity_level: 'Средняя активность', injuries_notes: false, goal: 'Набрать мышцы',
  level: 'beginner', training_days: [1, 3], equipment_ids: ['dumbbell'],
  session_duration_min: 45, excluded_exercise_ids: [],
}
export const prescription = { exercise_id: 3, position: 0, sets: 3, reps_min: 8, reps_max: 10, rest_sec: 60, weight_kg: 12.5 }
export const plan = {
  id: ids.plan, version: 3, title: 'Мой тренировочный план', source_type: 'manual',
  source_program_version_id: null, is_active: false,
  days: [{ id: ids.day, title: 'Силовая', position: 0, exercises: [prescription] }],
  schedule: { start_date: '2026-11-07', time_zone: 'Europe/Moscow', entries: [] },
}
export const program = {
  id: ids.program, title: 'Сила с гантелями', trainer: { id: 5, display_name: 'Тестовый тренер', verified: true },
  goal: 'Набрать мышцы', level: 'beginner', duration_weeks: 4, sessions_per_week: 3,
  equipment, price_minor: 0, currency: 'RUB', rating: null, review_count: 0, published_version_id: ids.second,
}
export const pageOf = (items, next_cursor = null) => ({ items, next_cursor })
export const token = (guest) => `fixture.${Buffer.from(JSON.stringify({ sub: guest ? '1' : '2', is_anonymous: guest })).toString('base64url')}.fixture`
export const failure = (code) => ({ code, message: 'Fixture API failure', request_id: 'fixture-test' })

/** Serve the Go v1/v2 wire shape and fail a test on unmatched network calls. */
export async function mockApi(page, handle, { guest = false } = {}) {
  await page.route((url) => url.pathname.startsWith('/api/'), async (route) => {
    const request = route.request()
    const url = new URL(request.url())
    const method = request.method()
    const custom = await handle({ request, method, url })
    if (custom) {
      const [status, json] = custom
      if (json === undefined) return route.abort()
      return route.fulfill({ status, json })
    }
    if (url.pathname === '/api/v1/auth/guest') return route.fulfill({ json: { token: token(guest) } })
    if (url.pathname === '/api/v1/auth/login') return route.fulfill({ json: { token: token(false) } })
    if (url.pathname === '/api/v1/profile') return route.fulfill({ json: { ...profile, fitness_level: 'Новичок' } })
    if (url.pathname === '/api/v2/equipment') return route.fulfill({ json: pageOf(equipment) })
    if (url.pathname === '/api/v2/exercises') return route.fulfill({ json: pageOf([exercise]) })
    return route.fulfill({ status: 501, json: failure('UNMATCHED_FIXTURE') })
  })
}
