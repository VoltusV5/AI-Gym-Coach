import { newId } from './plans.js'

/** Create a workout draft pinned to the exact revision shown when training starts. */
export function workoutDraft(plan, day, now = new Date()) {
  return {
    client_session_id: newId(), plan_id: plan.id, plan_version: plan.version, plan_day_id: day.id,
    started_at: now.toISOString(),
    slots: day.exercises.map((exercise) => ({
      exercise_id: exercise.exercise_id,
      position: exercise.position,
      sets: [{ reps: '', weight_kg: exercise.weight_kg ?? 0, is_warmup: false }],
    })),
  }
}

/** Build an immutable completion payload, excluding exercises and rows without actual reps. */
export function completion(draft, now = new Date()) {
  return {
    ...draft,
    finished_at: now.toISOString(),
    slots: draft.slots.map((slot) => ({
      ...slot,
      sets: slot.sets.filter((set) => Number.isInteger(set.reps) && set.reps > 0)
        .map((set) => ({ ...set, weight_kg: set.weight_kg === '' ? 0 : set.weight_kg })),
    })).filter((slot) => slot.sets.length),
  }
}

/** Read an account-scoped draft or result from the current tab. */
export function readTab(key) {
  try {
    return JSON.parse(sessionStorage.getItem(key))
  } catch {
    return null
  }
}

/** Persist draft state within the current tab. In-memory data remains usable if storage is unavailable. */
export function writeTab(key, value) {
  try {
    if (value === null) sessionStorage.removeItem(key)
    else sessionStorage.setItem(key, JSON.stringify(value))
  } catch {
    // The browser can disable storage in private or embedded contexts.
  }
}

export const numberLabel = new Intl.NumberFormat('ru', { maximumFractionDigits: 2 })
