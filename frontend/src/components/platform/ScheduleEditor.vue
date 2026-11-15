<template>
  <section class="platform-panel">
    <h2>Расписание</h2>
    <div class="platform-fields">
      <label
        >Дата начала<input
          :value="modelValue.start_date"
          type="date"
          required
          @input="change('start_date', $event.target.value)" /></label
      ><label
        >Часовой пояс<input
          :value="modelValue.time_zone"
          required
          placeholder="Europe/Moscow"
          @input="change('time_zone', $event.target.value)"
      /></label>
    </div>
    <div class="platform-fields">
      <label v-for="(label, i) in weekDays" :key="label"
        >{{ label
        }}<select
          :value="modelValue.entries.find((entry) => entry.weekday === i + 1)?.plan_day_id || ''"
          @change="assign(i + 1, $event.target.value)"
        >
          <option value="">Отдых</option>
          <option v-for="day in days" :key="day.id" :value="day.id">{{ day.title }}</option>
        </select></label
      >
    </div>
  </section>
</template>
<script setup>
import { weekDays, clone } from '@/utils/platform'
const props = defineProps({
  modelValue: { type: Object, required: true },
  days: { type: Array, required: true },
})
const emit = defineEmits(['update:modelValue'])
function change(key, value) {
  emit('update:modelValue', { ...clone(props.modelValue), [key]: value })
}
function assign(weekday, id) {
  const entries = props.modelValue.entries.filter((entry) => entry.weekday !== weekday)
  if (id) entries.push({ weekday, plan_day_id: id })
  change(
    'entries',
    entries.sort((a, b) => a.weekday - b.weekday),
  )
}
</script>
