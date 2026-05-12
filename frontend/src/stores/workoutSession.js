import { defineStore } from 'pinia'
import { getMockPlanGenerateResponse } from '@/mocks/workoutSession.mock'
import { workoutMocksEnabled } from '@/config/workoutMocks'
import api from '@/api/api'

const STORAGE_KEY = 'sportik_workout_session'

const ALT_DESC = ''

function emptySetsFromWeight(weight) {
  const w = weight != null && weight !== '' ? String(weight) : ''
  // Всегда начинаем с 3 полей для ввода
  return [
    { weightKg: w, reps: '' },
    { weightKg: '', reps: '' },
    { weightKg: '', reps: '' }
  ]
}

/** Подход заполнен: повторы > 0; вес — число ≥ 0 (0 кг допустим для упражнений с весом тела). */
function isSetFilled(set) {
  const w = String(set.weightKg ?? '').trim()
  const r = String(set.reps ?? '').trim()
  if (r === '' || Number(r) <= 0) return false
  if (w === '') return false
  const wn = Number(w)
  return Number.isFinite(wn) && wn >= 0
}

function findPlanDayA(planRoot) {
  const days = planRoot?.plan
  if (!Array.isArray(days) || days.length === 0) return null
  const a = days.find((d) => String(d?.day ?? '').toUpperCase() === 'A')
  return a ?? null
}

/**
 * Один слот из подмассива вариаций ТЗ.
 * @param {number} slotIndex
 * @param {Array<{ id: number, exercise_name: string, weight: number }>} variations
 */
function createSlot(slotIndex, variations, dayCode, dayName) {
  if (!variations?.length) return null
  const alternatives = variations.map((v) => ({
    id: Number(v.id),
    name: v.exercise_name,
    description: ALT_DESC,
    planWeight: v.weight
  }))
  const first = variations[0]
  return {
    slotIndex,
    dayCode,
    dayName,
    alternatives,
    selectedId: Number(first.id),
    sets: emptySetsFromWeight(first.weight)
  }
}

