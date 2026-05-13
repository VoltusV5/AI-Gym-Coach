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
  console.log('[API] Production → same-origin (empty baseURL)')
  return ''
}

const api = axios.create({
  baseURL: resolveApiBaseURL()
})


export function setAuthHeader(token) {
  if (token) {
    api.defaults.headers.common['Authorization'] = `Bearer ${token}`
  } else {
    delete api.defaults.headers.common['Authorization']
  }
}


export function staticUrl(path) {
  if (!path) return ''
  const base = import.meta.env.VITE_API_URL?.trim() || ''
  return base + path
}

export default api