<script setup>
import { metricNumber, weightTrend, metricPeriod, metricRange } from '@/utils/metrics'
import VolumeHelp from '@/components/ui/VolumeHelp.vue'
defineProps({ metrics: { type: Object, required: true } })
</script>

<template>
  <p v-if="metrics.period" class="feature-muted metric-period">
    {{ metricPeriod(metrics)
    }}<span v-if="metricRange(metrics)"> · {{ metricRange(metrics) }}</span>
  </p>
  <dl class="feature-metrics">
    <div v-if="metrics.workout_count !== undefined">
      <dt>Тренировки</dt>
      <dd>{{ metrics.workout_count }}</dd>
    </div>
    <div>
      <dt>{{ metrics.workout_count !== undefined ? 'Объём за период' : 'Объём тренировки' }}</dt>
      <dd>{{ metricNumber(metrics.volume_kg) }} кг</dd>
    </div>
    <div v-if="metrics.working_sets !== undefined">
      <dt>Рабочие подходы</dt>
      <dd>{{ metrics.working_sets }}</dd>
    </div>
    <div v-if="metrics.exercise_count !== undefined">
      <dt>Упражнения</dt>
      <dd>{{ metrics.exercise_count }}</dd>
    </div>
    <div v-if="metrics.duration_sec !== undefined">
      <dt>Время</dt>
      <dd>{{ Math.floor(metrics.duration_sec / 60) }} мин {{ metrics.duration_sec % 60 }} с</dd>
    </div>
    <div v-if="metrics.streak_weeks !== undefined">
      <dt>Серия, недель</dt>
      <dd>{{ metrics.streak_weeks }}</dd>
    </div>
    <div v-if="metrics.adherence !== undefined">
      <dt>Выполнение плана</dt>
      <dd>{{ metrics.adherence == null ? '—' : metricNumber(metrics.adherence * 100) + ' %' }}</dd>
    </div>
    <div v-if="metrics.weight_change_kg !== undefined">
      <dt>Изменение веса</dt>
      <dd>
        {{ weightTrend(metrics.weight_change_kg).value
        }}<span v-if="metrics.weight_change_kg != null" class="metric-caption">{{
          weightTrend(metrics.weight_change_kg).label
        }}</span>
      </dd>
    </div>
  </dl>
  <volume-help />
</template>
