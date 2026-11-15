import { Capacitor } from '@capacitor/core'
import { SecureStorage } from '@aparajita/capacitor-secure-storage'

const TOKEN_KEY = 'auth_token'
let writes = Promise.resolve()

function enqueue(write) {
  const operation = () => globalThis.navigator?.locks
    ? navigator.locks.request('sportik-auth-token', write) : write()
  const result = writes.then(operation)
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

async function storedToken() {
  if (!Capacitor.isNativePlatform() && localStorage.getItem(TOKEN_KEY)) return localStorage.getItem(TOKEN_KEY)
  const value = await SecureStorage.get(TOKEN_KEY).catch(() => null)
  return typeof value === 'string' ? value : null
}

export async function getToken() {
  await writes
  return storedToken()
}

export function removeToken(expected) {
  return enqueue(async () => {
    const current = await storedToken()
    if (expected !== undefined && current !== expected) return
    localStorage.removeItem(TOKEN_KEY)
    await SecureStorage.remove(TOKEN_KEY)
  })
}
