import axios from 'axios'

function resolveApiBaseURL() {
  const raw = import.meta.env.VITE_API_URL
  if (raw != null && String(raw).trim() !== '') {
    return String(raw).trim()
  }
  // В dev без явного URL — относительные пути → прокси в vite.config.js (нет проблем с CORS).
  if (import.meta.env.DEV) {
    return ''
  }
  return 'http://localhost:8080'
}

const api = axios.create({
  baseURL: resolveApiBaseURL(),
  timeout: 120000 // генерация плана ждёт ML до ~52 с на бэкенде
})

/**
 * Установка токена в заголовок Authorization
 * @param {string|null} token
 */
export function setAuthHeader(token) {
  if (token) {
    api.defaults.headers.common['Authorization'] = `Bearer ${token}`
  } else {
    delete api.defaults.headers.common['Authorization']
  }
}

export default api
