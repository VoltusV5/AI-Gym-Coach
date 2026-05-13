import { defineStore } from 'pinia'
import api from '@/api/api'

const STORAGE_KEY = 'sportik_notes'

function nowIso() {
  return new Date().toISOString()
}

function normalizeTitle(rawTitle, rawBody) {
  const t = String(rawTitle ?? '').trim()
  if (t) return t
  const body = String(rawBody ?? '').trim()
  if (!body) return '��� ��������'
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
    notes: []
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
      } catch (_) {
        this.notes = []
      }
    },

    persist() {
      try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(this.notes))
      } catch (_) {

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

    async fetchNotesFromApi() {
      const { data } = await api.get('/api/v1/notes')
      const items = Array.isArray(data?.items) ? data.items : []
      this.notes = items.map(toLocalNote)
      this.persist()
      return this.notes
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
      const exists = Boolean(this.noteById(id)) && isServerNoteId(id)
      const local = this.upsertNote(id, title, body)
      try {
        if (exists) return await this.updateNoteRemote(id, title, body)
        return await this.createNoteRemote(id, title, body)
      } catch (_) {
        return local
      }
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
      } catch (_) {

      }
    },

    async syncWithApi() {
      try {
        await this.fetchNotesFromApi()
      } catch (_) {

      }
    }
  }
})