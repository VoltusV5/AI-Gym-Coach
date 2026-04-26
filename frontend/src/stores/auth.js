import { defineStore } from 'pinia'
import api, { setAuthHeader } from '@/api/api'
import { useWorkoutPlanStore } from '@/stores/workoutPlan'
import { useWorkoutSessionStore } from '@/stores/workoutSession'
import { saveToken, getToken, removeToken } from '@/utils/auth'
import { isTrainingDaysComplete } from '@/utils/trainingDays'

// Профиль свежесозданного гостя в БД ещё пуст → 404/«no rows».
// Это не ошибка — даём онбордингу заполнить его POST'ами.
function profileNotYetCreated(err) {
  const status = Number(err?.response?.status)
  const data = err?.response?.data
  const text = String(typeof data === 'string' ? data : data?.message ?? data?.error ?? '').toLowerCase()
  return status === 404 || (status === 500 && (text.includes('no rows') || text.includes('not found')))
}

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
            if (Number(err?.response?.status) === 401) {
              await removeToken()
              this.token = null
              setAuthHeader(null)
              await this.guestLogin()
            } else if (profileNotYetCreated(err)) {
              this.profile = this.profile ?? {}
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

    /** Гостевой вход — POST /api/v1/auth/guest */
    async guestLogin() {
      try {
        const res = await api.post('/api/v1/auth/guest', {})
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
        try {
          await this.fetchProfile()
        } catch (err) {
          if (profileNotYetCreated(err)) {
            this.profile = this.profile ?? {}
          } else {
            throw err
          }
        }
        return token
      } catch (err) {
        console.error('Guest login error:', err)
        throw err
      }
    },

    /** Получение профиля — GET /api/v1/profile */
    async fetchProfile() {
      const { data } = await api.get('/api/v1/profile')
      this.profile = data
      return data
    },

    /**
     * Обновление профиля — POST /api/v1/profile.
     * (PATCH на бэкенде не реализован; ProfileHandler читает JSON-body на POST.)
     *
     * Бэкенд использует optimistic locking: ждёт `version` в теле и проверяет
     * `WHERE user_id=? AND version=?`. На 409 один раз перезагружаем профиль и
     * пробуем снова (флаг _retried предохраняет от бесконечного цикла).
     */
    async updateProfile(fields, _retried = false) {
      try {
        if (this.profile?.version == null) {
          await this.fetchProfile()
        }
        const version = this.profile?.version
        if (typeof version !== 'number') {
          console.error('[updateProfile] profile.version is not a number:', this.profile)
          throw new Error('Не удалось получить version профиля. Проверь GET /api/v1/profile.')
        }
        const body = { ...fields, version }
        const { data } = await api.post('/api/v1/profile', body)
        this.profile = data
        return data
      } catch (err) {
        const status = Number(err?.response?.status)
        if (status === 401) {
          await this.initialize()
          return this.updateProfile(fields, _retried)
        }
        if (status === 409 && !_retried) {
          await this.fetchProfile()
          return this.updateProfile(fields, true)
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
