import axios from 'axios'

function resolveApiBaseURL() {
  const raw = import.meta.env.VITE_API_URL?.trim()

  if (raw) {
    console.log('[API] Using VITE_API_URL:', raw)
    return raw
  }

  if (import.meta.env.DEV) {
    console.log('[API] DEV mode → relative URL (uses Vite proxy)')
    return ''
  }

  // Production fallback — должен быть порт БЭКЕНДА, а не фронтенда!
  console.log('[API] Production fallback → http://localhost:5050')
  return 'http://localhost:5050'
}

const api = axios.create({
  baseURL: resolveApiBaseURL()
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
