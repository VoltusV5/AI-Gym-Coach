import { defineStore } from 'pinia'
import api, { setAuthHeader } from '@/api/api'
import { useWorkoutPlanStore } from '@/stores/workoutPlan'
import { saveToken, getToken, removeToken } from '@/utils/auth'
import { isTrainingDaysComplete } from '@/utils/trainingDays'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: null,
    profile: null,
    isLoading: false,
    error: null
  }),

  getters: {
    // Определяем следующий шаг онбординга на основе данных профиля
    nextOnboardingStep: (state) => {
      if (!state.profile) return 'Welcome'

      // Порядок: рост/вес → пол → возраст → активность → травмы → цель → уровень → дни
      if (state.profile.height_cm == null || state.profile.weight_kg == null) return 'BodyMetrics'
      if (state.profile.gender == null || state.profile.gender === '') return 'Gender'
      if (state.profile.age == null || state.profile.age === 0) return 'Age'
      if (state.profile.activity_level == null || state.profile.activity_level === '') {
        return 'ActivityType'
      }
      if (state.profile.injuries_notes === null || state.profile.injuries_notes === undefined) {
        return 'HealthRestrictions'
      }
      if (state.profile.goal == null || state.profile.goal === '') return 'GoalSelection'
      if (state.profile.fitness_level == null || state.profile.fitness_level === '') {
        return 'FitnessLevel'
      }
      if (!isTrainingDaysComplete(state.profile.training_days_map)) return 'TrainingDays'

      return 'Home'
    }
  },

  actions: {
    /**
     * Инициализация приложения: проверка токена, получение профиля
     */
    async initialize() {
      this.isLoading = true
      this.error = null

      try {
        const token = await getToken()

        if (!token) {
          await this.guestLogin()
        } else {
          // Если токен есть - ставим заголовок и проверяем профиль
          this.token = token
          setAuthHeader(token)

          try {
            await this.fetchProfile()
          } catch (err) {
            // Если 401 ошибка - удаляем токен и пробуем войти ещё раз
            if (err.response && err.response.status === 401) {
              await removeToken()
              this.token = null
              setAuthHeader(null)
              await this.guestLogin()
            } else {
              throw err
            }
          }
        }
      } catch (err) {
        console.error('Auth initialization failed:', err)
        const msg = err?.message || ''
        const noResponse = err?.code === 'ERR_NETWORK' || !err?.response
        this.error = noResponse
          ? 'Не удаётся связаться с API. Запустите backend (go run) на порту 8080 и перезапустите npm run dev.'
          : msg || 'Connection error'
      } finally {
        this.isLoading = false
      }
    },

    /**
     * Гостевой вход (POST /auth/guest)
     */
    async guestLogin() {
      try {
        const res = await api.post('/auth/guest', {})
        const token =
          res.data?.token ||
          (typeof res.headers?.authorization === 'string' &&
          res.headers.authorization.startsWith('Bearer ')
            ? res.headers.authorization.slice(7)
            : null)
        if (!token) throw new Error('No token in guest response')
        this.token = token
        await saveToken(token)
        setAuthHeader(token)
        await this.fetchProfile()
        return token
      } catch (err) {
        console.error('Guest login error:', err)
        throw err
      }
    },

    /**
     * Получение профиля (GET /profile)
     */
    async fetchProfile() {
      try {
        const { data } = await api.get('/profile')
        this.profile = data
        return data
      } catch (err) {
        throw err
      }
    },

    /**
     * Обновление профиля (POST /profile)
     * @param {Object} fields
     */
    async updateProfile(fields) {
      try {
        const { data } = await api.post('/profile', fields)
        this.profile = data
        return data
      } catch (err) {
        // Если 401 - пробуем перелогиниться и повторить
        if (err.response && err.response.status === 401) {
          await this.initialize()
          return this.updateProfile(fields)
        }
        throw err
      }
    },

    /**
     * Запуск генерации плана (POST /api/v1/plans/generate)
     */
    async generatePlan() {
      try {
        const { data } = await api.post('/api/v1/plans/generate', {})
        return data
      } catch (err) {
        console.error('Plan generation error:', err)
        throw err
      }
    },

    /**
     * Сброс для тестов: новый гостевой токен и чистый профиль в сторе.
     * Не полагается на reload — токен снимается и сразу выдаётся новый.
     */
    async restartSessionForTesting() {
      useWorkoutPlanStore().clearPlan()
      await removeToken()
      this.token = null
      this.profile = null
      this.error = null
      setAuthHeader(null)
      await this.guestLogin()
    }
  }
})
