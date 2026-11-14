/**
 * Read non-authoritative session hints for client navigation. The API verifies the JWT.
 * @param {string|null} token
 * @returns {{sub?: string, is_anonymous?: boolean}|null}
 */
export function sessionClaims(token) {
  try {
    const body = token.split('.')[1].replaceAll('-', '+').replaceAll('_', '/')
    return JSON.parse(atob(body))
  } catch {
    return null
  }
}

/**
 * Restrict the sign-in return location to an internal path.
 * @param {unknown} value
 * @returns {string}
 */
export function safeReturnPath(value) {
  if (typeof value !== 'string' || !value.startsWith('/') || value.startsWith('//') ||
      /[\\\s]/.test(value) || [...value].some((char) => char.charCodeAt(0) < 32) ||
      /^\/auth(?:[/?#]|$)/i.test(value)) {
    return '/settings'
  }
  return value
}
