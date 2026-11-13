import { vi } from 'vitest'
function storage() {
  const values = new Map()
  return {
    getItem: (key) => values.get(String(key)) ?? null,
    setItem: (key, value) => values.set(String(key), String(value)),
    removeItem: (key) => values.delete(String(key)),
    clear: () => values.clear()
  }
}
vi.stubGlobal('localStorage', storage())
vi.stubGlobal('sessionStorage', storage())
