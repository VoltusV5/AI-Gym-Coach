import { defineStore } from 'pinia'
import { CanceledError } from 'axios'
import api, { setAuthHeader } from '@/api/api'
import { useNutritionStore } from '@/stores/nutrition'
import { clearDrafts } from '@/utils/reauthDrafts'
import { useWorkoutPlanStore } from '@/stores/workoutPlan'
import { useWorkoutSessionStore } from '@/stores/workoutSession'
import { saveToken, getToken, removeToken } from '@/utils/auth'
import { isTrainingDaysComplete } from '@/utils/trainingDays'
import { sessionClaims } from '@/utils/session'

const initializations = new WeakMap()
const authentications = new WeakMap()

/** Only the newest explicit authentication request may change the account. */
function requireLatestAuth(state, attempt) {
  if (authentications.get(state) !== attempt) throw new CanceledError('Authentication superseded')
}

/** Recognize a v1 account whose profile is still being created. */
function profileNotYetCreated(err) {
  const status = Number(err?.response?.status)
  const data = err?.response?.data
  const text = String(typeof data === 'string' ? data : data?.message ?? data?.error ?? '').toLowerCase()
  return status === 404 || (status === 500 && (text.includes('no rows') || text.includes('not found')))
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: null,
    lastSubject: null,
    accountEpoch: 0,
    pendingEquipment: false,
    profile: null,
    isLoading: false,
    error: null
  }),

  getters: {
    /** Account identity hint for scoping local data. */
    accountId: (state) => sessionClaims(state.token)?.sub ?? null,
    /** Guest credentials also support private training endpoints. */
    isGuest: (state) => sessionClaims(state.token)?.is_anonymous === true,

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

      return state.pendingEquipment ? 'Equipment' : 'Home'
    }
  },

  actions: {

    /** Initialize once per concurrent group of callers. Store a safe UI error on failure. */
    initialize() {
      const state = this.$state
      const current = initializations.get(state)
      if (current) return current
      const run = this.restoreSession().finally(() => initializations.delete(state))
      initializations.set(state, run)
      return run
    },

    /** Restore the local session or start an anonymous session for onboarding. */
    async restoreSession() {
      this.isLoading = true
      this.error = null
      try {
        const token = await getToken()
        if (token) {
          this.lastSubject = sessionClaims(token)?.sub ?? null
          this.token = token
          setAuthHeader(token)
          this.restoreEquipmentStep()
          await this.fetchProfile()
        } else {
          await this.guestLogin()
        }
      } catch (err) {
        if (profileNotYetCreated(err)) {
          this.profile = {}
        } else if (err?.response?.status === 401) {
          this.profile = null
        } else {
          this.error = 'Не удалось загрузить профиль. Проверьте соединение и повторите попытку.'
        }
      } finally {
        this.isLoading = false
      }
    },

    /**
     * Persist a new session and clear cached training data when the account changes.
     * @param {string} token
     */
    async acceptToken(token) {
      if (!token) throw new Error('Сервер не вернул ключ сессии')
      const previous = this.lastSubject
      const next = sessionClaims(token)?.sub
      if (previous !== (next ?? null)) {
        this.accountEpoch++
        clearDrafts()
        useNutritionStore().$reset()
        useWorkoutPlanStore().$reset()
        useWorkoutSessionStore().$reset()
        this.profile = null
      }
      this.error = null
      this.lastSubject = next ?? null
      this.token = token
      setAuthHeader(token)
      this.restoreEquipmentStep()
      await saveToken(token)
    },

    /** Restore whether this account entered the new equipment step. */
    restoreEquipmentStep() {
      const owner = this.accountId
      try {
        this.pendingEquipment = Boolean(owner && localStorage.getItem(`sportik-equipment-pending:${owner}`))
      } catch {
        this.pendingEquipment = false
      }
    },

    /** Remember onboarding progress for the current account. */
    setEquipmentStep(pending) {
      const owner = this.accountId
      this.pendingEquipment = pending
      if (!owner) return
      try {
        const key = `sportik-equipment-pending:${owner}`
        if (pending) localStorage.setItem(key, '1')
        else localStorage.removeItem(key)
      } catch {
        // The current tab can continue when local storage is blocked.
      }
    },

    /**
     * Invalidate only the credential used by the failed request.
     * @param {string} authorization
     */
    async expire(authorization) {
      const expired = this.token
      if (!expired || authorization !== `Bearer ${expired}`) return
      this.token = null
      this.profile = null
      setAuthHeader(null)
      useNutritionStore().$reset()
      await removeToken(expired)
    },

    /** Authenticate once and discard the result if another sign-in has since started. */
    async authenticate(path, body) {
      const state = this.$state
      const attempt = {}
      authentications.set(state, attempt)
      const { data, headers } = await api.post(`/api/v1/auth/${path}`, body)
      requireLatestAuth(state, attempt)
      const token = data?.token ?? headers.authorization?.match(/^Bearer (.+)$/)?.[1]
      await this.acceptToken(token)
      requireLatestAuth(state, attempt)
      await this.fetchProfile()
      requireLatestAuth(state, attempt)
      return data
    },

    /** Create and persist an anonymous session, then load the v1 profile. */
    async guestLogin() {
      await this.authenticate('guest', {})
      return this.token
    },

    /** Load the authenticated v1 profile. Rejects API errors except an absent new profile. */
    async fetchProfile() {
      const token = this.token
      try {
        const { data } = await api.get('/api/v1/profile')
        if (this.token === token) this.profile = data
        return data
      } catch (err) {
        if (!profileNotYetCreated(err)) throw err
        if (this.token === token) this.profile = {}
        return {}
      }
    },

    /** Save an onboarding step against the current version. Rejects conflicts and expiry. */
    async updateProfile(fields) {
      const token = this.token
      try {
        const { data } = await api.post('/api/v1/profile', {
          ...fields,
          version: this.profile?.version ?? 0,
        })
        if (this.token === token) this.profile = data
        return data
      } catch (err) {
        if (err?.response?.status === 409) await this.fetchProfile().catch(() => {})
        throw err
      }
    },

    /** Request generation via the legacy onboarding contract. */
    async generatePlan() {
      const { data } = await api.post('/api/v1/plans/generate', {}, { timeout: 55000 })
      return data
    },

    /** Register the current guest or create an account, then reload its profile. */
    register({ email, password }) {
      return this.authenticate('register', { email, password })
    },

    /** Sign in and load the authenticated account profile. */
    login({ email, password }) {
      return this.authenticate('login', { email, password })
    },

    /** Change the password using the authenticated v1 endpoint. */
    async changePassword({ currentPassword, newPassword }) {
      const payload = {
        current_password: currentPassword,
        new_password: newPassword
      }
      const { data } = await api.post('/api/v1/auth/change-password', payload)
      return data
    }
  }
})
