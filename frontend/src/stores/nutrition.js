import { defineStore } from 'pinia'
import api from '@/api/api'

function toNum(v) {
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}


export function nutritionJournalDayISO(d = new Date()) {
  return new Intl.DateTimeFormat('sv-SE', { timeZone: 'Europe/Moscow' }).format(d).slice(0, 10)
}


function resolveJournalDay(day, storeDay) {
  const tryOne = (v) => {
    if (typeof v !== 'string') return ''
    const s = v.trim().slice(0, 10)
    return /^\d{4}-\d{2}-\d{2}$/.test(s) ? s : ''
  }
  return tryOne(day) || tryOne(storeDay) || nutritionJournalDayISO()
}

export const useNutritionStore = defineStore('nutrition', {
  state: () => ({
    day: nutritionJournalDayISO(),
    entries: [],
    favorites: [],
    goal: null,
    stats: null,
    dishes: [],
    myDishes: [],
    selectedDish: null,
    dashboard: null,
    reports: null,
    loading: false,
    entriesByMeal: {
      breakfast: [],
      lunch: [],
      dinner: [],
      snack: []
    }
  }),

  actions: {
    _rebuildEntriesByMeal() {
      const e = this.entries || []
      this.entriesByMeal = {
        breakfast: e.filter((x) => x.meal_type === 'breakfast'),
        lunch: e.filter((x) => x.meal_type === 'lunch'),
        dinner: e.filter((x) => x.meal_type === 'dinner'),
        snack: e.filter((x) => x.meal_type === 'snack')
      }
    },


    mergeServerEntry(row) {
      if (row == null || row.id == null) return
      const id = Number(row.id)
      const normalized = {
        ...row,
        id,
        grams: toNum(row.grams),
        protein_g: toNum(row.protein_g),
        fat_g: toNum(row.fat_g),
        carbs_g: toNum(row.carbs_g),
        calories: toNum(row.calories),
        meal_type: row.meal_type || 'snack'
      }
      const idx = this.entries.findIndex((x) => Number(x.id) === id)
      if (idx >= 0) this.entries.splice(idx, 1, normalized)
      else this.entries.unshift(normalized)
      this._rebuildEntriesByMeal()
    },

    async hydrateAll(day) {
      this.loading = true
      try {
        this.day = resolveJournalDay(day, this.day)
        const [entriesRes, favoritesRes, dashboardRes] = await Promise.all([
          api.get('/api/v1/nutrition/entries', { params: { day: this.day } }),
          api.get('/api/v1/nutrition/favorites'),
          api.get('/api/v1/nutrition/dashboard', { params: { day: this.day } })
        ])
        this.entries = entriesRes.data?.items ?? []
        this._rebuildEntriesByMeal()
        this.favorites = favoritesRes.data?.items ?? []
        this.dashboard = dashboardRes.data ?? null
        try {
          const statsRes = await api.get('/api/v1/nutrition/stats')
          this.stats = statsRes.data ?? null
        } catch {
          this.stats = null
        }
        await this.fetchGoal()
      } finally {
        this.loading = false
      }
    },

    async searchDishes(query = '', limit = 20) {
      const { data } = await api.get('/api/v1/nutrition/dishes/search', {
        params: { q: String(query || '').trim(), limit }
      })
      this.dishes = data?.items ?? []
      return this.dishes
    },

    async createDish(payload) {
      const body = {
        title: String(payload.title ?? '').trim(),
        protein_g: Number(toNum(payload.protein_g)),
        fat_g: Number(toNum(payload.fat_g)),
        carbs_g: Number(toNum(payload.carbs_g)),
        base_grams: Math.max(1, Number(toNum(payload.base_grams || 100)))
      }
      const kcal = toNum(payload.calories)
      if (kcal > 0) body.calories = Number(kcal)
      const { data } = await api.post('/api/v1/nutrition/dishes', body)
      return data
    },

    async fetchMyDishes() {
      const { data } = await api.get('/api/v1/nutrition/dishes/mine')
      this.myDishes = data?.items ?? []
      return this.myDishes
    },

    async updateDish(id, payload) {
      const body = {
        title: String(payload.title ?? '').trim(),
        protein_g: Number(toNum(payload.protein_g)),
        fat_g: Number(toNum(payload.fat_g)),
        carbs_g: Number(toNum(payload.carbs_g)),
        base_grams: Math.max(1, Number(toNum(payload.base_grams || 100)))
      }
      const kcal = toNum(payload.calories)
      if (kcal > 0) body.calories = Number(kcal)
      const { data } = await api.patch(`/api/v1/nutrition/dishes/${id}`, body)
      return data
    },

    async deleteDish(id) {
      await api.delete(`/api/v1/nutrition/dishes/${id}`)
      this.myDishes = this.myDishes.filter((x) => Number(x.id) !== Number(id))
    },

    async fetchGoal() {
      const res = await api.get('/api/v1/nutrition/goals', {
        validateStatus: (status) => status === 200 || status === 204
      })
      if (res.status === 204 || res.data == null || res.data === '') {
        this.goal = null
      } else {
        const d = res.data
        this.goal = {
          protein_g: toNum(d?.protein_g),
          fat_g: toNum(d?.fat_g),
          carbs_g: toNum(d?.carbs_g),
          calories: toNum(d?.calories),
          updated_at: d?.updated_at ?? null
        }
      }
      this.applyGoalFallbackFromDashboard()
    },

    applyGoalFallbackFromDashboard() {
      const g = this.dashboard?.goal
      const hasDashGoal = g && Number(g.calories) > 0
      if (!hasDashGoal) return
      const curKcal = Number(this.goal?.calories)
      if (!this.goal || !Number.isFinite(curKcal) || curKcal <= 0) {
        this.goal = {
          protein_g: g.protein_g,
          fat_g: g.fat_g,
          carbs_g: g.carbs_g,
          calories: g.calories,
          updated_at: this.goal?.updated_at ?? null
        }
        return
      }
      const pg = toNum(this.goal.protein_g)
      const fg = toNum(this.goal.fat_g)
      const cg = toNum(this.goal.carbs_g)
      if (pg <= 0 && fg <= 0 && cg <= 0 && (toNum(g.protein_g) > 0 || toNum(g.fat_g) > 0 || toNum(g.carbs_g) > 0)) {
        this.goal = {
          ...this.goal,
          protein_g: g.protein_g,
          fat_g: g.fat_g,
          carbs_g: g.carbs_g
        }
      }
    },

    async recalculateGoal(payload = '') {
      const body = typeof payload === 'string' ? { target: payload } : payload
      const { data } = await api.post('/api/v1/nutrition/goals/recalculate', body)
      this.goal = {
        protein_g: toNum(data?.protein_g),
        fat_g: toNum(data?.fat_g),
        carbs_g: toNum(data?.carbs_g),
        calories: toNum(data?.calories),
        updated_at: data?.updated_at ?? null
      }
      await this.fetchStats()
      await this.fetchDashboard(this.day)
      this.applyGoalFallbackFromDashboard()
      return this.goal
    },

    async fetchStats() {
      const { data } = await api.get('/api/v1/nutrition/stats')
      this.stats = data
      return data
    },

    async fetchDashboard(day) {
      const d = resolveJournalDay(day, this.day)
      const { data } = await api.get('/api/v1/nutrition/dashboard', { params: { day: d } })
      this.dashboard = data
      this.applyGoalFallbackFromDashboard()
      return data
    },

    async fetchReports(days = 30) {
      try {
        const { data } = await api.get('/api/v1/nutrition/analytics', { params: { days } })
        this.reports = data
        return data
      } catch {
        this.reports = null
        return null
      }
    },

    async addEntry(payload) {
      if (!payload?.day) {
        this.day = nutritionJournalDayISO()
      }
      const grams = Math.max(1, toNum(payload.grams || 100))
      const dayStr = resolveJournalDay(payload?.day, this.day)
      this.day = dayStr
      const body = {
        dish_id: payload.dish_id != null && payload.dish_id !== '' ? Number(payload.dish_id) : undefined,
        meal_type: payload.meal_type || 'snack',
        grams,
        title: String(payload.title ?? '').trim() || 'Без названия',
        protein_g: toNum(payload.protein_g),
        fat_g: toNum(payload.fat_g),
        carbs_g: toNum(payload.carbs_g),
        calories: toNum(payload.calories),
        day: dayStr
      }
      if (payload.consumed_at) body.consumed_at = payload.consumed_at
      const { data } = await api.post('/api/v1/nutrition/entries', body)
      if (data) this.mergeServerEntry(data)
      await this.hydrateAll(this.day)
      if (data?.id != null && !this.entries.some((e) => Number(e.id) === Number(data.id))) {
        this.mergeServerEntry(data)
        await this.fetchDashboard(this.day)
      }
      try {
        await this.fetchStats()
      } catch {

      }
      return data
    },

    async updateEntry(id, payload) {
      const grams = Math.max(1, toNum(payload.grams || 100))
      const dayStr = resolveJournalDay(payload?.day, this.day)
      this.day = dayStr
      const body = {
        meal_type: payload.meal_type || 'snack',
        grams,
        title: String(payload.title ?? '').trim() || 'Без названия',
        protein_g: toNum(payload.protein_g),
        fat_g: toNum(payload.fat_g),
        carbs_g: toNum(payload.carbs_g),
        calories: toNum(payload.calories),
        day: dayStr
      }
      if (payload.consumed_at) body.consumed_at = payload.consumed_at
      await api.patch(`/api/v1/nutrition/entries/${id}`, body)
      await this.hydrateAll(this.day)
      await this.fetchStats().catch(() => {})
    },

    async deleteEntry(id) {
      await api.delete(`/api/v1/nutrition/entries/${id}`)
      this.entries = this.entries.filter((x) => Number(x.id) !== Number(id))
      this._rebuildEntriesByMeal()
      await this.hydrateAll(this.day)
      await this.fetchStats().catch(() => {})
    },

    async addFavorite(payload) {
      const body = {
        title: String(payload.title ?? '').trim(),
        protein_g: toNum(payload.protein_g),
        fat_g: toNum(payload.fat_g),
        carbs_g: toNum(payload.carbs_g),
        unit_type: payload.unit_type || 'gram'
      }
      const { data } = await api.post('/api/v1/nutrition/favorites', body)
      this.favorites.unshift({ id: data.id, ...body })
      return data
    },

    async deleteFavorite(id) {
      await api.delete(`/api/v1/nutrition/favorites/${id}`)
      this.favorites = this.favorites.filter((x) => Number(x.id) !== Number(id))
    },

    async saveWater(amountMl, day) {
      const d = resolveJournalDay(day, this.day)
      this.day = d
      await api.post('/api/v1/nutrition/water', { amount_ml: Math.max(0, Number(amountMl || 0)), day: d })
      await this.fetchDashboard(d)
    },

    async saveWeight(weightKg, day) {
      const d = resolveJournalDay(day, this.day)
      this.day = d
      await api.post('/api/v1/nutrition/weight', { weight_kg: Number(weightKg), day: d })
      await this.fetchDashboard(d)
    },

    setSelectedDish(dish) {
      this.selectedDish = dish || null
    },

    clearSelectedDish() {
      this.selectedDish = null
    }
  }
})
