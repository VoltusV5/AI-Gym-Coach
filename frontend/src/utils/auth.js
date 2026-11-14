import { Capacitor } from '@capacitor/core'
import { SecureStorage } from '@aparajita/capacitor-secure-storage'

const TOKEN_KEY = 'auth_token'
let writes = Promise.resolve()

function enqueue(write) {
  const result = writes.then(write)
  writes = result.catch(() => undefined)
  return result
}

export function saveToken(token) {
  return enqueue(async () => {
    if (Capacitor.isNativePlatform()) {
      await SecureStorage.set(TOKEN_KEY, token)
      localStorage.removeItem(TOKEN_KEY)
    } else {
      localStorage.setItem(TOKEN_KEY, token)
    }
  })
}

export async function getToken() {
  await writes
  if (!Capacitor.isNativePlatform()) return localStorage.getItem(TOKEN_KEY)
  const value = await SecureStorage.get(TOKEN_KEY)
  return typeof value === 'string' ? value : null
}

export function removeToken() {
  return enqueue(async () => {
    localStorage.removeItem(TOKEN_KEY)
    if (Capacitor.isNativePlatform()) await SecureStorage.remove(TOKEN_KEY)
  })
}
