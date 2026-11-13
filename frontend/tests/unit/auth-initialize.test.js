import { beforeEach, afterEach, it, expect } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import MockAdapter from 'axios-mock-adapter'
import api, { setAuthHeader } from '@/api/api'
import { useAuthStore } from '@/stores/auth'
import { saveToken } from '@/utils/auth'
let mock
beforeEach(() => { setActivePinia(createPinia()); localStorage.clear(); mock = new MockAdapter(api) })
afterEach(() => { mock.restore(); setAuthHeader(null) })
it('clears the previous identity before loading the current profile', async () => {
  localStorage.setItem('sportik_account_id', '101')
  await saveToken('second-account')
  mock.onGet('/api/v2/me/session').reply(200, { id: 102, roles: ['user'], is_anonymous: false })
  mock.onGet('/api/v1/profile').reply(200, { version: 3, age: 30 })
  const auth = useAuthStore()
  await auth.initialize()
  expect(auth.identity.id).toBe(102)
  expect(auth.profile).toEqual({ version: 3, age: 30 })
  expect(mock.history.get.map((request) => request.url)).toEqual(['/api/v2/me/session', '/api/v1/profile'])
})
