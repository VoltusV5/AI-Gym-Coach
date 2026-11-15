import api from './api.js'

const base = '/api/v2'

/** Return HTTP status for an API error, or zero for transport failures. */
export function apiStatus(error) {
  return error?.response?.status ?? 0
}

/** Convert API failures to safe, actionable messages without exposing transport config. */
export function errorText(error, missing = 'Эти данные недоступны.') {
  const status = apiStatus(error)
  if (status === 401) return 'Сессия истекла. Войдите, чтобы продолжить.'
  if (status === 403) return 'У вашего аккаунта нет доступа к этому разделу.'
  if (status === 404) return missing
  if (status === 409) return 'Данные изменились на сервере. Проверьте последнюю версию и повторите действие.'
  if (status === 422 || status === 400) return 'Проверьте заполненные поля и выбранные значения.'
  if (status === 429) return 'Слишком много запросов. Повторите через некоторое время.'
  if (status >= 500) return 'Сервис временно недоступен. Повторите попытку позже.'
  return 'Не удалось связаться с сервером. Проверьте соединение и повторите попытку.'
}

/** Send a request to the deployed Go v2 contract and return its decoded body. */
export async function v2(method, path, { data, params, signal, key, anonymous, timeout } = {}) {
  const response = await api.request({
    method,
    url: base + path,
    data,
    params,
    signal,
    timeout,
    paramsSerializer: { indexes: null },
    headers: {
      ...(key && { 'Idempotency-Key': key }),
      ...(anonymous && { Authorization: null }),
    },
  })
  return response.data
}

/** Fetch every page of a small reference list using the server cursor. */
export async function getReference(path, params = {}) {
  const items = []
  const cursors = new Set()
  let cursor
  do {
    const page = await v2('GET', path, { params: { ...params, limit: 100, cursor } })
    items.push(...(page.items ?? []))
    cursor = page.next_cursor
    if (cursor && cursors.has(cursor)) throw new Error('Invalid reference cursor')
    cursors.add(cursor)
  } while (cursor)
  return items
}
