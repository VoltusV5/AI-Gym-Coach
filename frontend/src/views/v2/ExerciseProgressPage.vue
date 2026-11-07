<script setup>
import { computed, watch } from 'vue'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import { v2 } from '@/api/v2'
import { useRead } from '@/composables/useRequest'
import { dateLabel } from '@/utils/pages'
import { numberLabel } from '@/utils/workoutV2'

const props = defineProps({ exerciseId: { type: String, required: true } })
const { data, error, loading, status, load } = useRead((signal) =>
  v2('GET', `/progress/exercises/${encodeURIComponent(props.exerciseId)}`, { signal }),
)
const max = computed(() => Math.max(1, ...(data.value?.points ?? []).map((point) => point.working_weight_kg)))
watch(() => props.exerciseId, load, { immediate: true })
</script>

<template>
  <FeaturePage eyebrow="Мой прогресс" title="Динамика упражнения" back="/progress/history">
    <div class="feature-stack">
      <RequestState :loading="loading" :error="error" :status="status" :retry="load"
        :empty="data && !data.points.length ? 'Для этого упражнения пока нет рабочих подходов.' : ''" />
      <section v-if="data?.points.length" class="feature-card">
        <h2>Рабочий вес по тренировкам</h2>
        <div class="feature-plot" aria-hidden="true">
          <span v-for="point in data.points" :key="point.session_id" :style="{ height: (100 * point.working_weight_kg / max) + '%' }" />
        </div>
        <div class="feature-overflow">
          <table class="feature-table" aria-label="Рабочий вес по тренировкам">
            <thead><tr><th scope="col">Дата</th><th scope="col">Вес, кг</th><th scope="col">Повторы</th><th scope="col">Объём, кг</th></tr></thead>
            <tbody><tr v-for="point in data.points" :key="point.session_id">
              <td><RouterLink class="feature-link" :to="`/progress/workouts/${point.session_id}`">{{ dateLabel(point.finished_at) }}</RouterLink></td>
              <td>{{ numberLabel.format(point.working_weight_kg) }}</td><td>{{ point.reps }}</td><td>{{ numberLabel.format(point.volume_kg) }}</td>
            </tr></tbody>
          </table>
        </div>
        <p v-for="record in data.records" :key="record.metric">{{ record.message }}</p>
      </section>
    </div>
  </FeaturePage>
</template>
