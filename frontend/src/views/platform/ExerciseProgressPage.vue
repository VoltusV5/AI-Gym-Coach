<template>
  <platform-layout
    :title="names.name(Number(route.params.id))"
    subtitle="Рабочие подходы: разминка исключена. Сравнение доступно после сохранённых занятий."
    ><template #actions><router-link to="/progress">К прогрессу</router-link></template
    ><async-state
      :loading="load.loading.value"
      :error="load.error.value"
      :empty="!progress?.points.length"
      @retry="refresh"
      ><template #empty>Для этого упражнения пока нет рабочих подходов.</template
      ><template v-if="progress"
        ><section class="platform-panel">
          <h2>Рабочий объём, кг</h2>
          <svg
            v-if="progress.points.length > 1"
            viewBox="0 0 600 180"
            role="img"
            aria-label="Изменение рабочего объёма по тренировкам"
            class="progress-chart"
          >
            <polyline :points="chartPoints" fill="none" stroke="currentColor" stroke-width="3" />
          </svg>
          <div class="progress-table">
            <table>
              <caption>
                Результаты по занятиям
              </caption>
              <thead>
                <tr>
                  <th>Дата</th>
                  <th>Рабочий вес, кг</th>
                  <th>Повторения</th>
                  <th>Объём, кг</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="point in progress.points" :key="point.session_id">
                  <td>
                    <router-link :to="`/workouts/${point.session_id}`">{{
                      new Date(point.finished_at).toLocaleDateString('ru-RU')
                    }}</router-link>
                  </td>
                  <td>{{ point.working_weight_kg }}</td>
                  <td>{{ point.reps }}</td>
                  <td>{{ point.volume_kg }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
        <section v-if="progress.records.length" class="platform-panel">
          <h2>Личные рекорды</h2>
          <p v-for="(item, index) in progress.records" :key="index">{{ item.message }}</p>
        </section></template
      ></async-state
    ></platform-layout
  >
</template>
<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/api/api'
import { useRequest } from '@/composables/useAsync'
import { useViewLoad } from '@/composables/useViewLoad'
import { useExerciseNames } from '@/composables/useExerciseNames'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const route = useRoute(),
  load = useRequest(),
  names = useExerciseNames(),
  progress = computed(() => load.data.value)
const chartPoints = computed(() => {
  const points = progress.value?.points || [],
    max = Math.max(1, ...points.map((point) => point.volume_kg))
  return points
    .map(
      (point, index) =>
        `${15 + (index * 570) / Math.max(1, points.length - 1)},${165 - (point.volume_kg / max) * 150}`,
    )
    .join(' ')
})
async function refresh() {
  await Promise.all([
    load.run(async () => (await api.get(`/api/v2/progress/exercises/${route.params.id}`)).data),
    names.load([Number(route.params.id)]),
  ])
}
useViewLoad(refresh)
</script>
<style scoped>
.progress-chart {
  width: 100%;
  color: var(--sportik-brand);
}
.progress-table {
  overflow-x: auto;
}
table {
  width: 100%;
  border-collapse: collapse;
  text-align: left;
}
td,
th {
  padding: 12px;
  border-bottom: 1px solid var(--sportik-border);
}
caption {
  text-align: left;
  margin-bottom: 12px;
}
</style>
