<script setup>
import { computed, watch } from 'vue'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import ProgressMetrics from '@/components/ui/ProgressMetrics.vue'
import { getReference, v2 } from '@/api/v2'
import { useRead } from '@/composables/useRequest'
import { useAuthStore } from '@/stores/auth'
import { dateLabel } from '@/utils/pages'
import { numberLabel, readTab } from '@/utils/workoutV2'

const props = defineProps({ sessionId: { type: String, required: true }, success: Boolean })
const auth = useAuthStore()
const { data, loading, error, status, load } = useRead(async (signal) => {
  const workout = await v2('GET', `/workouts/${encodeURIComponent(props.sessionId)}`, { signal })
  const exercises = await getReference('/exercises').catch(() => [])
  return { workout, exercises }
}, 'Тренировка не найдена или вам недоступна.')
const result = computed(() => data.value && readTab(`sportik-result:${auth.accountId}:${props.sessionId}`))
const names = computed(() => new Map((data.value?.exercises ?? []).map(({ id, name }) => [id, name])))
watch([() => props.sessionId, () => auth.token], load, { immediate: true })
</script>

<template>
  <FeaturePage eyebrow="Мой прогресс" :title="success && data ? 'Тренировка сохранена' : 'Тренировка'" back="/progress/history">
    <div class="feature-stack">
      <RequestState :loading="loading" :error="error" :status="status" :retry="load" />
      <template v-if="data">
        <section class="feature-card">
          <h2>{{ dateLabel(data.workout.finished_at) }}</h2>
          <p class="feature-muted">Версия плана {{ data.workout.plan_version }}</p>
          <ProgressMetrics v-if="result" :metrics="result.metrics" />
          <p v-for="item in result?.highlights ?? []" :key="`${item.exercise_id}-${item.metric}`">{{ item.message }}</p>
          <RouterLink class="feature-button" to="/progress">Смотреть прогресс</RouterLink>
        </section>
        <section v-for="slot in data.workout.slots" :key="slot.position" class="feature-card">
          <h2>{{ names.get(slot.exercise_id) ?? `Упражнение №${slot.exercise_id}` }}</h2>
          <p v-for="(set, index) in slot.sets" :key="index">
            {{ index + 1 }}. {{ set.reps }} повт. × {{ numberLabel.format(set.weight_kg) }} кг{{ set.is_warmup ? ' · разминка' : '' }}
          </p>
          <RouterLink class="feature-link" :to="`/progress/exercises/${slot.exercise_id}`">Динамика упражнения</RouterLink>
        </section>
      </template>
    </div>
  </FeaturePage>
</template>
