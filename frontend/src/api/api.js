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

  // Production без VITE_API_URL: запросы на тот же origin (nginx в Docker проксирует /api → backend).
  // Прямой localhost:5050 ломает сценарии с другого устройства и обход прокси.
  console.log('[API] Production → same-origin (empty baseURL)')
  return ''
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
