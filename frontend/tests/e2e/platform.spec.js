import { test, expect } from '@playwright/test'

const programId = '18000000-0000-4000-8000-000000000001'
const versionId = '19000000-0000-4000-8000-000000000001'
const planId = '20000000-0000-4000-8000-000000000001'
const dayId = '21000000-0000-4000-8000-000000000001'
const card = {
  id: programId,
  title: 'Сильнее за четыре недели',
  trainer: { id: 201, display_name: 'Тренер Алексей', verified: true },
  goal: 'muscle_gain',
  level: 'beginner',
  duration_weeks: 4,
  sessions_per_week: 3,
  equipment: [],
  price_minor: 0,
  currency: 'RUB',
  rating: null,
  review_count: 0,
  published_version_id: versionId,
}
const profile = {
  version: 1,
  age: 25,
  height_cm: 175,
  weight_kg: 75,
  gender: 'м',
  activity_level: 'Средняя активность',
  injuries_notes: false,
  goal: 'muscle_gain',
  level: 'beginner',
  training_days: [1, 3],
  equipment_ids: [],
  session_duration_min: 60,
  excluded_exercise_ids: [],
}
const plan = {
  id: planId,
  version: 1,
  title: 'Мой силовой план',
  source_type: 'manual',
  source_program_version_id: null,
  is_active: false,
  days: [{ id: dayId, title: 'Первый день', position: 0, exercises: [] }],
  schedule: { time_zone: 'Europe/Moscow', start_date: '2026-11-13', entries: [] },
}

async function setup(page, { registered = true, roles = ['user'] } = {}) {
  await page.addInitScript(() => localStorage.setItem('auth_token', 'synthetic-test-session'))
  await page.route('**/api/**', async (route) => {
    const request = route.request(),
      path = new URL(request.url()).pathname
    if (!path.startsWith('/api/')) return route.continue()
    let data
    if (path === '/api/v2/me/session')
      data = {
        id: 101,
        email: 'fixture@example.com',
        roles,
        is_anonymous: !registered,
        blocked: false,
      }
    else if (path === '/api/v1/profile')
      data = {
        version: 1,
        age: 25,
        height_cm: 175,
        weight_kg: 75,
        gender: 'м',
        activity_level: 'Средняя активность',
        injuries_notes: false,
        goal: 'Набрать мышцы',
        fitness_level: 'Новичок',
        training_days_map: ['mon', 'wed'],
      }
    else if (path === '/api/v2/equipment')
      data = { items: [{ id: 'barbell', code: 'barbell', name: 'Штанга' }], next_cursor: null }
    else if (path === '/api/v2/exercises')
      data = { items: [{ id: 1, name: 'Приседания', equipment_ids: [] }], next_cursor: null }
    else if (path === '/api/v2/catalog/programs') data = { items: [card], next_cursor: null }
    else if (path === `/api/v2/catalog/programs/${programId}`)
      data = {
        ...card,
        description: 'Три тренировки в неделю для уверенного старта.',
        access: 'preview',
        allowed_actions: ['access'],
        days: [],
        allow_public_remix: false,
        is_favorite: false,
      }
    else if (path === '/api/v2/me/profile') data = profile
    else if (path === `/api/v2/me/plans/${planId}`) data = plan
    else if (path === '/api/v2/events') data = { accepted: 1, duplicate_count: 0 }
    else return route.fulfill({ status: 404, json: { error: { message: 'Not found' } } })
    await route.fulfill({ json: data })
  })
}

test('catalog, filters and confirmed free access', async ({ page }) => {
  await setup(page)
  let granted = false
  await page.route(`**/api/v2/programs/${programId}/access`, (route) => {
    granted = true
    return route.fulfill({
      status: 201,
      json: { id: 'entitlement', program_id: programId, status: 'active' },
    })
  })
  await page.route(`**/api/v2/catalog/programs/${programId}`, (route) =>
    route.fulfill({
      json: {
        ...card,
        description: 'Три тренировки в неделю для уверенного старта.',
        access: granted ? 'free' : 'preview',
        days: [],
        is_favorite: false,
      },
    }),
  )
  await page.goto('/catalog')
  await expect(page.getByRole('heading', { name: 'Найдите свой тренировочный план' })).toBeVisible()
  await page.getByRole('combobox', { name: /^Уровень/ }).selectOption('beginner')
  const filtered = page.waitForRequest(
    (request) =>
      request.url().includes('/catalog/programs?') && request.url().includes('level=beginner'),
  )
  await page.getByRole('button', { name: 'Показать программы' }).click()
  await filtered
  await page.getByRole('link', { name: card.title }).click()
  await page.getByRole('button', { name: 'Получить бесплатно' }).click()
  await expect(page.getByRole('button', { name: 'Создать свою версию' })).toBeVisible()
  await page.screenshot({
    path: `/tmp/sportik-catalog-${test.info().project.name}.png`,
    fullPage: true,
  })
})

test('profile conflict keeps the edited form', async ({ page }) => {
  await setup(page)
  await page.route('**/api/v2/me/profile', async (route) => {
    if (route.request().method() === 'PATCH')
      return route.fulfill({
        status: 409,
        json: { error: { code: 'CONFLICT', message: 'Conflict' } },
      })
    return route.fulfill({ json: profile })
  })
  await page.goto('/profile')
  await page.getByLabel('Вес, кг', { exact: true }).fill('76.5')
  await page.getByRole('button', { name: 'Сохранить профиль' }).click()
  await expect(page.getByRole('alert')).toContainText('Данные изменились')
  await expect(page.getByLabel('Вес, кг', { exact: true })).toHaveValue('76.5')
})

