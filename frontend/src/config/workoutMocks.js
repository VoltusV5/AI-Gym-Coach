/**
 * Моки тренировки: без плана с API подставляется тот же JSON, что в getMockPlanGenerateResponse() (ТЗ).
 *
 * VITE_USE_WORKOUT_MOCKS=true  — если плана нет, на /home показывается демо fullbody A/B/C (сессия — день A).
 * VITE_USE_WORKOUT_MOCKS=false — только ответ бэкенда; без generate список пуст.
 *
 * В .env / Docker build-args.
 */
export function workoutMocksEnabled() {
  const v = import.meta.env.VITE_USE_WORKOUT_MOCKS
  if (v === 'false' || v === '0') return false
  if (v === 'true' || v === '1') return true
  /* По умолчанию: в dev чаще нужен превью UI; в production-сборке задайте явно false */
  return import.meta.env.DEV === true
}
