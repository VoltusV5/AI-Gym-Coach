<script setup>
import { ref, watch } from 'vue'
import { onIonViewWillEnter } from '@ionic/vue'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import ProgressMetrics from '@/components/ui/ProgressMetrics.vue'
import { v2 } from '@/api/v2'
import { useRead, useWrite } from '@/composables/useRequest'
import { localDate } from '@/utils/pages'
import { timeZone } from '@/utils/plans'

const period = ref('week')
const day = ref(localDate())
const weight = ref('')
const saved = ref(false)
const zone = timeZone()
const { data, loading, error, status, load } = useRead((signal) =>
  v2('GET', '/progress', { params: { period: period.value, time_zone: zone }, signal }),
)
const { saving, error: weightError, status: weightStatus, run } = useWrite()
watch(period, load)
onIonViewWillEnter(load)

/** Record a measurement for the chosen local calendar day. */
function recordWeight() {
  return run(async () => {
    saved.value = false
    await v2('POST', '/progress/weight', { data: { day: day.value, weight_kg: weight.value } })
    saved.value = true
    weight.value = ''
    await load()
  })
}
</script>

<template>
  <FeaturePage eyebrow="Статистика" title="Мой прогресс" subtitle="Фактические тренировки, регулярность и динамика нагрузки." tab="progress">
    <div class="feature-stack">
      <nav class="feature-row" aria-label="Разделы прогресса">
        <RouterLink class="feature-button secondary" to="/progress/history">История тренировок</RouterLink>
        <RouterLink class="feature-button secondary" to="/progress/summaries">Сводки по периодам</RouterLink>
      </nav>
      <section class="feature-card" aria-label="Итоги периода">
        <label>Период<select v-model="period"><option value="week">Неделя</option><option value="month">Месяц</option><option value="all">Всё время</option></select></label>
        <p class="feature-muted">Часовой пояс: {{ zone }}</p>
        <RequestState :loading="loading" :error="error" :status="status" :retry="load"
          :empty="data && data.workout_count === 0 && data.weight_change_kg == null ? 'Пока нет тренировок за этот период. После первого занятия здесь появятся объём и регулярность.' : ''" />
        <ProgressMetrics v-if="data && (data.workout_count || data.weight_change_kg != null)" :metrics="data" />
        <p v-for="item in data?.highlights ?? []" :key="`${item.exercise_id}-${item.metric}`">{{ item.message }}</p>
      </section>
      <form class="feature-card" @submit.prevent="recordWeight">
        <h2>Измерение веса</h2>
        <fieldset class="feature-fields" :disabled="saving">
          <legend>Новое измерение</legend>
          <label>Дата<input v-model="day" type="date" required /></label>
          <label>Вес, кг<input v-model.number="weight" type="number" min="0.01" step="any" required /></label>
        </fieldset>
        <button class="feature-button" :disabled="saving">{{ saving ? 'Сохраняем…' : 'Записать вес' }}</button>
        <p v-if="saved" class="feature-success" role="status">Измерение сохранено</p>
        <RequestState :error="weightError" :status="weightStatus" />
      </form>
    </div>
  </FeaturePage>
</template>
