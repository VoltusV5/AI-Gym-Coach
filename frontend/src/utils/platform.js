export const levelNames = {
  beginner: 'Начальный уровень',
  intermediate: 'Средний уровень',
  advanced: 'Продвинутый уровень',
}
export const goalNames = {
  muscle_gain: 'Набор мышц',
  weight_loss: 'Снижение веса',
  strength: 'Сила',
  endurance: 'Выносливость',
  fitness: 'Общая подготовка',
}
export const weekDays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']
export function money(minor, currency = 'RUB') {
  return new Intl.NumberFormat('ru-RU', { style: 'currency', currency }).format(Number(minor) / 100)
}
export function clone(value) {
  return JSON.parse(JSON.stringify(value))
}
export function localDate(date = new Date()) {
  return new Intl.DateTimeFormat('sv-SE', {
    timeZone: Intl.DateTimeFormat().resolvedOptions().timeZone,
  }).format(date)
}
export function emptyPlan() {
  return {
    title: '',
    days: [],
    schedule: {
      time_zone: Intl.DateTimeFormat().resolvedOptions().timeZone || 'Europe/Moscow',
      start_date: localDate(),
      entries: [],
    },
  }
}
export function planPayload(plan) {
  return {
    title: plan.title,
    days: clone(plan.days),
    schedule: clone(plan.schedule),
    ...(plan.version ? { version: plan.version } : {}),
  }
}
