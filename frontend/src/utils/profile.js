/**
 * @typedef {object} Profile
 * @property {number} version
 * @property {number|null} age
 * @property {number|null} height_cm
 * @property {number|null} weight_kg
 * @property {string|null} gender
 * @property {string|null} activity_level
 * @property {boolean|null} injuries_notes
 * @property {string|null} goal
 * @property {'beginner'|'intermediate'|'advanced'|null} level
 * @property {number[]} training_days
 * @property {string[]} equipment_ids
 * @property {number|null} session_duration_min
 * @property {number[]} excluded_exercise_ids
 */

export const weekdays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']
export const levels = { beginner: 'Новичок', intermediate: 'Любитель', advanced: 'Продвинутый' }

/**
 * Build a typed PATCH containing changed fields and the version read by the editor.
 * @param {Profile} current
 * @param {Profile} draft
 * @param {Array<keyof Profile>} fields
 * @returns {Partial<Profile> & {version: number}}
 */
export function profilePatch(current, draft, fields) {
  const patch = { version: current.version }
  for (const field of fields) {
    const value = draft[field]
    if (value !== null && value !== '' && JSON.stringify(value) !== JSON.stringify(current[field])) {
      patch[field] = value
    }
  }
  return patch
}
