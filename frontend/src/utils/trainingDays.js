/** Проверка: дни тренировок заданы (массив с бэка или объект из старого мока) */
export function isTrainingDaysComplete(raw) {
  if (raw == null) return false
  if (Array.isArray(raw)) return raw.length > 0
  if (typeof raw === 'object') return Object.values(raw).some(Boolean)
  return false
}

/** Нормализация ответа API в объект { mon: bool, ... } для UI */
export function trainingDaysToSelection(raw) {
  const base = {
    mon: false,
    tue: false,
    wed: false,
    thu: false,
    fri: false,
    sat: false,
    sun: false
  }
  if (Array.isArray(raw)) {
    const next = { ...base }
    for (const k of raw) {
      if (k in next) next[k] = true
    }
    return next
  }
  if (raw && typeof raw === 'object') {
    return { ...base, ...raw }
  }
  return base
}

/** UI → массив ключей для PATCH profile */
export function selectionToTrainingDaysArray(sel) {
  return Object.entries(sel)
    .filter(([, v]) => v)
    .map(([k]) => k)
}
