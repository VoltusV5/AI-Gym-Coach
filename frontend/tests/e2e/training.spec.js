import { test, expect } from '@playwright/test'
import { ids, profile, plan, program, pageOf, mockApi, failure, token } from './fixtures.js'

test('profile retains a fractional edit after network failure and rebases on 409', async ({ page }) => {
  let server = { ...profile }
  const attempts = []
  await mockApi(page, ({ method, url, request }) => {
    if (url.pathname !== '/api/v2/me/profile') return
    if (method === 'GET') return [200, server]
    const body = request.postDataJSON()
    attempts.push(body)
    if (attempts.length === 1) return [0]
    if (attempts.length === 2) {
      server = { ...server, age: 29, version: 5 }
      return [409, failure('CONFLICT')]
    }
    server = { ...server, ...body, version: 6 }
    return [200, server]
  })
  await page.goto('/settings/profile')
  const weight = page.getByLabel('Вес, кг')
  await weight.fill('71.25')
  await page.getByRole('button', { name: 'Сохранить профиль' }).click()
  await expect(page.getByRole('alert')).toContainText('Проверьте соединение')
  await expect(weight).toHaveValue('71.25')
  await page.getByRole('button', { name: 'Сохранить профиль' }).click()
  await page.getByRole('button', { name: 'Перечитать профиль, сохранив мой ввод' }).click()
  await expect(page.getByLabel('Возраст')).toHaveValue('29')
  await expect(weight).toHaveValue('71.25')
  await page.getByRole('button', { name: 'Сохранить профиль' }).click()
  await expect(page.getByText('Профиль сохранён')).toBeVisible()
  expect(attempts).toEqual([{ version: 4, weight_kg: 71.25 }, { version: 4, weight_kg: 71.25 }, { version: 5, weight_kg: 71.25 }])
  await page.reload()
  await expect(weight).toHaveValue('71.25')
})

test('expired profile session offers sign-in without retrying the write', async ({ page }) => {
  let writes = 0
  await mockApi(page, ({ method, url }) => {
    if (url.pathname !== '/api/v2/me/profile') return
    if (method === 'GET') return [200, profile]
    writes++
    return [401, failure('UNAUTHORIZED')]
  })
  await page.goto('/settings/profile')
  await page.getByLabel('Вес, кг').fill('75')
  await page.getByRole('button', { name: 'Сохранить профиль' }).click()
  await expect(page.getByText('Сессия истекла. Войдите, чтобы продолжить.')).toBeVisible()
  expect(writes).toBe(1)
  await page.getByRole('link', { name: 'Войти', exact: true }).click()
  await page.getByLabel('Email').fill('fixture@example.test')
  await page.getByLabel('Пароль', { exact: true }).fill('fixture-password')
  await page.getByRole('button', { name: 'Войти', exact: true }).click()
  await expect(page).toHaveURL(/\/settings\/profile$/)
  await expect(page.getByLabel('Вес, кг')).toHaveValue('75')
})

