import { beforeEach, afterEach, describe, it, expect } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import MockAdapter from 'axios-mock-adapter'
import api from '@/api/api'
import { useNotesStore } from '@/stores/notes'
let mock
beforeEach(() => { setActivePinia(createPinia()); localStorage.clear(); mock = new MockAdapter(api) })
afterEach(() => mock.restore())
describe('notes persistence', () => {
  it('preserves the draft and rejects a failed save instead of showing success', async () => {
    const store = useNotesStore()
    const draft = store.createEmptyDraft()
    mock.onPost('/api/v1/notes').reply(503)
    await expect(store.saveNoteWithApi(draft.id, 'Title', 'Body')).rejects.toBeDefined()
    expect(store.noteById(draft.id)).toBeDefined()
    expect(store.error).toBeTruthy()
  })
  it('follows the server cursor and stops at the last page', async () => {
    const store = useNotesStore()
    mock.onGet('/api/v1/notes').reply((config) => config.params.cursor
      ? [200, { items: [{ id: 2, title: 'second' }], next_cursor: null }]
      : [200, { items: [{ id: 1, title: 'first' }], next_cursor: 'opaque-cursor' }])
    await store.fetchNotesFromApi()
    await store.fetchNotesFromApi(true)
    await store.fetchNotesFromApi(true)
    expect(store.notes.map((note) => note.id)).toEqual(['1', '2'])
    expect(mock.history.get).toHaveLength(2)
  })
  it('keeps a server note when deletion fails', async () => {
    const store = useNotesStore()
    store.upsertNote('42', 'Title', 'Body')
    mock.onDelete('/api/v1/notes/42').reply(500)
    await expect(store.deleteNoteWithApi('42')).rejects.toBeDefined()
    expect(store.noteById('42')).toBeDefined()
  })
})