test('plan copy has a fresh identity and failed save keeps the draft', async ({ page }) => {
  await setup(page)
  let sent
  await page.route(`**/api/v2/me/plans/${planId}`, async (route) => {
    if (route.request().method() === 'PATCH') {
      sent = route.request().postDataJSON()
      return route.fulfill({ status: 503, json: { error: { message: 'Unavailable' } } })
    }
    return route.fulfill({ json: plan })
  })
  await page.goto(`/plans/${planId}`)
  await page.getByRole('button', { name: 'Копировать день' }).click()
  await expect(page.getByLabel('Название дня')).toHaveCount(2)
  await page.getByRole('button', { name: 'Сохранить план' }).click()
  await expect(page.getByRole('alert')).toBeVisible()
  expect(sent.days[0].id).not.toBe(sent.days[1].id)
  await expect(page.getByLabel('Название дня').nth(1)).toHaveValue('Первый день — копия')
})

test('ordinary account is denied administrative navigation', async ({ page }) => {
  await setup(page)
  await page.goto('/admin/users')
  await expect(page.getByRole('heading', { name: 'Этот раздел недоступен' })).toBeVisible()
  await expect(page).toHaveURL(/access-denied/)
})

test('pending workout can be recovered after the plan was archived', async ({ page }) => {
  await setup(page)
  const pending = {
    client_session_id: crypto.randomUUID(),
    plan_id: planId,
    plan_version: 1,
    plan_day_id: dayId,
    started_at: '2026-11-14T10:00:00Z',
    finished_at: '2026-11-14T10:30:00Z',
    slots: [{ exercise_id: 1, position: 0, sets: [{ reps: 8, weight_kg: 20, is_warmup: false }] }],
  }
  await page.addInitScript(
    ({ pending, planId }) =>
      sessionStorage.setItem(
        `sportik_workout_v2:101:${planId}`,
        JSON.stringify({ account_id: 101, input: pending, pending }),
      ),
    { pending, planId },
  )
  await page.route(`**/api/v2/me/plans/${planId}`, (route) =>
    route.fulfill({ status: 404, json: { error: { message: 'Archived' } } }),
  )
  let submitted
  await page.route('**/api/v2/workouts/complete', (route) => {
    submitted = route.request().postDataJSON()
    return route.fulfill({
      status: 201,
      json: {
        session_id: 'saved',
        metrics: { volume_kg: 160, working_sets: 1, duration_sec: 1800 },
        highlights: [],
      },
    })
  })
  await page.goto(`/training/${planId}`)
  await page.getByRole('button', { name: 'Повторить отправку результата' }).click()
  await expect(page.getByRole('heading', { name: 'Тренировка сохранена' })).toBeVisible()
  expect(submitted).toEqual(pending)
})
test('admin edits a seeded exercise with a relative image path', async ({ page }) => {
  await setup(page, { roles: ['user', 'admin'] })
  const exercise = {
    id: 1,
    name: 'Приседания',
    muscle_group: 'Ноги',
    muscle_subgroup: null,
    equipment_ids: [],
    safe_for_injuries: true,
    image_url: '/static/images/fixture.webp',
    video_url: null,
    is_active: true,
  }
  let submitted
  await page.route('**/api/v2/admin/exercises**', (route) => {
    if (route.request().method() === 'PATCH') {
      submitted = route.request().postDataJSON()
      return route.fulfill({ json: { ...exercise, ...submitted } })
    }
    return route.fulfill({ json: { items: [exercise], next_cursor: null } })
  })
  await page.goto('/admin/exercises')
  await page.getByRole('button', { name: 'Изменить', exact: true }).click()
  await page.getByLabel('Название', { exact: true }).fill('Приседания со штангой')
  await page.getByLabel('Причина изменения').fill('Уточнение названия')
  page.once('dialog', (dialog) => dialog.accept())
  await page.getByRole('button', { name: 'Сохранить запись' }).click()
  await expect(page.getByRole('status').filter({ hasText: 'Сохранено' })).toBeVisible()
  expect(submitted.image_url).toBe('/static/images/fixture.webp')
})
test('analytics includes current operations without a future date', async ({ page }) => {
  await setup(page, { roles: ['user', 'admin'] })
  let query
  await page.route('**/api/v2/admin/analytics**', (route) => {
    query = new URL(route.request().url()).searchParams
    return route.fulfill({
      json: {
        users: 1,
        active_users: 1,
        workouts: 1,
        published_programs: 1,
        sales: 1,
        gross_minor: 10000,
        refund_minor: 0,
        net_minor: 10000,
        commission_accrued_minor: 2000,
        commission_paid_minor: 0,
        currency: 'RUB',
        top_programs: [],
        top_trainers: [],
      },
    })
  })
  await page.goto('/admin/analytics')
  await expect(page.getByText('Оборот до возвратов', { exact: true })).toBeVisible()
  expect(query.has('to')).toBe(false)
  await expect(page.getByLabel('По текущее время, включая сегодняшние операции')).toBeChecked()
})
test('moderation shows the review behind a complaint', async ({ page }) => {
  await setup(page, { roles: ['user', 'admin'] })
  await page.route('**/api/v2/admin/reports**', (route) =>
    route.fulfill({
      json: {
        items: [
          {
            id: crypto.randomUUID(),
            target_type: 'review',
            target_id: crypto.randomUUID(),
            reason: 'Проверьте отзыв',
            status: 'open',
            target_review: {
              id: crypto.randomUUID(),
              program_id: programId,
              overall: 2,
              text: 'Проверяемый текст отзыва',
              status: 'hidden',
            },
          },
        ],
        next_cursor: null,
      },
    }),
  )
  await page.goto('/admin/reports')
  await expect(page.getByRole('heading', { name: 'Отзыв, на который пожаловались' })).toBeVisible()
  await expect(page.getByText('Проверяемый текст отзыва', { exact: true })).toBeVisible()
})