test('catalog preserves filters and cursor, deduplicates and obtains access after login', async ({ page }) => {
  let granted = false
  let grants = 0
  let shouldFail = true
  await mockApi(page, async ({ method, url, request }) => {
    const path = url.pathname
    if (path === '/api/v2/catalog/programs') {
      if (shouldFail) {
        shouldFail = false
        return [503, failure('UNAVAILABLE')]
      }
      const { searchParams: query } = url
      expect(query.get('price_type')).toBe('free')
      if (query.get('goal') === 'Сбросить вес') return [200, pageOf([])]
      if (query.get('cursor')) {
        expect(query.get('goal')).toBe('Набрать мышцы')
        return [200, pageOf([program, { ...program, id: ids.second, title: 'Другая программа' }])]
      }
      return [200, pageOf([program], 'cursor-2')]
    }
    if (path === '/api/v2/catalog/programs/' + ids.program) {
      return [200, { ...program, description: 'Описание для теста', access: granted ? 'free' : 'preview',
        allowed_actions: granted ? [] : ['access'], days: granted ? plan.days : undefined }]
    }
    if (path.startsWith('/api/v2/catalog/programs/')) return [404, failure('NOT_FOUND')]
    if (method === 'POST' && path === '/api/v2/programs/' + ids.program + '/access') {
      expect(request.headers().authorization).toContain('Bearer')
      grants++
      granted = true
      return [201, { id: ids.second, program_id: ids.program, status: 'active' }]
    }
  }, { guest: true })
  await page.goto('/marketplace')
  await page.getByRole('button', { name: 'Повторить' }).click()
  await page.getByLabel('Цель').fill('Сбросить вес')
  await page.getByRole('button', { name: 'Применить фильтры' }).click()
  await expect(page.getByText(/Программы не найдены/)).toBeVisible()
  await page.getByLabel('Цель').fill('Набрать мышцы')
  await page.getByRole('button', { name: 'Применить фильтры' }).click()
  await page.getByRole('button', { name: 'Показать ещё' }).click()
  await expect(page.getByRole('heading', { name: program.title })).toHaveCount(1)
  await expect(page.getByRole('heading', { name: 'Другая программа' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Показать ещё' })).toHaveCount(0)
  await page.getByRole('link', { name: 'Открыть программу ' + program.title }).click()
  await expect(page.getByText('Описание для теста')).toBeVisible()
  await page.getByRole('link', { name: 'Войти и получить доступ' }).click()
  await page.getByLabel('Email').fill('fixture@example.test')
  await page.getByLabel('Пароль', { exact: true }).fill('fixture-password')
  await page.getByRole('button', { name: 'Войти', exact: true }).click()
  expect(grants).toBe(0)
  await page.getByRole('button', { name: 'Получить бесплатно' }).click()
  await expect(page.getByText('Программа доступна')).toBeVisible()
  expect(grants).toBe(1)
  await page.reload()
  await expect(page.getByText('Программа доступна')).toBeVisible()
  await page.getByRole('link', { name: 'Назад' }).click()
  await expect(page.getByLabel('Цель')).toHaveValue('Набрать мышцы')
  await page.goto('/marketplace/programs/' + ids.second)
  await expect(page.getByRole('alert')).toContainText('Эта программа недоступна')
})

test('plan editor copies days, assigns schedule, preserves a conflicting draft', async ({ page }) => {
  let record = structuredClone(plan)
  let changes = 0
  await mockApi(page, ({ url, method, request }) => {
    if (url.pathname !== '/api/v2/me/plans/' + ids.plan) return
    if (method === 'GET') return [200, record]
    const data = request.postDataJSON()
    if (++changes === 1) {
      record.version++
      record.schedule.start_date = '2026-11-08'
      return [409, failure('CONFLICT')]
    }
    expect(data.days.map((day) => day.position)).toEqual([0, 1])
    expect(data.days[1].id).not.toBe(data.days[0].id)
    expect(data.schedule.entries[0].plan_day_id).toBe(data.days[1].id)
    expect(data.schedule.start_date).toBe('2026-11-08')
    record = { ...record, ...data, version: data.version + 1 }
    return [200, record]
  })
  await page.goto(`/plans/${ids.plan}/edit`)
  await page.getByLabel('Название плана').fill('План с расписанием')
  await page.getByRole('button', { name: 'Копировать день' }).click()
  await page.getByRole('combobox', { name: 'Пн', exact: true }).selectOption({ label: 'Силовая (копия)' })
  await page.getByRole('button', { name: 'Сохранить план' }).click()
  await page.getByRole('button', { name: 'Перечитать план' }).click()
  await expect(page.getByLabel('Название плана')).toHaveValue('План с расписанием')
  await page.getByRole('button', { name: /Применить мой ввод/ }).click()
  await page.getByRole('button', { name: 'Сохранить план' }).click()
  await expect(page.getByText('План сохранён')).toBeVisible()
  await page.reload()
  await expect(page.getByLabel('Название плана')).toHaveValue('План с расписанием')
})

test('generation waits, handles network and ML failure, opens the shared editor', async ({ page }) => {
  const keys = []
  await mockApi(page, async ({ url, request }) => {
    if (url.pathname === '/api/v2/me/profile') return [200, profile]
    if (url.pathname === '/api/v2/plans/generate') {
      keys.push(request.headers()['idempotency-key'])
      expect(request.postDataJSON()).toEqual({ profile_version: 4 })
      if (keys.length === 1) {
        await new Promise((resolve) => setTimeout(resolve, 600))
        return [0]
      }
      if (keys.length === 2) return [503, failure('ML_UNAVAILABLE')]
      return [201, { plan_id: ids.plan, version: 1, model_version: 'fixture', warnings: [] }]
    }
    if (url.pathname === '/api/v2/me/plans/' + ids.plan) return [200, { ...plan, source_type: 'ai' }]
  })
  await page.goto('/plans/generate')
  await page.getByRole('button', { name: 'Сгенерировать план' }).click()
  await expect(page.getByRole('button', { name: 'Готовим план…' })).toBeDisabled()
  await expect(page.getByRole('alert')).toContainText('Проверьте соединение')
  await page.getByRole('button', { name: 'Повторить генерацию' }).click()
  await expect(page.getByRole('alert')).toContainText('Генератор временно недоступен')
  await page.getByRole('button', { name: 'Повторить генерацию' }).click()
  await page.getByRole('link', { name: 'Открыть редактор' }).click()
  await expect(page.getByLabel('Название плана')).toHaveValue(plan.title)
  expect(keys).toHaveLength(3)
  expect(new Set(keys).size).toBe(1)
})

test('workout retries the same payload, progress shows confirmed metrics and history', async ({ page }) => {
  const requests = []
  let workout
  await mockApi(page, ({ url, request }) => {
    if (url.pathname === '/api/v2/me/plans/' + ids.plan) return [200, plan]
    if (url.pathname === '/api/v2/workouts/complete') {
      const body = request.postDataJSON()
      requests.push({ body, key: request.headers()['idempotency-key'] })
      workout = { ...body, id: ids.workout }
      if (requests.length === 1) return [0]
      return [201, { session_id: ids.workout, metrics: { volume_kg: 100, working_sets: 1, exercise_count: 1, duration_sec: 90 }, highlights: [], new_achievements: [] }]
    }
    if (url.pathname === '/api/v2/workouts/' + ids.workout) return [200, workout]
    if (url.pathname === '/api/v2/workouts') return [200, pageOf([workout])]
    if (url.pathname === '/api/v2/progress') return [200, {
      workout_count: 1, volume_kg: 100, adherence: null, streak_weeks: 1, weight_change_kg: null, highlights: [],
    }]
  })
  await page.goto(`/workout/session/${ids.plan}/${ids.day}`)
  await page.getByLabel('Подход 1 · повторы').fill('8')
  await page.getByRole('button', { name: 'Завершить тренировку' }).click()
  await expect(page.getByRole('alert')).toContainText('Проверьте соединение')
  await page.reload()
  await page.getByRole('button', { name: 'Повторить сохранение' }).click()
  await expect(page.getByRole('heading', { name: 'Тренировка сохранена' })).toBeVisible()
  await expect(page.getByText('100 кг')).toBeVisible()
  expect(requests).toHaveLength(2)
  expect(requests[0]).toEqual(requests[1])
  await page.getByRole('link', { name: 'Смотреть прогресс' }).click()
  await expect(page.getByText('100 кг')).toBeVisible()
  await page.getByRole('link', { name: 'История тренировок' }).click()
  await expect(page.getByRole('link', { name: 'Открыть тренировку' })).toHaveCount(1)
})

test('legacy notes follow v1 offsets without duplicates and stop at the last page', async ({ page }) => {
  const offsets = []
  await mockApi(page, ({ url }) => {
    if (url.pathname !== '/api/v1/notes') return
    const offset = Number(url.searchParams.get('offset'))
    const limit = Number(url.searchParams.get('limit'))
    offsets.push(offset)
    const ids = offset ? [20, 21] : Array.from({ length: 20 }, (_, index) => index + 1)
    return [200, {
      items: ids.map((id) => ({ id, title: 'Заметка №' + id, body: 'Текст', created_at: '2026-11-07T08:00:00Z' })),
      total: 22, limit, offset,
    }]
  })
  await page.goto('/notes')
  await page.getByRole('button', { name: 'Показать ещё' }).click()
  await expect(page.getByText('Заметка №20', { exact: true })).toHaveCount(1)
  await expect(page.getByText('Заметка №21', { exact: true })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Показать ещё' })).toHaveCount(0)
  expect(offsets).toEqual([0, 20])
})

test('legacy workout keeps input on API error and saves through the v1 contract', async ({ page }) => {
  const workoutPlan = {
    split: 'fullbody',
    plan: [{ day: 'A', day_name: 'Силовая', exercises: [[{ id: 3, exercise_name: 'Жим гантелей', weight: 10 }]] }],
  }
  let writes = 0
  await mockApi(page, ({ url, request }) => {
    if (url.pathname === '/api/v2/me/plans') return [200, pageOf([])]
    if (url.pathname !== '/api/v1/workouts/complete') return
    expect(request.postDataJSON().slots).toEqual([{ slot_index: 0, exercise_id: 3, sets: [{ reps: 8, weight_kg: 10 }] }])
    if (++writes === 1) return [503, failure('UNAVAILABLE')]
    return [200, { ok: true, saved_id: 'fixture-workout', new_achievements: [], session_highlights: [] }]
  })
  await page.goto('/home')
  await expect(page.getByRole('button', { name: 'Начать тренировку' })).toBeDisabled()
  await page.evaluate((plan) => localStorage.setItem('sportik_workout_plan:2', JSON.stringify(plan)), workoutPlan)
  await page.reload()
  await page.getByRole('button', { name: 'Начать тренировку' }).click()
  const reps = page.getByPlaceholder('Повторы').first()
  await reps.fill('8')
  await page.getByRole('button', { name: 'Завершить тренировку' }).click()
  await expect(page.getByRole('alert')).toContainText('Сервис временно недоступен')
  await expect(reps).toHaveValue('8')
  await page.getByRole('button', { name: 'Завершить тренировку' }).click()
  await expect(page.getByRole('heading', { name: 'Тренировка завершена!' })).toBeVisible()
  await expect(page).toHaveURL(/\/workout\/success$/)
  expect(writes).toBe(2)
})

test('plan creation reconciles a lost response without creating a duplicate', async ({ page }) => {
  let saved
  let writes = 0
  await mockApi(page, ({ url, method, request }) => {
    if (url.pathname === '/api/v2/me/plans/' + ids.plan) return [200, saved]
    if (url.pathname !== '/api/v2/me/plans') return
    if (method === 'GET') return [200, pageOf(saved ? [saved] : [])]
    writes++
    saved = { ...plan, ...request.postDataJSON() }
    return [0]
  })
  await page.goto('/plans/new')
  await page.getByLabel('Название плана').fill('Созданный план')
  await page.getByRole('button', { name: 'Создать план', exact: true }).click()
  await page.getByRole('button', { name: 'Проверить и повторить' }).click()
  await expect(page.getByLabel('Название плана')).toHaveValue('Созданный план')
  expect(writes).toBe(1)
})

test('moderation clears stale conflicts and removes private documents when access is revoked', async ({ page }) => {
  const first = { id: ids.verification, trainer_id: 5, status: 'pending', qualification: 'Первая заявка', document_refs: ['private/first'] }
  const second = { ...first, id: ids.second, trainer_id: 6, qualification: 'Вторая заявка', document_refs: ['private/second'] }
  await mockApi(page, ({ url, method }) => {
    if (url.pathname === '/api/v2/admin/verifications' && url.searchParams.get('cursor')) return [403, failure('FORBIDDEN')]
    if (url.pathname === '/api/v2/admin/verifications') return [200, pageOf([first, second], 'next')]
    if (method === 'PATCH' && url.pathname.endsWith(ids.verification)) return [409, failure('CONFLICT')]
  })
  await page.goto('/admin/verifications')
  const firstRow = page.locator('article').filter({ hasText: 'Первая заявка' })
  await firstRow.getByRole('button', { name: 'Рассмотреть' }).click()
  await firstRow.getByRole('combobox', { name: 'Статус' }).selectOption('rejected')
  await firstRow.getByRole('textbox', { name: 'Причина' }).fill('Тестовая причина')
  await firstRow.getByRole('button', { name: 'Сохранить решение' }).click()
  await expect(firstRow.getByRole('alert')).toContainText('Данные изменились')
  const secondRow = page.locator('article').filter({ hasText: 'Вторая заявка' })
  await secondRow.getByRole('button', { name: 'Рассмотреть' }).click()
  await secondRow.getByRole('combobox', { name: 'Статус' }).selectOption('approved')
  await secondRow.getByRole('textbox', { name: 'Причина' }).fill('Тестовая причина')
  await expect(secondRow.getByRole('button', { name: 'Сохранить решение' })).toBeEnabled()
  await page.getByRole('button', { name: 'Показать ещё' }).click()
  await expect(page.getByRole('alert')).toContainText('нет доступа')
  await expect(page.getByText('private/first')).toHaveCount(0)
  await expect(page.getByText('private/second')).toHaveCount(0)
})

test('an in-flight nutrition refresh cannot merge data into the next account', async ({ page }) => {
  let statsRequested
  const statsStarted = new Promise((resolve) => { statsRequested = resolve })
  let releaseStats
  const statsPending = new Promise((resolve) => { releaseStats = resolve })
  await mockApi(page, async ({ method, url }) => {
    if (url.pathname === '/api/v2/me/plans') return [200, pageOf([])]
    if (url.pathname === '/api/v1/nutrition/entries' && method === 'POST') {
      return [201, { id: 12, title: 'Старая запись', meal_type: 'lunch' }]
    }
    if (url.pathname === '/api/v1/nutrition/entries' || url.pathname === '/api/v1/nutrition/favorites') {
      return [200, { items: [] }]
    }
    if (url.pathname === '/api/v1/nutrition/dashboard') return [200, {}]
    if (url.pathname === '/api/v1/nutrition/stats') {
      statsRequested()
      await statsPending
      return [200, {}]
    }
  }, { guest: true })
  await page.goto('/home')
  await expect(page.getByRole('link', { name: 'Открыть библиотеку планов' })).toBeVisible()
  const start = await page.evaluate(() => {
    // Vite serves these source paths in the browser test server.
    const nutritionModule = '/src/stores/nutrition.js'
    return import(nutritionModule).then(({ useNutritionStore }) => {
      globalThis.sportikPendingNutrition = useNutritionStore().addEntry({ title: 'Старая запись' }).catch((err) => err.code)
      return true
    })
  })
  expect(start).toBe(true)
  await statsStarted
  try {
    await page.evaluate(async () => {
      const authModule = '/src/stores/auth.js'
      const { useAuthStore } = await import(authModule)
      await useAuthStore().login({ email: 'fixture@example.test', password: 'fixture-password' })
    })
  } finally {
    releaseStats()
  }
  const result = await page.evaluate(async () => {
    const nutritionModule = '/src/stores/nutrition.js'
    const { useNutritionStore } = await import(nutritionModule)
    return { error: await globalThis.sportikPendingNutrition, entries: useNutritionStore().entries }
  })
  expect(result).toEqual({ error: 'ERR_CANCELED', entries: [] })
})

test('a late expiry keeps a different credential saved by another tab', async ({ page }) => {
  await mockApi(page, () => undefined, { guest: true })
  await page.goto('/auth/login')
  await expect(page.getByRole('heading', { name: 'Вход в Спортик' })).toBeVisible()
  const kept = await page.evaluate(async ({ expired, replacement }) => {
    const authModule = '/src/stores/auth.js'
    const storageModule = '/src/utils/auth.js'
    const [{ useAuthStore }, { getToken, saveToken }] = await Promise.all([import(authModule), import(storageModule)])
    await saveToken(replacement)
    await useAuthStore().expire('Bearer ' + expired)
    return (await getToken()) === replacement
  }, { expired: token(true), replacement: token(false) })
  expect(kept).toBe(true)
})

test('a delayed sign-in cannot override a more recent account choice', async ({ page }) => {
  let resolveStarted
  const started = new Promise((resolve) => { resolveStarted = resolve })
  let resolveFirst
  const firstPending = new Promise((resolve) => { resolveFirst = resolve })
  await mockApi(page, async ({ url, request }) => {
    if (url.pathname !== '/api/v1/auth/login') return
    if (request.postDataJSON().email === 'first@example.test') {
      resolveStarted()
      await firstPending
      return [200, { token: token(true) }]
    }
    return [200, { token: token(false) }]
  }, { guest: true })
  await page.goto('/auth/login')
  await expect(page.getByRole('heading', { name: 'Вход в Спортик' })).toBeVisible()
  await page.evaluate(async () => {
    const module = '/src/stores/auth.js'
    const { useAuthStore } = await import(module)
    globalThis.sportikFirstLogin = useAuthStore().login({ email: 'first@example.test', password: 'fixture-password' })
      .then(() => 'unexpected-success', (err) => err.code)
  })
  await started
  try {
    await page.evaluate(async () => {
      const module = '/src/stores/auth.js'
      const { useAuthStore } = await import(module)
      await useAuthStore().login({ email: 'second@example.test', password: 'fixture-password' })
    })
  } finally {
    resolveFirst()
  }
  const current = await page.evaluate(async () => {
    const module = '/src/stores/auth.js'
    const { useAuthStore } = await import(module)
    return { first: await globalThis.sportikFirstLogin, owner: useAuthStore().accountId }
  })
  expect(current).toEqual({ first: 'ERR_CANCELED', owner: '2' })
})
