import { defineStore } from 'pinia'
import api, { setAuthHeader } from '@/api/api'
import { useWorkoutPlanStore } from '@/stores/workoutPlan'
import { useWorkoutSessionStore } from '@/stores/workoutSession'
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
          ? 'Не удаётся связаться с API. Backend — порт 5050; в Docker пересоберите frontend после правок nginx (docker compose up -d --build frontend).'
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
        let res
        try {
          // Сначала «короткий» путь — без лишнего 404 в консоли, если /api/v1 ещё не задеплоен
          res = await api.post('/auth/guest', {})
        } catch (_) {
          res = await api.post('/api/v1/auth/guest', {})
        }
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
        let data
        try {
          data = (await api.get('/profile')).data
        } catch (_) {
          data = (await api.get('/api/v1/profile')).data
        }
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
        let data
        try {
          data = (await api.post('/profile', fields)).data
        } catch (_) {
          data = (await api.patch('/api/v1/profile', fields)).data
        }
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

    async register({ email, password, name }) {
      const { data } = await api.post('/api/v1/auth/register', { email, password, name })
      const token = data?.token ?? null
      if (token) {
        this.token = token
        await saveToken(token)
        setAuthHeader(token)
      }
      this.profile = data?.user ?? null
      return data
    },

    async login({ email, password }) {
      const { data } = await api.post('/api/v1/auth/login', { email, password })
      const token = data?.token ?? null
      if (token) {
        this.token = token
        await saveToken(token)
        setAuthHeader(token)
      }
      this.profile = data?.user ?? null
      return data
    },

    async changePassword({ currentPassword, newPassword }) {
      const payload = {
        current_password: currentPassword,
        new_password: newPassword
      }
      const { data } = await api.post('/api/v1/auth/change-password', payload)
      return data
    },

    /**
     * Сброс для тестов: новый гостевой токен и чистый профиль в сторе.
     * Не полагается на reload — токен снимается и сразу выдаётся новый.
     */
    async restartSessionForTesting() {
      useWorkoutPlanStore().clearPlan()
      useWorkoutSessionStore().clear()
      await removeToken()
      this.token = null
      this.profile = null
      this.error = null
      setAuthHeader(null)
      await this.guestLogin()
    }
  }
})
