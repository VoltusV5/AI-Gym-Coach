import { localDate } from './pages.js'

export const sourceLabels = { manual: 'Ручной', ai: 'AI-план', fork: 'Копия программы' }

/** Generate a UUID v4 for retries and new plan days, including on a local HTTP origin. */
export function newId() {
  const bytes = crypto.getRandomValues(new Uint8Array(16))
  bytes[6] = (bytes[6] & 15) | 64
  bytes[8] = (bytes[8] & 63) | 128
  const hex = Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('')
  return [hex.slice(0, 8), hex.slice(8, 12), hex.slice(12, 16), hex.slice(16, 20), hex.slice(20)].join('-')
}

/** Return the browser's IANA timezone for backend scheduling and reports. */
export function timeZone() {
  return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC'
}

/** Create a new day with an ID that can be referenced by the schedule before saving. */
export function newDay(title) {
  return { id: newId(), title, position: 0, exercises: [] }
}

/** Produce the write DTO with current array ordering and no presentation fields. */
export function writeDays(days) {
  return days.map((day, position) => ({
    id: day.id,
    title: day.title.trim(),
    position,
    exercises: day.exercises.map((exercise, position) => ({
      exercise_id: exercise.exercise_id,
      position,
      sets: exercise.sets,
      reps_min: exercise.reps_min,
      reps_max: exercise.reps_max,
      weight_kg: exercise.weight_kg === '' ? null : exercise.weight_kg,
      rest_sec: exercise.rest_sec,
    })),
  }))
}

/** Build an initial manual plan using the viewer's local date and timezone. */
export function newPlan(title) {
  return {
    title: title.trim(),
    days: [newDay('День 1')],
    schedule: { time_zone: timeZone(), start_date: localDate(), entries: [] },
  }
}

/** Replace or remove a single ISO weekday in a plan schedule. */
export function assignDay(entries, weekday, dayId) {
  const result = entries.filter((entry) => entry.weekday !== weekday)
  if (dayId) result.push({ weekday, plan_day_id: dayId })
  return result.sort((a, b) => a.weekday - b.weekday)
}

/**
 * Transfer changed editor fields to a fresh version. Ordered prescriptions remain
 * one unit because server slots have positions and no stable per-slot identity.
 */
export function rebasePlan(original, draft, latest) {
  const changed = (a, b) => JSON.stringify(a) !== JSON.stringify(b)
  const result = { ...latest, schedule: { ...latest.schedule } }
  for (const field of ['title', 'days']) {
    if (changed(original[field], draft[field])) result[field] = draft[field]
  }
  for (const field of ['time_zone', 'start_date']) {
    if (original.schedule[field] !== draft.schedule[field]) result.schedule[field] = draft.schedule[field]
  }
  const dayIds = new Set(result.days.map(({ id }) => id))
  let entries = latest.schedule.entries
  for (let weekday = 1; weekday <= 7; weekday++) {
    const assigned = (plan) => plan.schedule.entries.find((entry) => entry.weekday === weekday)?.plan_day_id
    const next = assigned(draft)
    if (next !== assigned(original)) entries = assignDay(entries, weekday, next)
  }
  result.schedule.entries = entries.filter((entry) => dayIds.has(entry.plan_day_id))
  return result
}
