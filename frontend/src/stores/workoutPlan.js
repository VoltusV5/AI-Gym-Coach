import { defineStore } from 'pinia'

const STORAGE_KEY = 'sportik_workout_plan'

/**
 * Ответ бэкенда после ML: { split, plan: [{ day, day_name, exercises: [[{ id, exercise_name, weight }]] }] }
 */
export const useWorkoutPlanStore = defineStore('workoutPlan', {
  state: () => ({
    plan: null
  }),

  getters: {
    /** Плоский список упражнений с подписью дня */
    flatExercises(state) {
      if (!state.plan?.plan?.length) return []
      const out = []
      for (const day of state.plan.plan) {
        const blocks = day.exercises || []
        for (const block of blocks) {
          for (const ex of block) {
            out.push({
              id: ex.id,
              exercise_name: ex.exercise_name,
              weight: ex.weight,
              day: day.day,
              day_name: day.day_name
            })
          }
        }
      }
      return out
    },

    exerciseCount() {
      return this.flatExercises.length
    },

    /** Заглушка длительности: ~6 мин на блок + база */
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
        sessionStorage.setItem(STORAGE_KEY, JSON.stringify(data))
      } catch (_) {
        /* ignore */
      }
    },

    hydrateFromStorage() {
      try {
        const raw = sessionStorage.getItem(STORAGE_KEY)
        if (raw) this.plan = JSON.parse(raw)
      } catch (_) {
        this.plan = null
      }
    },

    clearPlan() {
      this.plan = null
      try {
        sessionStorage.removeItem(STORAGE_KEY)
      } catch (_) {
        /* ignore */
      }
    }
  }
})
