import { defineStore } from 'pinia'
import { useAuthStore } from '@/stores/auth'

const STORAGE_KEY = 'sportik_workout_plan'

/** Scope browser caches to the active authenticated or anonymous account. */
function storageKey() {
  const id = useAuthStore().accountId
  return id ? `${STORAGE_KEY}:${id}` : null
}



export const useWorkoutPlanStore = defineStore('workoutPlan', {
  state: () => ({
    plan: null
  }),

  getters: {

    flatExercises(state) {
      if (!state.plan?.plan?.length) return []
      const dayA = state.plan.plan.find(
        (d) => String(d?.day ?? '').toUpperCase() === 'A'
      )
      if (!dayA) return []
      const out = []
      const blocks = dayA.exercises || []
      for (let j = 0; j < blocks.length; j++) {
        const vars = blocks[j]
        if (!Array.isArray(vars) || !vars.length) continue
        const ex = vars[0]
        out.push({
          id: ex.id,
          exercise_name: ex.exercise_name,
          weight: ex.weight,
          day: dayA.day,
          day_name: dayA.day_name,
          slot_index: j
        })
      }
      return out
    },

    exerciseCount() {
      return this.flatExercises.length
    },


    estimatedMinutes() {
      const n = this.exerciseCount
      if (n === 0) return 0
      return Math.max(20, Math.round(n * 6))
    },

    splitLabel(state) {
      return state.plan?.split || ''
    }
  },

  actions: {
    setPlanFromApi(data) {
      this.plan = data
      try {
        const key = storageKey()
        if (key) localStorage.setItem(key, JSON.stringify(data))
      } catch {
        // Local cache is optional when browser storage is unavailable.
      }
    },

    hydrateFromStorage() {
      try {
        const key = storageKey()
        const raw = key && localStorage.getItem(key)
        if (raw) this.plan = JSON.parse(raw)
      } catch {
        this.plan = null
      }
    },

    clearPlan() {
      this.plan = null
      try {
        const key = storageKey()
        if (key) localStorage.removeItem(key)
      } catch {
        // Local cache is optional when browser storage is unavailable.
      }
    }
  }
})