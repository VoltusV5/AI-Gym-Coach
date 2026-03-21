import axios from 'axios'

const api = axios.create({
  baseURL: 'http://localhost:9091', // бэкенд
  timeout: 10000,
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
