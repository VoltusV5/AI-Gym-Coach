import { test, expect } from '@playwright/test'
import { readFileSync } from 'node:fs'
const fixture = process.env.SPORTIK_FIXTURE_FILE
  ? JSON.parse(readFileSync(process.env.SPORTIK_FIXTURE_FILE, 'utf8'))
  : null
async function settled(page) {
  await expect(page.locator('ion-router-outlet > .ion-page:visible')).toHaveCount(1)
}
async function navigate(page, action) {
  const transition = page
    .locator('ion-router-outlet')
    .evaluate(
      (outlet) =>
        new Promise((resolve) =>
          outlet.addEventListener('ionNavDidChange', () => resolve(), { once: true }),
        ),
    )
  await Promise.all([transition, action()])
  await settled(page)
}
test.beforeEach(async ({ page }) => {
  test.skip(!fixture, 'Run from the isolated Go integration harness')
  await page.addInitScript((token) => localStorage.setItem('auth_token', token), fixture.token)
})
test('dashboard and compatible programs use actual server data', async ({ page }) => {
  await page.goto('/home')
  await expect(page.getByRole('heading', { name: 'Сегодня — шаг вперёд' })).toBeVisible()
  await expect(page.getByRole('heading', { name: 'Ваше следующее занятие' })).toBeVisible()
  await settled(page)
  await page.screenshot({ path: test.info().outputPath('dashboard.png'), fullPage: true })
  await page.goto('/recommendations')
  await expect(page.getByRole('heading', { name: 'Программы под ваши условия' })).toBeVisible()
  await expect(page.getByRole('alert')).toHaveCount(0)
})
test('personal plan, workout and progress persist through the Go API', async ({ page }) => {
  const title = `Браузерный план ${test.info().project.name}`
  await page.goto('/plans/new')
  await page.getByLabel('Название плана').fill(title)
  await page.getByRole('button', { name: 'Добавить тренировочный день' }).click()
  await page.getByRole('combobox', { name: /^Добавить или заменить упражнение/ }).selectOption('1')
  await page.getByRole('button', { name: 'Добавить упражнение', exact: true }).click()
  await page.getByRole('button', { name: 'Сохранить план' }).click()
  await expect(page).toHaveURL(/\/plans\/[0-9a-f-]{36}/)
  await page.getByRole('button', { name: 'Сделать активным' }).click()
  await expect(page.getByRole('button', { name: 'План активен' })).toBeVisible()
  await page.getByRole('link', { name: 'Начать тренировку' }).click()
  await expect(page.getByRole('button', { name: 'День 1', exact: true })).toBeVisible()
  await settled(page)
  await page.getByRole('button', { name: 'День 1', exact: true }).click()
  await page.getByLabel('Подход 1 · повторения').fill('10')
  await page.getByRole('button', { name: 'Завершить и сохранить тренировку' }).click()
  await expect(page.getByRole('heading', { name: 'Тренировка сохранена' })).toBeVisible()
  const trainingPath = new URL(page.url()).pathname
  await page.getByRole('link', { name: 'Открыть прогресс' }).click()
  await settled(page)
  await expect(page.getByRole('heading', { name: 'Ваш прогресс', exact: true })).toBeVisible()
  await settled(page)
  await page.getByRole('link', { name: 'Мои планы', exact: true }).click()
  await expect(
    page.getByRole('heading', { name: 'Мои тренировочные планы', exact: true }),
  ).toBeVisible()
  await settled(page)
  await page.getByRole('link', { name: title, exact: true }).click()
  await expect(page.getByRole('heading', { name: title, exact: true, level: 1 })).toBeVisible()
  await settled(page)
  await page.getByRole('link', { name: 'Начать тренировку' }).click()
  await expect(page.getByRole('button', { name: 'День 1', exact: true })).toBeVisible()
  await settled(page)
  await expect(page).toHaveURL(new RegExp(trainingPath))
  await page.getByRole('button', { name: 'День 1', exact: true }).click()
  await page.getByRole('button', { name: 'Завершить и сохранить тренировку' }).click()
  await expect(page.getByRole('heading', { name: 'Тренировка сохранена' })).toBeVisible()

  await page.getByRole('link', { name: 'Открыть прогресс' }).click()
  await settled(page)
  await expect(page.getByRole('heading', { name: 'Ваш прогресс' })).toBeVisible()
  await expect(page.getByRole('link', { name: /2026/ }).first()).toBeVisible()
  await page.reload()
  await expect(page.getByText('Завершённых тренировок')).toBeVisible()
  await page.screenshot({ path: test.info().outputPath('progress.png'), fullPage: true })
})
test('favorites and saved share library survive navigation', async ({ page }) => {
  await page.goto('/catalog')
  const card = page.locator('.platform-card').filter({ hasText: 'Version two' })
  await card.getByRole('link', { name: 'Version two', exact: true }).click()
  await expect(
    page.getByRole('heading', { name: 'Version two', exact: true, level: 1 }),
  ).toBeVisible()
  await settled(page)
  await expect(page.getByRole('button', { name: /В избранное|Убрать из избранного/ })).toBeVisible()
  const grant = page.getByRole('button', { name: 'Получить бесплатно', exact: true })
  if (await grant.isVisible()) await grant.click()
  await expect(page.getByRole('button', { name: 'Создать свою версию', exact: true })).toBeVisible()
  const favorite = page.getByRole('button', { name: 'В избранное', exact: true })
  if (await favorite.isVisible()) await favorite.click()
  await expect(page.getByRole('button', { name: 'Убрать из избранного' })).toBeVisible()
  await page.goto('/favorites')
  await expect(page.getByRole('link', { name: 'Version two', exact: true })).toBeVisible()
  await page.goto('/saved-shares')
  await expect(page.getByRole('heading', { name: 'Сохранённые ссылки' })).toBeVisible()
  await expect(page.getByRole('alert')).toHaveCount(0)
})

