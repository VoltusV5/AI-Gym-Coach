<script setup>
import { ref, watch } from 'vue'
import { onIonViewWillEnter } from '@ionic/vue'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import ProgressMetrics from '@/components/ui/ProgressMetrics.vue'
import { v2 } from '@/api/v2'
import { usePages } from '@/composables/useRequest'
import { dateLabel } from '@/utils/pages'
import { timeZone } from '@/utils/plans'

const props = defineProps({ summaries: Boolean })
const period = ref('week')
const { items, cursor, loading, error, status, reset, more, retry } = usePages((cursor, signal) => {
  const params = props.summaries ? { cursor, period: period.value, time_zone: timeZone() } : { cursor }
  return v2('GET', props.summaries ? '/progress/summaries' : '/workouts', { params, signal })
})
watch(period, reset)
onIonViewWillEnter(reset)
</script>

<template>
  <FeaturePage eyebrow="Мой прогресс" :title="summaries ? 'Сводки' : 'История тренировок'" back="/progress">
    <div class="feature-stack">
      <div v-if="summaries" class="feature-card">
        <label>Период сводки<select v-model="period"><option value="week">Неделя</option><option value="month">Месяц</option></select></label>
      </div>
      <article v-for="item in items" :key="item.id" class="feature-card">
        <h2>{{ summaries ? item.period_start : dateLabel(item.finished_at) }}</h2>
        <template v-if="summaries">
          <ProgressMetrics :metrics="item.metrics" />
          <p v-for="message in item.messages" :key="message">{{ message }}</p>
        </template>
        <template v-else>
          <p class="feature-muted">Упражнений: {{ item.slots.length }} · Версия плана {{ item.plan_version }}</p>
          <RouterLink class="feature-button secondary" :to="`/progress/workouts/${item.id}`">Открыть тренировку</RouterLink>
        </template>
      </article>
      <RequestState :loading="loading" :error="error" :status="status" :retry="retry" :empty="items.length ? '' : 'После завершения тренировок здесь появятся данные.'"/>
      <button v-if="cursor && !error" class="feature-button secondary" type="button" :disabled="loading" @click="more">Показать ещё</button>
    </div>
  </FeaturePage>
</template>
