<script setup>
import { computed, reactive, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import { v2, getReference } from '@/api/v2'
import { usePages, useRead } from '@/composables/useRequest'
import { useAuthStore } from '@/stores/auth'
import { levels as levelLabels } from '@/utils/profile'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const levels = ['новичок', 'любитель', 'продвинутый', 'beginner', 'intermediate', 'advanced']
const filters = computed(() => ({
  goal: typeof route.query.goal === 'string' ? route.query.goal : '',
  level: typeof route.query.level === 'string' ? route.query.level : '',
  duration_weeks: /^\d+$/.test(route.query.duration_weeks ?? '') ? String(route.query.duration_weeks) : '',
  equipment: typeof route.query.equipment === 'string' ? route.query.equipment : '',
}))

const form = reactive({ ...filters.value })

const { items, cursor, loading, error, status, reset, more, retry } = usePages((cursor, signal) => {
  const params = { price_type: 'free', limit: 12, cursor }
  for (const [key, value] of Object.entries(filters.value)) {
    if (value) params[key] = value
  }
  return v2('GET', '/catalog/programs', { params, signal, anonymous: true })
})
const equipment = useRead(() => auth.token ? getReference('/equipment') : Promise.resolve([]))

/** Apply the selected filters and share them through the URL. */
function applyFilters() {
  const query = Object.fromEntries(
    Object.entries({ ...form, goal: form.goal.trim() }).filter(([, value]) => value !== '' && value != null),
  )
  return router.replace({ name: 'Marketplace', query })
}

/** Clear both unsubmitted filters and the applied URL parameters. */
function clearFilters() {
  Object.assign(form, { goal: '', level: '', duration_weeks: '', equipment: '' })
  return applyFilters()
}

watch(() => route.name === 'Marketplace' && JSON.stringify(filters.value), (key) => {
  if (key) {
    Object.assign(form, filters.value)
    reset()
  }
}, { immediate: true })
watch(() => auth.token, equipment.load, { immediate: true })
</script>

<template>
  <FeaturePage eyebrow="Маркетплейс" title="Каталог программ" subtitle="Найдите бесплатный план под вашу цель, опыт и оборудование." tab="catalog">
    <div class="feature-stack">
      <form class="feature-card catalog-filters" aria-label="Фильтры каталога" @submit.prevent="applyFilters">
        <div class="feature-fields">
          <label>Цель
            <input v-model="form.goal" type="search" placeholder="Например, Набрать мышцы" maxlength="200" />
          </label>
          <label>Уровень
            <select v-model="form.level">
              <option value="">Любой</option>
              <option v-for="level in levels" :key="level" :value="level">{{ levelLabels[level] ?? level + ' · классические программы' }}</option>
            </select>
          </label>
          <label>Длительность, недель
            <input type="number" min="1" v-model="form.duration_weeks" placeholder="Любая" />
          </label>
          <label>Оборудование
            <select v-model="form.equipment" :disabled="equipment.loading.value || !auth.token">
              <option value="">Любое</option>
              <option v-for="item in equipment.data.value ?? []" :key="item.id" :value="item.id">{{ item.name }}</option>
            </select>
          </label>
        </div>
        <div class="feature-row">
          <button class="feature-button" type="submit">Применить фильтры</button>
          <button type="button" class="feature-button secondary" @click="clearFilters">Сбросить фильтры</button>
          <RouterLink v-if="!auth.token" class="feature-link" :to="{ path: '/auth/login', query: { return_to: route.fullPath } }">
            Войти для выбора оборудования
          </RouterLink>
        </div>
        <RequestState :error="equipment.error.value" :status="equipment.status.value" :retry="equipment.load" />
      </form>

      <div v-if="items.length" class="feature-grid">
        <article v-for="program in items" :key="program.id" class="feature-card catalog-card">
          <span class="feature-badge">Бесплатно</span>
          <h2>{{ program.title }}</h2>
          <p class="feature-muted">{{ program.trainer.display_name }} · {{ levelLabels[program.level] ?? program.level }}</p>
          <p class="feature-muted">{{ program.goal }} · {{ program.duration_weeks }} нед. · {{ program.sessions_per_week }} в неделю</p>
          <RouterLink class="feature-button" :aria-label="'Открыть программу ' + program.title" :to="{ name: 'Program', params: { programId: program.id }, query: route.query }">
            Открыть программу
          </RouterLink>
        </article>
      </div>
      <RequestState :loading="loading" :error="error" :status="status" :retry="retry" :empty="items.length ? '' : 'Программы не найдены. Измените фильтры или сбросьте их.'" />
      <button v-if="cursor && !error" type="button" class="feature-button secondary" :disabled="loading" @click="more">
        Показать ещё
      </button>
    </div>
  </FeaturePage>
</template>

<style scoped>
.catalog-card { display: flex; flex-direction: column; align-items: flex-start; }
.catalog-card .feature-button { margin-top: auto; }
@media (min-width: 360px) and (max-width: 540px) {
  .catalog-filters .feature-fields { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .catalog-filters label { grid-template-rows: 2.4em auto; }
}
</style>
