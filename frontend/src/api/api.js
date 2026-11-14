import axios from 'axios'

const api = axios.create({ baseURL: import.meta.env.VITE_API_URL?.trim() || '', timeout: 30000, paramsSerializer: { indexes: null } })
let activeToken = null
let revision = 0
let expired = () => undefined

export function setAuthHeader(token) {
  if (activeToken !== token) revision += 1
  activeToken = token || null
  if (activeToken) api.defaults.headers.common.Authorization = `Bearer ${activeToken}`
  else delete api.defaults.headers.common.Authorization
}

export function onSessionExpired(handler) {
  expired = handler
}

api.interceptors.request.use((config) => {
  config.sessionRevision = revision
  config.sessionToken = activeToken
  return config
}, undefined, { synchronous: true })

function current(config) {
  return config?.sessionRevision === revision
}

api.interceptors.response.use(
  (response) => {
    if (!current(response.config)) throw new axios.CanceledError('Сессия изменилась')
    return response
  },
  (error) => {
    if (error.config && !current(error.config)) return Promise.reject(new axios.CanceledError('Сессия изменилась'))
    if (error.response?.status === 401 && error.config?.sessionToken) {
      void expired(error.config.sessionToken)
    }
    return Promise.reject(error)
  }
)

export function errorMessage(error) {
  if (axios.isCancel(error)) return 'Сессия изменилась. Откройте данные заново.'
  const data = error?.response?.data
  if (error?.response?.status === 409) return 'Данные изменились. Обновите их и проверьте свои правки перед сохранением.'
  if (error?.response?.status === 401) return 'Сессия истекла. Войдите снова; запрос не был повторён.'
  if (error?.response?.status === 403) return 'Действие недоступно для вашей учётной записи.'
  const status = Number(error?.response?.status)
  if (status === 400 && data?.error?.message === 'Invalid request') return 'Проверьте заполненные поля и выбранные значения.'
  if (status === 404) return 'Объект не найден или больше недоступен.'
  if (status === 503 || status === 504) return 'Сервис временно недоступен. Повторите попытку позже.'
  if (typeof data?.error?.message === 'string') return data.error.message
  if (typeof data?.message === 'string') return data.message
  return error?.response ? 'Не удалось выполнить действие. Проверьте данные и повторите.' : 'Нет связи с сервером. Проверьте соединение и повторите.'
}

export function staticUrl(path) {
  if (!path) return ''
  if (/^https?:\/\//i.test(path)) return path
  return (import.meta.env.VITE_API_URL?.trim() || '') + path
}

export default api
