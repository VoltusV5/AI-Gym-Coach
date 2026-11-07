<script setup>
import { computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import { v2, getReference } from '@/api/v2'
import { useRead, useWrite } from '@/composables/useRequest'
import { useAuthStore } from '@/stores/auth'
import { levels } from '@/utils/profile'

const props = defineProps({ programId: { type: String, required: true } })
const route = useRoute()
const auth = useAuthStore()
const path = computed(() => `/catalog/programs/${encodeURIComponent(props.programId)}`)
const { data: program, loading, error, status, load } = useRead(
  (signal) => v2('GET', path.value, { signal }), 'Эта программа недоступна или снята с публикации.',
)
const { saving, error: accessError, status: accessStatus, run } = useWrite()
const exerciseList = useRead(() => getReference('/exercises'))
const names = computed(() => new Map((exerciseList.data.value ?? []).map(({ id, name }) => [id, name])))

watch([() => props.programId, () => auth.token], load, { immediate: true })
watch(() => Boolean(program.value?.days?.length), (hasDays) => {
  if (hasDays && auth.token && !exerciseList.data.value) exerciseList.load()
})

/** Request the entitlement and reload details from the server. */
function grant() {
  return run(async () => {
    await v2('POST', `/programs/${encodeURIComponent(props.programId)}/access`)
    await load()
  })
}
</script>

<template>
  <FeaturePage eyebrow="Маркетплейс" :title="program?.title ?? 'Программа'" :back="{ name: 'Marketplace', query: route.query }">
    <RequestState :loading="loading" :error="error" :status="status" :retry="load" />
    <div v-if="program" class="feature-stack">
      <section class="feature-card">
        <div class="feature-row">
          <span class="feature-badge">{{ program.price_minor === 0 ? 'Бесплатно' : 'Предпросмотр' }}</span>
          <span v-if="program.trainer.verified" class="feature-badge">Тренер подтверждён</span>
        </div>
        <p><RouterLink class="feature-link" :to="`/trainers/${program.trainer.id}`">{{ program.trainer.display_name }}</RouterLink></p>
        <p class="feature-text">{{ program.description }}</p>
        <p class="feature-muted">{{ program.goal }} · {{ levels[program.level] ?? program.level }} · {{ program.duration_weeks }} нед. · {{ program.sessions_per_week }} в неделю</p>
        <p class="feature-muted">Оборудование: {{ program.equipment?.map((item) => item.name).join(', ') || 'без оборудования' }}</p>
        <p v-if="program.access !== 'preview'" class="feature-success" role="status">Программа доступна</p>
        <template v-else-if="program.allowed_actions?.includes('access')">
          <RouterLink v-if="!auth.token || auth.isGuest" class="feature-button" :to="{ path: '/auth/login', query: { return_to: route.fullPath } }">
            Войти и получить доступ
          </RouterLink>
          <button v-else type="button" class="feature-button" :disabled="saving" @click="grant">
            {{ saving ? 'Получаем доступ…' : 'Получить бесплатно' }}
          </button>
        </template>
        <RequestState :error="accessError" :status="accessStatus" />
      </section>

      <section v-for="day in program.days ?? []" :key="day.id" class="feature-card">
        <h2>{{ day.title }}</h2>
        <p v-for="exercise in day.exercises" :key="exercise.position">
          {{ names.get(exercise.exercise_id) ?? `Упражнение №${exercise.exercise_id}` }} · {{ exercise.sets }} × {{ exercise.reps_min }}–{{ exercise.reps_max }}
        </p>
      </section>
    </div>
  </FeaturePage>
</template>