export const useWorkoutSessionStore = defineStore('workoutSession', {
  state: () => ({
    split: '',
    dayCode: 'A',
    dayName: '',
    slots: [],
    currentIndex: 0,
    source: 'none' // 'api' | 'mock' | 'none' | 'restored'
  }),

  getters: {
    slotCount: (s) => s.slots.length,

    currentSlot(state) {
      return state.slots[state.currentIndex] ?? null
    },

    selectedExercise(state) {
      const slot = state.slots[state.currentIndex]
      if (!slot) return null
      return (
        slot.alternatives.find((a) => Number(a.id) === Number(slot.selectedId)) ?? null
      )
    },

    homeRows(state) {
      return state.slots.map((slot) => {
        const ex =
          slot.alternatives.find((a) => Number(a.id) === Number(slot.selectedId)) ??
          slot.alternatives[0]
        const w = slot.sets[0]?.weightKg
        const weightNum =
          w !== '' && w != null && String(w).trim() !== '' ? Number(w) : null
        return {
          id: ex.id,
          exercise_name: ex.name,
          weight: Number.isFinite(weightNum) ? weightNum : null,
          day: slot.dayCode,
          day_name: slot.dayName
        }
      })
    },

    slotStatusList(state) {
      return state.slots.map((slot, i) => {
        const filledSetsCount = (slot.sets || []).filter(isSetFilled).length
        return {
          index: i,
          complete: filledSetsCount >= 3,
          isCurrent: i === state.currentIndex
        }
      })
    },

    /** Хотя бы один подход (вес + повторы) заполнен где угодно — можно слать complete (частичный JSON). */
    isReadyForComplete(state) {
      if (!state.slots.length) return false
      return state.slots.some((slot) => (slot.sets || []).some(isSetFilled))
    }
  },

  actions: {
    persist() {
      try {
        const payload = {
          split: this.split,
          dayCode: this.dayCode,
          dayName: this.dayName,
          slots: this.slots,
          currentIndex: this.currentIndex,
          source: this.source
        }
        localStorage.setItem(STORAGE_KEY, JSON.stringify(payload))
      } catch (_) {
        /* ignore */
      }
    },

    hydrate() {
      try {
        const raw = localStorage.getItem(STORAGE_KEY)
        if (!raw) return
        const data = JSON.parse(raw)
        if (data?.slots?.length) {
          this.split = data.split || ''
          this.dayCode = data.dayCode || 'A'
          this.dayName = data.dayName || ''
          this.slots = data.slots
          this.currentIndex = Math.min(
            data.currentIndex ?? 0,
            Math.max(0, data.slots.length - 1)
          )
          this.source = data.source || 'restored'
        }
      } catch (_) {
        /* ignore */
      }
    },

    clear() {
      this.split = ''
      this.dayCode = 'A'
      this.dayName = ''
      this.slots = []
      this.currentIndex = 0
      this.source = 'none'
      try {
        localStorage.removeItem(STORAGE_KEY)
      } catch (_) {
        /* ignore */
      }
    },

    /**
     * Построить сессию из ответа POST /api/v1/plans/generate (ТЗ: только день A).
     * @param {{ split?: string, plan?: unknown[] }} planRoot
     * @param {{ sourceOverride?: string }} [opts]
     */
    buildFromApiPlan(planRoot, opts = {}) {
      const dayA = findPlanDayA(planRoot)
      const slotArrays = dayA?.exercises

      if (!Array.isArray(slotArrays) || slotArrays.length === 0) {
        this.clear()
        return
      }

      const slots = []
      for (let j = 0; j < slotArrays.length; j++) {
        const variations = slotArrays[j]
        if (!Array.isArray(variations) || variations.length === 0) continue
        const slot = createSlot(
          j,
          variations,
          String(dayA.day ?? 'A'),
          String(dayA.day_name ?? '')
        )
        if (slot) slots.push(slot)
      }

      if (slots.length === 0) {
        this.clear()
        return
      }

      this.split = planRoot.split || ''
      this.dayCode = String(dayA.day ?? 'A')
      this.dayName = String(dayA.day_name ?? '')
      this.slots = slots
      this.currentIndex = 0
      this.source = opts.sourceOverride ?? 'api'
      this.persist()
    },

    syncWithPlanStore(workoutPlanStore) {
      const plan = workoutPlanStore.plan
      if (plan?.plan?.length) {
        if (this.source === 'api' && this.slots.length > 0) return
        if (this.source === 'mock' && this.slots.length > 0) return
        if (this.source === 'restored' && this.slots.length > 0) return
        this.buildFromApiPlan(plan)
        return
      }
      this.clear()
    },

    setCurrentIndex(i) {
      const n = this.slots.length
      if (n === 0) return
      this.currentIndex = Math.max(0, Math.min(i, n - 1))
      this.persist()
    },

    goPrev() {
      this.setCurrentIndex(this.currentIndex - 1)
    },

    goNext() {
      this.setCurrentIndex(this.currentIndex + 1)
    },

    updateSet(setIndex, field, value) {
      const slot = this.slots[this.currentIndex]
      if (!slot) return
      
      // Ensure the set object exists (should always for initial 3)
      if (!slot.sets[setIndex]) {
        slot.sets[setIndex] = { weightKg: '', reps: '' }
      }
      
      slot.sets[setIndex][field] = value
      
      // Если это последний подход в списке и в нём что-то начали писать — добавляем следующий
      if (setIndex === slot.sets.length - 1 && String(value).trim() !== '') {
        slot.sets.push({ weightKg: '', reps: '' })
      }

      this.persist()
    },

    selectAlternative(slotIndex, alternativeId) {
      const slot = this.slots[slotIndex]
      if (!slot) return
      const id = Number(alternativeId)
      const alt = slot.alternatives.find((a) => Number(a.id) === id)
      if (!alt) return
      if (Number(slot.selectedId) === id) return
      slot.selectedId = id
      slot.sets = emptySetsFromWeight(alt.planWeight)
      this.persist()
    },

    slotIsComplete(slot) {
      const filledSetsCount = (slot.sets || []).filter(isSetFilled).length
      return filledSetsCount >= 3
    },

    /**
     * Тело POST /api/v1/workouts/complete: только слоты с ≥1 заполненным подходом;
     * в каждом слоте — только заполненные подходы.
     */
    buildCompletePayload() {
      const slots = []
      for (let idx = 0; idx < this.slots.length; idx++) {
        const slot = this.slots[idx]
        const filledSets = (slot.sets || [])
          .filter(isSetFilled)
          .map((s) => ({
            weight_kg: Number(String(s.weightKg ?? '').trim()),
            reps: Number(String(s.reps ?? '').trim())
          }))
        if (filledSets.length === 0) continue
        slots.push({
          slot_index: slot.slotIndex ?? idx,
          exercise_id: Number(slot.selectedId),
          sets: filledSets
        })
      }
      return {
        day_code: this.dayCode || 'A',
        finished_at: new Date().toISOString(),
        slots
      }
    },

    async submitCompleteWorkout() {
      const payload = this.buildCompletePayload()
      if (!payload.slots.length) {
        throw new Error('Нет ни одного заполненного подхода')
      }
      const { data } = await api.post('/api/v1/workouts/complete', payload)
      this.clear()
      return data
    }
  }
})
