import { defineStore } from 'pinia'

const STORAGE_KEY = 'sportik_workout_plan'

/**
 * Ответ POST /api/v1/plans/generate (ТЗ):
 * { split, plan: [{ day, day_name, exercises: [[вариации...]] }] }
 * exercises[j] — слот; внутри вариации { id, exercise_name, weight }.
 */
export const useWorkoutPlanStore = defineStore('workoutPlan', {
  state: () => ({
    plan: null
  }),

  getters: {
    /** День A: по одному «главному» упражнению на слот (первая вариация). */
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
        localStorage.setItem(STORAGE_KEY, JSON.stringify(data))
      } catch (_) {
        /* ignore */
      }
    },

    hydrateFromStorage() {
      try {
        const raw = localStorage.getItem(STORAGE_KEY)
        if (raw) this.plan = JSON.parse(raw)
      } catch (_) {
        this.plan = null
      }
    },

    clearPlan() {
      this.plan = null
      try {
        localStorage.removeItem(STORAGE_KEY)
      } catch (_) {
        /* ignore */
      }
    }
  }
})
