import { describe, it, expect, vi, beforeEach } from 'vitest'
vi.mock('@capacitor/core', () => ({ Capacitor: { isNativePlatform: vi.fn(() => true) } }))
vi.mock('@aparajita/capacitor-secure-storage', () => ({ SecureStorage: { get: vi.fn(), set: vi.fn(), remove: vi.fn() } }))
import { SecureStorage } from '@aparajita/capacitor-secure-storage'
import { Capacitor } from '@capacitor/core'
import { saveToken, getToken, removeToken } from '@/utils/auth'

beforeEach(() => { localStorage.clear(); vi.clearAllMocks(); Capacitor.isNativePlatform.mockReturnValue(true) })
describe('token storage', () => {
  it('uses the installed native API signature', async () => {
    SecureStorage.get.mockResolvedValue('credential')
    await saveToken('credential')
    expect(SecureStorage.set).toHaveBeenCalledWith('auth_token', 'credential')
    expect(await getToken()).toBe('credential')
    expect(SecureStorage.get).toHaveBeenCalledWith('auth_token')
    await removeToken()
    expect(SecureStorage.remove).toHaveBeenCalledWith('auth_token')
  })
  it('uses browser storage without native calls', async () => {
    Capacitor.isNativePlatform.mockReturnValue(false)
    await saveToken('browser-credential')
    expect(await getToken()).toBe('browser-credential')
    expect(SecureStorage.set).not.toHaveBeenCalled()
    await removeToken()
    expect(await getToken()).toBeNull()
  })
})