test('registration, logout and login use the actual authentication service', async ({ page }) => {
  const email = `browser-${test.info().project.name}-${Date.now()}@example.test`,
    password = 'SyntheticFixturePass2026'
  await page.goto('/login')
  await settled(page)
  await page.getByRole('button', { name: 'Выйти', exact: true }).click()
  await settled(page)
  await page.getByRole('button', { name: 'Создать аккаунт', exact: true }).click()
  await page.getByLabel('Электронная почта').fill(email)
  await page.getByLabel('Пароль', { exact: true }).fill(password)
  await page.getByRole('button', { name: 'Зарегистрироваться', exact: true }).click()
  await expect(page).toHaveURL(/\/home$/)
  await settled(page)
  await settled(page)
  await page.getByRole('button', { name: 'Выйти', exact: true }).click()
  await settled(page)
  await page.getByLabel('Электронная почта').fill(email)
  await page.getByLabel('Пароль', { exact: true }).fill(password)
  await page.locator('form').getByRole('button', { name: 'Войти', exact: true }).click()
  await expect(page).toHaveURL(/\/home$/)
  await settled(page)
})
test('sandbox purchase, private fork and refund follow server state', async ({ page }) => {
  await page.goto('/catalog')
  await page
    .locator('.platform-card')
    .filter({ hasText: 'Licensed fixture' })
    .getByRole('link', { name: 'Licensed fixture', exact: true })
    .click()
  await page.getByRole('button', { name: 'К тестовой оплате' }).click()
  await page.getByRole('button', { name: 'Подтвердить тестовую оплату' }).click()
  await expect(page.getByText('Состояние: Подтверждён')).toBeVisible()
  const checkoutPath = new URL(page.url()).pathname
  await page.getByRole('link', { name: 'Открыть купленную программу' }).click()
  await page.getByRole('button', { name: 'Создать свою версию' }).click()
  await expect(page).toHaveURL(/\/plans\/[0-9a-f-]{36}/)
  const forkPath = new URL(page.url()).pathname
  await page.goto(checkoutPath)
  page.once('dialog', (dialog) => dialog.accept())
  await page.getByRole('button', { name: 'Тестовый возврат' }).click()
  await expect(page.getByText('Состояние: Возвращён')).toBeVisible()
  await page.goto(forkPath)
  await expect(page.getByLabel('Название плана')).toHaveValue('Licensed fixture — моя версия')
})

test('declined sandbox payment grants no access', async ({ page }) => {
  await page.goto('/catalog')
  await page
    .locator('.platform-card')
    .filter({ hasText: 'Licensed fixture' })
    .getByRole('link', { name: 'Licensed fixture', exact: true })
    .click()
  await expect(page.getByRole('heading', { name: 'Licensed fixture', level: 1 })).toBeVisible()
  const programPath = new URL(page.url()).pathname
  await page.getByRole('button', { name: 'К тестовой оплате' }).click()
  await page.getByRole('button', { name: 'Отклонить тестовый платёж' }).click()
  await expect(page.getByText('Состояние: Отклонён')).toBeVisible()
  await page.goto(programPath)
  await expect(page.getByRole('button', { name: 'К тестовой оплате' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Создать свою версию' })).toHaveCount(0)
})

test('public snapshot is saved, copied and revoked through the library', async ({ page }) => {
  const title = `Снимок ${test.info().project.name}`
  await page.goto('/plans/new')
  await page.getByLabel('Название плана').fill(title)
  await page.getByRole('button', { name: 'Добавить тренировочный день' }).click()
  await page.getByRole('button', { name: 'Сохранить план' }).click()
  await expect(page).toHaveURL(/\/plans\/[0-9a-f-]{36}/)
  await page.getByRole('button', { name: 'Создать публичную ссылку' }).click()
  await expect(page.getByLabel('Ссылка', { exact: true })).toHaveValue(/\/share\//)
  const sharePath = new URL(await page.getByLabel('Ссылка', { exact: true }).inputValue()).pathname
  await page.goto(sharePath)
  await navigate(page, () =>
    page.getByRole('button', { name: 'Сохранить ссылку', exact: true }).click(),
  )
  await expect(page).toHaveURL(/\/saved-shares\/[0-9a-f-]{36}/)
  await expect(page.getByRole('heading', { name: title, exact: true, level: 1 })).toBeVisible()
  await page.getByRole('button', { name: 'Создать свою версию', exact: true }).click()
  await expect(page.getByLabel('Название плана')).toHaveValue(`${title} — моя версия`)
  await page.goto(sharePath)
  page.once('dialog', (dialog) => dialog.accept())
  await page.getByRole('button', { name: 'Отозвать ссылку', exact: true }).click()
  await expect(page.getByRole('alert')).toBeVisible()
  await page.goto('/saved-shares')
  await expect(
    page
      .locator('article')
      .filter({ hasText: title })
      .getByText('Ссылка отозвана или срок её действия закончился.'),
  ).toBeVisible()
})
