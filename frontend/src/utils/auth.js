import { SecureStorage } from '@aparajita/capacitor-secure-storage'

const TOKEN_KEY = 'auth_token'

/** Serialize credential writes across web tabs when the browser supports Web Locks. */
function withTokenLock(operation) {
  const locks = globalThis.navigator?.locks
  return locks ? locks.request('sportik-auth-token', operation) : operation()
}

/** Store a credential using Capacitor Secure Storage. Rejects if the write fails. */
export function saveToken(token) {
  return withTokenLock(async () => {
    await SecureStorage.set(TOKEN_KEY, token)
    localStorage.removeItem(TOKEN_KEY)
  })
}

/** Read the current credential, with migration support for older web installs. */
export async function getToken() {
  const token = await SecureStorage.get(TOKEN_KEY).catch(() => null)
  if (typeof token === 'string') return token
  return localStorage.getItem(TOKEN_KEY)
}

/** Remove only the expired credential. Another tab may have already signed in. */
export function removeToken(expected) {
  return withTokenLock(async () => {
    if (localStorage.getItem(TOKEN_KEY) === expected) localStorage.removeItem(TOKEN_KEY)
    if (await SecureStorage.get(TOKEN_KEY) === expected) await SecureStorage.remove(TOKEN_KEY)
  })
}
