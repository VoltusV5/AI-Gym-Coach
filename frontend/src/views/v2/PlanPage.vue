<script setup>
import { computed, ref } from 'vue'
import { useRouter } from 'vue-router'
import { onIonViewWillEnter } from '@ionic/vue'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import { getReference, v2 } from '@/api/v2'
import { useRead, useWrite } from '@/composables/useRequest'
import { sourceLabels } from '@/utils/plans'
import { weekdays } from '@/utils/profile'

const props = defineProps({ planId: { type: String, required: true } })
const router = useRouter()
const path = computed(() => `/me/plans/${encodeURIComponent(props.planId)}`)
const { data: plan, loading, error, status, load } = useRead(
  (signal) => v2('GET', path.value, { signal }), 'План не найден или вам недоступен.',
)
const exercises = useRead(() => getReference('/exercises'))
const names = computed(() => new Map((exercises.data.value ?? []).map(({ id, name }) => [id, name])))
const { saving, error: writeError, status: writeStatus, run } = useWrite()
const archiveOpen = ref(false)
onIonViewWillEnter(() => {
  load()
  exercises.load()
})

/** Activate the version shown on screen. */
function activate() {
  return run(async () => {
    plan.value = await v2('POST', path.value + '/activate', { data: { version: plan.value.version } })
  })
}

/** Archive after explicit confirmation and return to the library. */
function archive() {
  return run(async () => {
    await v2('DELETE', path.value)
    await router.replace('/plans')
  })
}
</script>

<template>
  <FeaturePage eyebrow="Мои планы" :title="plan?.title ?? 'План тренировки'" back="/plans">
    <RequestState :loading="loading" :error="error" :status="status" :retry="load" />
    <div v-if="plan" class="feature-stack">
      <section class="feature-card">
        <div class="feature-row">
          <span class="feature-badge">{{ sourceLabels[plan.source_type] ?? plan.source_type }}</span>
          <span v-if="plan.is_active" class="feature-badge" role="status">Активный план</span>
          <span class="feature-muted">Версия {{ plan.version }}</span>
        </div>
        <p>Старт {{ plan.schedule.start_date }} · {{ plan.schedule.time_zone }}</p>
        <p v-if="plan.schedule.entries.length" class="feature-muted">
          {{ plan.schedule.entries.map((entry) => weekdays[entry.weekday - 1] + ' · ' + plan.days.find((day) => day.id === entry.plan_day_id)?.title).join(', ') }}
        </p>
        <div class="feature-row">
          <RouterLink class="feature-button" :to="`/plans/${plan.id}/edit`">Редактировать</RouterLink>
          <button v-if="!plan.is_active" type="button" class="feature-button secondary" :disabled="saving || writeStatus === 409" @click="activate">Сделать активным</button>
          <button type="button" class="feature-button secondary" :disabled="saving" @click="archiveOpen = !archiveOpen">Архивировать</button>
        </div>
        <div v-if="archiveOpen" class="feature-row">
          <p>Убрать план из библиотеки?</p>
          <button type="button" class="feature-button" :disabled="saving" @click="archive">Подтвердить архивирование</button>
          <button type="button" class="feature-button secondary" :disabled="saving" @click="archiveOpen = false">Отмена</button>
        </div>
        <RequestState :error="writeError" :status="writeStatus" />
        <button v-if="writeStatus === 409" class="feature-button secondary" type="button" @click="load().then(() => { writeStatus = 0; writeError = '' })">Перечитать план</button>
      </section>

      <p v-if="!plan.days.length" class="feature-state">Добавьте день и упражнения в редакторе.</p>
      <section v-for="day in plan.days" :key="day.id" class="feature-card">
        <h2>{{ day.title }}</h2>
        <p v-for="exercise in day.exercises" :key="exercise.position">
          {{ names.get(exercise.exercise_id) ?? `Упражнение №${exercise.exercise_id}` }} · {{ exercise.sets }} × {{ exercise.reps_min }}–{{ exercise.reps_max }}
        </p>
        <RouterLink v-if="day.exercises.length" class="feature-button" :to="`/workout/session/${plan.id}/${day.id}`">Начать тренировку</RouterLink>
      </section>
    </div>
  </FeaturePage>
</template>
