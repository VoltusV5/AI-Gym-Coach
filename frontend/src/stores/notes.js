import { defineStore } from 'pinia'
import api from '@/api/api'
import { apiStatus, errorText } from '@/api/v2'
import { useAuthStore } from '@/stores/auth'
import { appendPage } from '@/utils/pages'

const STORAGE_KEY = 'sportik_notes'
const PAGE_SIZE = 20

/** Return the current owner's cache key. Anonymous identities still have a user ID. */
function notesKey(owner) {
  return `${STORAGE_KEY}:${owner}`
}

function nowIso() {
  return new Date().toISOString()
}

function normalizeTitle(rawTitle, rawBody) {
  const t = String(rawTitle ?? '').trim()
  if (t) return t
  const body = String(rawBody ?? '').trim()
  if (!body) return 'Без названия'
  return body.slice(0, 32)
}

function toLocalNote(dto) {
  return {
    id: String(dto.id),
    title: String(dto.title ?? ''),
    body: String(dto.body ?? ''),
    createdAt: String(dto.created_at ?? dto.createdAt ?? nowIso()),
    updatedAt: String(dto.updated_at ?? dto.updatedAt ?? nowIso())
  }
}

function toApiPayload(title, body) {
  return {
    title: String(title ?? ''),
    body: String(body ?? '')
  }
}

function isServerNoteId(id) {
  return /^\d+$/.test(String(id ?? '').trim())
}

export const useNotesStore = defineStore('notes', {
  state: () => ({
    notes: [],
    owner: null,
    offset: 0,
    total: 0,
    loading: false,
    error: '',
    status: 0,
    refreshFailed: false,
  }),

  getters: {
    sortedNotes(state) {
      return [...state.notes].sort(
        (a, b) => new Date(b.updatedAt).getTime() - new Date(a.updatedAt).getTime()
      )
    },

    noteById: (state) => (id) => state.notes.find((n) => n.id === id) ?? null
  },

  actions: {
    /** Load only the cache belonging to the current account. */
    hydrate() {
      const owner = useAuthStore().accountId
      if (this.owner === owner) return
      this.$reset()
      this.owner = owner
      if (!owner) return
      try {
        const raw = localStorage.getItem(notesKey(owner))
        if (!raw) return
        const parsed = JSON.parse(raw)
        if (!Array.isArray(parsed)) return
        this.notes = parsed
      } catch {
        this.notes = []
      }
    },

    /** Persist the current account's notes, including local drafts. */
    persist() {
      if (!this.owner || this.owner !== useAuthStore().accountId) return
      try {
        localStorage.setItem(notesKey(this.owner), JSON.stringify(this.notes))
      } catch {
        // Local cache is optional when browser storage is unavailable.
      }
    },

    createEmptyDraft() {
      const ts = Date.now()
      const id = `note-${ts}`
      const iso = nowIso()
      const note = {
        id,
        title: '',
        body: '',
        createdAt: iso,
        updatedAt: iso
      }
      this.notes.push(note)
      this.persist()
      return note
    },

    /** Fetch the next v1 offset page or refresh it from the beginning. */
    async fetchNotesFromApi(reset = false) {
      const auth = useAuthStore()
      this.hydrate()
      if (this.loading || !auth.token) return
      const owner = this.owner
      this.loading = true
      this.error = ''
      this.status = 0
      this.refreshFailed = reset
      try {
        const offset = reset ? 0 : this.offset
        const { data } = await api.get('/api/v1/notes', { params: { offset, limit: PAGE_SIZE } })
        if (this.owner !== owner || auth.accountId !== owner) return
        const items = (data.items ?? []).map(toLocalNote)
        const previous = reset ? this.notes.filter((note) => !isServerNoteId(note.id)) : this.notes
        this.notes = appendPage(previous, items)
        this.total = data.total
        this.offset = items.length ? offset + items.length : data.total
        this.persist()
      } catch (err) {
        if (this.owner === owner) {
          this.status = apiStatus(err)
          this.error = errorText(err)
        }
      } finally {
        if (this.owner === owner) this.loading = false
      }
    },

    /** Retry the failed page, including a first-page refresh over an existing cache. */
    retryNotes() {
      return this.fetchNotesFromApi(this.refreshFailed)
    },

    saveNote(id, title, body) {
      const note = this.notes.find((n) => n.id === id)
      if (!note) return null
      note.title = normalizeTitle(title, body)
      note.body = String(body ?? '')
      note.updatedAt = nowIso()
      this.persist()
      return note
    },

    upsertNote(id, title, body) {
      const existing = this.notes.find((n) => n.id === id)
      if (existing) return this.saveNote(id, title, body)

      const iso = nowIso()
      const note = {
        id,
        title: normalizeTitle(title, body),
        body: String(body ?? ''),
        createdAt: iso,
        updatedAt: iso
      }
      this.notes.push(note)
      this.persist()
      return note
    },

    async createNoteRemote(id, title, body) {
      const owner = this.owner
      const payload = toApiPayload(title, body)
      const { data } = await api.post('/api/v1/notes', payload)
      const saved = toLocalNote(data || {})
      if (useAuthStore().accountId !== owner) return saved
      const idx = this.notes.findIndex((n) => n.id === id)
      if (idx >= 0) this.notes[idx] = saved
      else this.notes.push(saved)
      this.persist()
      return saved
    },

    async updateNoteRemote(id, title, body) {
      const owner = this.owner
      const payload = toApiPayload(title, body)
      const { data } = await api.patch(`/api/v1/notes/${id}`, payload)
      const saved = toLocalNote(data || {})
      if (useAuthStore().accountId !== owner) return saved
      const idx = this.notes.findIndex((n) => n.id === id)
      if (idx >= 0) this.notes[idx] = saved
      else this.notes.push(saved)
      this.persist()
      return saved
    },

    async saveNoteWithApi(id, title, body) {
      const exists = Boolean(this.noteById(id)) && isServerNoteId(id)
      this.upsertNote(id, title, body)
      if (exists) return this.updateNoteRemote(id, title, body)
      return this.createNoteRemote(id, title, body)
    },

    deleteNote(id) {
      const before = this.notes.length
      this.notes = this.notes.filter((n) => n.id !== id)
      if (this.notes.length !== before) this.persist()
    },

    async deleteNoteWithApi(id) {
      this.deleteNote(id)
      if (!isServerNoteId(id)) return
      try {
        await api.delete(`/api/v1/notes/${id}`)
      } catch {
        // Local cache is optional when browser storage is unavailable.
      }
    }
  }
})
