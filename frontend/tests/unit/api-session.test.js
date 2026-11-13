import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest'
import MockAdapter from 'axios-mock-adapter'
import api, { setAuthHeader, onSessionExpired } from '@/api/api'

let mock
beforeEach(() => { mock = new MockAdapter(api); setAuthHeader('first-account') })
afterEach(() => { mock.restore(); setAuthHeader(null); onSessionExpired(() => undefined) })
describe('session boundaries', () => {
  it('discards a late successful response from the previous account', async () => {
    let resolve
    mock.onGet('/private').reply(() => new Promise((done) => { resolve = done }))
    const request = api.get('/private')
    const rejected = expect(request).rejects.toMatchObject({ code: 'ERR_CANCELED' })
    await vi.waitFor(() => expect(resolve).toBeTypeOf('function'))
    setAuthHeader('second-account')
    resolve([200, { private: 'first-account' }])
    await rejected
  })
  it('does not expire the current account when an earlier request fails', async () => {
    let resolve
    const expire = vi.fn()
    onSessionExpired(expire)
    mock.onGet('/private').reply(() => new Promise((done) => { resolve = done }))
    const request = api.get('/private')
    const rejected = expect(request).rejects.toMatchObject({ code: 'ERR_CANCELED' })
    await vi.waitFor(() => expect(resolve).toBeTypeOf('function'))
    setAuthHeader('second-account')
    resolve([401, {}])
    await rejected
    expect(expire).not.toHaveBeenCalled()
  })
  it('reports expiry without replaying the failed write', async () => {
    const expire = vi.fn()
    onSessionExpired(expire)
    mock.onPost('/write').reply(401)
    await expect(api.post('/write', { version: 1 })).rejects.toMatchObject({ response: { status: 401 } })
    expect(mock.history.post).toHaveLength(1)
    expect(expire).toHaveBeenCalledWith('first-account')
  })
})

it('captures the credential and revision together before a synchronous account switch', async () => {
  mock.onGet('/private').reply(200, { private: 'first-account' })
  const request = api.get('/private')
  setAuthHeader('second-account')
  await expect(request).rejects.toMatchObject({ code: 'ERR_CANCELED' })
  expect(mock.history.get[0].sessionToken).toBe('first-account')
})
