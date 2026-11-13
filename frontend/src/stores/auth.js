import { defineStore } from 'pinia'
import api, { setAuthHeader, errorMessage } from '@/api/api'
import { useWorkoutPlanStore } from '@/stores/workoutPlan'
import { useWorkoutSessionStore } from '@/stores/workoutSession'
import { saveToken, getToken, removeToken } from '@/utils/auth'
import { useNotesStore } from '@/stores/notes'
import { useNutritionStore } from '@/stores/nutrition'
import { isTrainingDaysComplete } from '@/utils/trainingDays'
function profileNotYetCreated(err) {
  const status = Number(err?.response?.status)
  const data = err?.response?.data
  const text = String(typeof data === 'string' ? data : data?.message ?? data?.error ?? '').toLowerCase()
  return status === 404 || (status === 500 && (text.includes('no rows') || text.includes('not found')))
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: null,
    identity: null,
    profileV2: null,
    sessionAttempt: 0,
    sessionEpoch: 0,
    initialized: false,
    profile: null,
    isLoading: false,
    error: null
  }),

  getters: {
    isRegistered: (state) => Boolean(state.identity && !state.identity.is_anonymous),
    roles: (state) => state.identity?.roles || [],
    nextOnboardingStep: (state) => {
      if (!state.profile) return 'Welcome'
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

    clearAccountData() {
      this.sessionEpoch += 1
      this.profile = null
      this.profileV2 = null
      this.identity = null
      useWorkoutPlanStore().clearPlan()
      useWorkoutSessionStore().clear()
      useNotesStore().$reset()
      useNutritionStore().$reset()
      for (const key of ['sportik_notes', 'sportik_workout_plan', 'sportik_workout_session']) localStorage.removeItem(key)
    },

    async fetchIdentity() {
      const { data } = await api.get('/api/v2/me/session')
      const previous = localStorage.getItem('sportik_account_id')
      if (previous && previous !== String(data.id)) this.clearAccountData()
      localStorage.setItem('sportik_account_id', String(data.id))
      this.identity = data
      return data
    },

    async expire(expected) {
      if (this.token !== expected) return
      const stored = await getToken()
      if (this.token !== expected) return
      if (stored && stored !== expected) {
        this.token = stored
        setAuthHeader(stored)
        this.clearAccountData()
        await this.fetchIdentity()
        return
      }
      this.sessionAttempt += 1
      this.token = null
      setAuthHeader(null)
      this.clearAccountData()
      this.error = 'Сессия истекла. Войдите снова.'
      await removeToken()
    },

    async logout() {
      this.sessionAttempt += 1
      this.token = null
      setAuthHeader(null)
      this.clearAccountData()
      await removeToken()
    },

    async fetchProfileV2() {
      const { data } = await api.get('/api/v2/me/profile')
      this.profileV2 = data
      return data
    },

    async saveProfileV2(fields) {
      const { data } = await api.patch('/api/v2/me/profile', { ...fields, version: this.profileV2?.version })
      this.profileV2 = data
      return data
    },

    async initialize() {
      this.isLoading = true
      this.error = null

      try {
        const token = await getToken()

        if (!token) {
          await this.guestLogin()
        } else {
          this.token = token
          setAuthHeader(token)

          try {
            await this.fetchIdentity()
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
        const msg = errorMessage(err)
        const noResponse = err?.code === 'ERR_NETWORK' || !err?.response
        this.error = noResponse
          ? 'Не удаётся связаться с API. Backend — порт 5050; в Docker пересоберите frontend после правок nginx (docker compose up -d --build frontend).'
          : msg || 'Connection error'
      } finally {
        this.isLoading = false
        this.initialized = true
      }
    },


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
        await this.fetchIdentity()
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


    async fetchProfile() {
      const { data } = await api.get('/api/v1/profile')
      this.profile = data
      return data
    },

    async updateProfile(fields) {
      const { data } = await api.post('/api/v1/profile', { ...fields, version: this.profile?.version ?? 0 })
      this.profile = data
      return data
    },

    async generatePlan() {
      try {
        const { data } = await api.post('/api/v1/plans/generate', {})
        return data
      } catch (err) {
        console.error('Plan generation error:', err)
        throw err
      }
    },

    async authenticate(kind, { email, password }) {
      const attempt = ++this.sessionAttempt
      const { data } = await api.post(`/api/v1/auth/${kind}`, { email, password })
      if (attempt !== this.sessionAttempt) throw new Error('Выбрана другая сессия')
      if (!data?.token) throw new Error('Сервер не вернул токен')
      this.clearAccountData()
      this.token = data.token
      setAuthHeader(data.token)
      await saveToken(data.token)
      if (attempt !== this.sessionAttempt) throw new Error('Выбрана другая сессия')
      await this.fetchIdentity()
      await this.fetchProfile()
      this.error = null
      return data
    },

    register(credentials) { return this.authenticate('register', credentials) },
    login(credentials) { return this.authenticate('login', credentials) },

    async changePassword({ currentPassword, newPassword }) {
      const payload = {
        current_password: currentPassword,
        new_password: newPassword
      }
      const { data } = await api.post('/api/v1/auth/change-password', payload)
      return data
    },


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