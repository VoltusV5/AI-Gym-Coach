import { defineStore } from 'pinia'
import api, { setAuthHeader } from '@/api/api'
import { saveToken, getToken, removeToken } from '@/utils/auth'

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

      if (state.profile.gender === null) return 'Gender'
      if (state.profile.age === null) return 'BirthYear'
      if (state.profile.height_cm === null || state.profile.weight_kg === null) return 'BodyMetrics'
      if (state.profile.activity_level === null) return 'ActivityType'
      if (state.profile.fitness_level === null) return 'FitnessLevel'
      if (state.profile.injuries_notes === null) return 'HealthRestrictions'
      if (state.profile.goal === null) return 'GoalSelection'
      if (state.profile.training_days_map === null) return 'TrainingDays'

      return 'Home' // Всё заполнено
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
        this.error = 'Connection error'
      } finally {
        this.isLoading = false
      }
    },

    /**
     * Гостевой вход (POST /auth/guest)
     */
    async guestLogin() {
      try {
        const { data } = await api.post('/auth/guest', {})
        const token = data.token
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
     * Обновление профиля (PATCH /profile)
     * @param {Object} fields
     */
    async updateProfile(fields) {
      try {
        const { data } = await api.patch('/profile', fields)
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
    }
  }
})
