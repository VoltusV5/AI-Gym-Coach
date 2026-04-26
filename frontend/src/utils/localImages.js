/**
 * Ассеты из src/img/ (кириллические имена файлов из Figma).
 */
const modules = import.meta.glob('../img/*.{png,jpg,jpeg,webp,svg,JPG,PNG}', {
  eager: true,
  import: 'default'
})

function normalizePath(p) {
  return p.replace(/\\/g, '/').toLowerCase()
}

function firstUrl(predicate) {
  const hit = Object.entries(modules).find(([path]) => predicate(normalizePath(path)))
  return hit ? hit[1] : null
}

/** Аполлон слева на Welcome (подложка) */
export function getWelcomeApolloLeftUrl() {
  return firstUrl(
    (n) =>
      n.includes('welcomepage') ||
      (n.includes('апполон') && n.includes('welcom')) ||
      (n.includes('апполон') && n.includes('welcome'))
  )
}

/** Фото «чуть выше центра на главной» — на экране приветствия под контентом */
export function getWelcomeHeroPhotoUrl() {
  return firstUrl((n) => n.includes('главной') || n.includes('центра'))
}

/** Картинка «вопрос» внизу онбординга */
export function getOnboardingBottomIllustrationUrl() {
  return firstUrl((n) => n.includes('вопрос') && n.includes('заполнения'))
}

/** Восклицательный знак рядом с «вопросом» */
export function getOnboardingExclamationIllustrationUrl() {
  return firstUrl((n) => n.includes('воскл') && n.includes('заполнения'))
}

/** Аполлон фон тренировки — шапка списка упражнений / генерация плана */
export function getWorkoutBackgroundImageUrl() {
  return firstUrl((n) => n.includes('тренировк') && n.includes('апполон'))
}

/** Шапка с Аполлоном на любых экранах (питание, заметки и т.д.) — с запасным вариантом, если glob отличается в сборке */
export function getApolloHeaderImageUrl() {
  return (
    getWorkoutBackgroundImageUrl() ||
    getWelcomeApolloLeftUrl() ||
    firstUrl((n) => n.includes('апполон'))
  )
}

/** Иконки нижнего меню: main, workout, notes, nutrition, settings */
export function getHomeTabIconUrls() {
  const pick = (substr) => firstUrl((n) => n.includes('меню снизу') && n.includes(substr))

  return {
    main: pick('main') || firstUrl((n) => n.includes('меню') && n.includes('main')),
    workout: pick('гантел'),
    notes: pick('замет'),
    nutrition: pick('питан'),
    settings: pick('настрой')
  }
}
