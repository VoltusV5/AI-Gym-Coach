import api from '@/api/api'

export async function referenceItems(endpoint) {
  const items = []
  const cursors = new Set()
  let cursor
  do {
    const { data } = await api.get(endpoint, { params: { limit: 100, cursor } })
    if (!Array.isArray(data?.items)) throw new Error('Некорректный ответ справочника')
    items.push(...data.items)
    cursor = data.next_cursor
    if (cursor && cursors.has(cursor)) throw new Error('Не удалось загрузить справочник целиком')
    cursors.add(cursor)
  } while (cursor)
  return items
}
