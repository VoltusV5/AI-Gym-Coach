export function createWorkoutDraft(plan, day, accountId) {
  return {
    account_id: accountId,
    key: crypto.randomUUID(),
    pending: null,
    input: {
      client_session_id: crypto.randomUUID(),
      plan_id: plan.id,
      plan_version: plan.version,
      plan_day_id: day.id,
      started_at: new Date().toISOString(),
      slots: day.exercises.map((exercise) => ({
        exercise_id: exercise.exercise_id,
        position: exercise.position,
        sets: Array.from({ length: exercise.sets }, () => ({
          reps: exercise.reps_min,
          weight_kg: exercise.weight_kg ?? 0,
          is_warmup: false,
        })),
      })),
    },
  }
}
export function freezeWorkout(draft) {
  if (!draft.pending)
    draft.pending = JSON.parse(
      JSON.stringify({ ...draft.input, finished_at: new Date().toISOString() }),
    )
  return draft.pending
}
export function readWorkoutDraft(accountId, planId) {
  try {
    const value = JSON.parse(sessionStorage.getItem(workoutStorageKey(accountId, planId)))
    return value?.account_id === accountId && value.input?.plan_id === planId ? value : null
  } catch {
    return null
  }
}

export function workoutStorageKey(accountId, planId) {
  return `sportik_workout_v2:${accountId}:${planId}`
}
export function storeWorkoutDraft(draft) {
  sessionStorage.setItem(
    workoutStorageKey(draft.account_id, draft.input.plan_id),
    JSON.stringify(draft),
  )
}
export function removeWorkoutDraft(accountId, planId) {
  sessionStorage.removeItem(workoutStorageKey(accountId, planId))
}
export function clearWorkoutDrafts() {
  for (let index = sessionStorage.length - 1; index >= 0; index -= 1) {
    const key = sessionStorage.key(index)
    if (key?.startsWith('sportik_workout_v2')) sessionStorage.removeItem(key)
  }
}
