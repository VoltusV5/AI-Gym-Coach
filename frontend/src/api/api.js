import axios, { CanceledError } from 'axios'

const baseURL = (import.meta.env?.VITE_API_URL ?? '').trim().replace(/\/$/, '')
const api = axios.create({ baseURL, timeout: 15000 })

/**
 * Check whether a request still belongs to the current authenticated session.
 * Anonymous responses and the authentication endpoints can complete without a bearer.
 */
export function currentRequest(config) {
  const credential = config?.headers?.get?.('Authorization') ?? config?.headers?.Authorization
  return !credential || config.url?.startsWith('/api/v1/auth/') ||
    credential === api.defaults.headers.common.Authorization
}

api.interceptors.response.use(
  (response) => {
    if (!currentRequest(response.config)) throw new CanceledError('Session changed')
    return response
  },
  (error) => {
    if (!currentRequest(error.config)) throw new CanceledError('Session changed')
    throw error
  },
)

/**
 * Set the default bearer credential for subsequent API calls.
 * @param {string|null} token
 */
export function setAuthHeader(token) {
  if (token) api.defaults.headers.common.Authorization = `Bearer ${token}`
  else delete api.defaults.headers.common.Authorization
}

/**
 * Resolve a backend asset path with the same origin as the API.
 * @param {string|null} path
 * @returns {string}
 */
export function staticUrl(path) {
  return path ? baseURL + path : ''
}

export default api
