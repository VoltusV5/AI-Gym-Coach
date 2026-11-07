/**
 * Append page items in server order while retaining the first occurrence of each ID.
 * @template {{id: string|number}} T
 * @param {T[]} current
 * @param {T[]|null} incoming
 * @returns {T[]}
 */
export function appendPage(current, incoming) {
  const seen = new Set(current.map(({ id }) => id))
  return current.concat((incoming ?? []).filter(({ id }) => {
    if (seen.has(id)) return false
    seen.add(id)
    return true
  }))
}

/** Generate a local ISO calendar date without shifting the date to UTC. */
export function localDate(date = new Date()) {
  return [
    date.getFullYear(),
    String(date.getMonth() + 1).padStart(2, '0'),
    String(date.getDate()).padStart(2, '0'),
  ].join('-')
}

/** Format a server timestamp in the viewer's timezone. */
export function dateLabel(date) {
  return new Intl.DateTimeFormat('ru', { dateStyle: 'medium' }).format(new Date(date))
}
