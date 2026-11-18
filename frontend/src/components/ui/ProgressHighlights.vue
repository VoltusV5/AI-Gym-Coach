<script setup>
import { computed, watch } from 'vue'
import { useExerciseNames } from '@/composables/useExerciseNames'
import { metricNumber } from '@/utils/metrics'

const props = defineProps({ items: { type: Array, default: () => [] } })
const names = useExerciseNames()
const groups = computed(() => {
  const groups = new Map()
  for (const item of props.items) {
    const key = item.exercise_id ?? 'general'
    if (!groups.has(key)) groups.set(key, { exercise: item.exercise_id, records: new Map() })
    const records = groups.get(key).records
    const metric = item.metric || item.message
    if (!records.has(metric) || item.current >= records.get(metric).current)
      records.set(metric, item)
  }
  return [...groups.values()]
})
watch(
  () => props.items,
  (items) => {
    const ids = [...new Set(items.map((item) => item.exercise_id).filter((id) => id != null))]
    if (ids.length) names.load(ids)
  },
  { immediate: true },
)
const labels = { working_weight_kg: 'Лучший рабочий вес', volume_kg: 'Лучший объём за занятие' }
</script>

<template>
  <ul class="progress-records">
    <li v-for="(group, index) in groups" :key="group.exercise ?? index">
      <h3 v-if="group.exercise != null">
        <router-link :to="`/progress/exercises/${group.exercise}`">{{
          names.name(group.exercise)
        }}</router-link>
      </h3>
      <dl>
        <div v-for="[key, item] in group.records" :key="key">
          <template v-if="labels[item.metric]"
            ><dt>{{ labels[item.metric] }}</dt>
            <dd>{{ metricNumber(item.current) }} кг</dd></template
          >
          <template v-else
            ><dt>Результат</dt>
            <dd>{{ item.message }}</dd></template
          >
        </div>
      </dl>
    </li>
  </ul>
</template>

<style scoped>
.progress-records {
  list-style: none;
  padding: 0;
  margin: 0;
}
.progress-records li {
  padding: 16px 0;
  border-top: 1px solid var(--sportik-border);
}
.progress-records li:first-child {
  border-top: 0;
  padding-top: 0;
}
.progress-records li:last-child {
  padding-bottom: 0;
}
.progress-records h3 {
  margin: 0 0 12px;
  font-size: 0.97rem;
}
.progress-records a {
  color: inherit;
  text-decoration: none;
}
.progress-records a:hover {
  text-decoration: underline;
}
.progress-records dl {
  display: flex;
  flex-wrap: wrap;
  gap: 14px 28px;
  margin: 0;
}
.progress-records dt {
  color: var(--sportik-text-soft);
  font-size: 0.82rem;
}
.progress-records dd {
  margin: 4px 0 0;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
</style>
