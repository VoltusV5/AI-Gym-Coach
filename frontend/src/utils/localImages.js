
const modules = import.meta.glob('../img
export function getWelcomeApolloLeftUrl() {
  return firstUrl(
    (n) =>
      n.includes('welcomepage') ||
      (n.includes('апполон') && n.includes('welcom')) ||
      (n.includes('апполон') && n.includes('welcome'))
  )
}


export function getWelcomeHeroPhotoUrl() {
  return firstUrl((n) => n.includes('главной') || n.includes('центра'))
}


export function getOnboardingBottomIllustrationUrl() {
  return firstUrl((n) => n.includes('вопрос') && n.includes('заполнения'))
}


export function getOnboardingExclamationIllustrationUrl() {
  return firstUrl((n) => n.includes('воскл') && n.includes('заполнения'))
}


export function getWorkoutBackgroundImageUrl() {
  return firstUrl((n) => n.includes('тренировк') && n.includes('апполон'))
}


export function getApolloHeaderImageUrl() {
  return (
    getWorkoutBackgroundImageUrl() ||
    getWelcomeApolloLeftUrl() ||
    firstUrl((n) => n.includes('апполон'))
  )
}


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