
export function workoutMocksEnabled() {
  const v = import.meta.env.VITE_USE_WORKOUT_MOCKS
  if (v === 'false' || v === '0') return false
  if (v === 'true' || v === '1') return true

  return import.meta.env.DEV === true
}