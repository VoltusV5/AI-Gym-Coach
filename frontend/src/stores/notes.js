import { defineStore } from 'pinia'
import api, { errorMessage } from '@/api/api'

const STORAGE_KEY = 'sportik_notes'

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
    nextCursor: null,
    loading: false,
    error: null
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
    hydrate() {
      try {
        const raw = localStorage.getItem(STORAGE_KEY)
        if (!raw) return
        const parsed = JSON.parse(raw)
        if (!Array.isArray(parsed)) return
        this.notes = parsed
      } catch {
        this.notes = []
      }
    },

    persist() {
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(this.notes))
      } catch {

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

    async fetchNotesFromApi(append = false) {
      if (this.loading || (append && !this.nextCursor)) return this.notes
      this.loading = true
      this.error = null
      try {
        const { data } = await api.get('/api/v1/notes', { params: { limit: 20, cursor: append ? this.nextCursor : undefined } })
        const incoming = (data?.items || []).map(toLocalNote)
        const drafts = this.notes.filter((note) => !isServerNoteId(note.id))
        this.notes = append ? [...this.notes, ...incoming.filter((note) => !this.noteById(note.id))] : [...drafts, ...incoming]
        this.nextCursor = data?.next_cursor ?? null
        this.persist()
        return this.notes
      } catch (error) {
        this.error = errorMessage(error)
        throw error
      } finally {
        this.loading = false
      }
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
      const payload = toApiPayload(title, body)
      const { data } = await api.post('/api/v1/notes', payload)
      const saved = toLocalNote(data || {})
      const idx = this.notes.findIndex((n) => n.id === id)
      if (idx >= 0) this.notes[idx] = saved
      else this.notes.push(saved)
      this.persist()
      return saved
    },

    async updateNoteRemote(id, title, body) {
      const payload = toApiPayload(title, body)
      const { data } = await api.patch(`/api/v1/notes/${id}`, payload)
      const saved = toLocalNote(data || {})
      const idx = this.notes.findIndex((n) => n.id === id)
      if (idx >= 0) this.notes[idx] = saved
      else this.notes.push(saved)
      this.persist()
      return saved
    },

    async saveNoteWithApi(id, title, body) {
      this.error = null
      try {
        if (isServerNoteId(id)) return await this.updateNoteRemote(id, title, body)
        return await this.createNoteRemote(id, title, body)
      } catch (error) {
        this.error = errorMessage(error)
        throw error
      }
    },

    deleteNote(id) {
      const before = this.notes.length
      this.notes = this.notes.filter((n) => n.id !== id)
      if (this.notes.length !== before) this.persist()
    },

    async deleteNoteWithApi(id) {
      this.error = null
      try {
        if (isServerNoteId(id)) await api.delete(`/api/v1/notes/${id}`)
        this.deleteNote(id)
      } catch (error) {
        this.error = errorMessage(error)
        throw error
      }
    },

    async syncWithApi() {
      try { await this.fetchNotesFromApi() } catch { return false }
      return true
    }
  }
})
