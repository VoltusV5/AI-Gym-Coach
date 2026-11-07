<script setup>
import { computed, ref, watch } from 'vue'
import { newDay, newId } from '@/utils/plans'

const days = defineModel({ type: Array, required: true })
const props = defineProps({
  exercises: { type: Array, required: true },
  equipment: { type: Array, default: () => [] },
})
const emit = defineEmits(['removed-day'])
const query = ref('')
const selectedEquipment = ref('')
const target = ref(null)
const exerciseId = ref('')
const names = computed(() => new Map(props.exercises.map(({ id, name }) => [id, name])))
const options = computed(() => props.exercises.filter((exercise) => {
  const needle = query.value.trim().toLocaleLowerCase('ru')
  return (!selectedEquipment.value || exercise.equipment_ids?.includes(selectedEquipment.value)) &&
    (!needle || `${exercise.name} ${exercise.muscle_group}`.toLocaleLowerCase('ru').includes(needle))
}))
watch(options, (items) => {
  if (!items.some(({ id }) => id === exerciseId.value)) exerciseId.value = ''
})

/** Add a day with a stable key for new exercise and schedule entries. */
function addDay() {
  days.value.push(newDay(`День ${days.value.length + 1}`))
}

/** Copy prescriptions under a new day ID. Schedule assignments stay with the original. */
function copyDay(day) {
  days.value.push({
    ...newDay(`${day.title.slice(0, 191)} (копия)`),
    exercises: day.exercises.map((exercise) => ({ ...exercise, editorId: newId() })),
  })
}

/** Remove the day and let the parent clear schedule references. */
function removeDay(index) {
  const [day] = days.value.splice(index, 1)
  if (target.value?.dayId === day.id) target.value = null
  emit('removed-day', day.id)
}

/** Open the exercise picker for a new or existing slot. */
function choose(day, index = null) {
  target.value = { dayId: day.id, index }
  query.value = ''
  selectedEquipment.value = ''
  exerciseId.value = ''
}

/** Apply the selected reference while preserving existing prescription values on replacement. */
function apply(day) {
  const index = target.value.index
  if (!exerciseId.value) return
  if (index === null) {
    day.exercises.push({
      editorId: newId(), exercise_id: exerciseId.value, sets: 3, reps_min: 8, reps_max: 12,
      weight_kg: null, rest_sec: 60,
    })
  } else {
    day.exercises[index].exercise_id = exerciseId.value
  }
  target.value = null
}

/** Move one prescription to its neighbor without changing its fields. */
function move(list, from, to) {
  if (to < 0 || to >= list.length) return
  const [entry] = list.splice(from, 1)
  list.splice(to, 0, entry)
}
</script>

<template>
  <section class="feature-stack" aria-label="Дни и упражнения">
    <article v-for="(day, dayIndex) in days" :key="day.id" class="feature-card">
      <h3>День {{ dayIndex + 1 }}</h3>
      <label>Название дня <input v-model.trim="day.title" required maxlength="200" /></label>
      <fieldset v-for="(exercise, index) in day.exercises" :key="exercise.editorId ?? exercise.position" class="feature-stack">
        <legend>{{ names.get(exercise.exercise_id) ?? `Упражнение №${exercise.exercise_id}` }}</legend>
        <div class="feature-fields">
          <label>Подходы<input v-model.number="exercise.sets" type="number" required min="1" max="100" /></label>
          <label>Повторы от<input v-model.number="exercise.reps_min" type="number" required min="1" /></label>
          <label>Повторы до<input v-model.number="exercise.reps_max" type="number" required :min="exercise.reps_min || 1" /></label>
          <label>Вес, кг<input v-model.number="exercise.weight_kg" type="number" min="0" step="any" /></label>
          <label>Отдых, сек<input v-model.number="exercise.rest_sec" type="number" required min="0" /></label>
        </div>
        <div class="feature-row">
          <button type="button" class="feature-button secondary" :disabled="index === 0 || !!target" @click="move(day.exercises, index, index - 1)">Выше</button>
          <button type="button" class="feature-button secondary" :disabled="index === day.exercises.length - 1 || !!target" @click="move(day.exercises, index, index + 1)">Ниже</button>
          <button type="button" class="feature-button secondary" @click="choose(day, index)">Заменить</button>
          <button type="button" class="feature-button secondary" @click="target = null; day.exercises.splice(index, 1)">Удалить упражнение</button>
        </div>
      </fieldset>
      <fieldset v-if="target?.dayId === day.id" class="feature-stack">
        <legend>Выбор упражнения</legend>
        <div class="feature-fields">
          <label>Поиск по названию или группе <input v-model="query" type="search" /></label>
          <label>Оборудование
            <select v-model="selectedEquipment">
              <option value="">Любое</option>
              <option v-for="item in equipment" :key="item.id" :value="item.id">{{ item.name }}</option>
            </select>
          </label>
        </div>
        <label>Упражнение
          <select v-model="exerciseId">
            <option value="">Выберите упражнение</option>
            <option v-for="exercise in options" :key="exercise.id" :value="exercise.id">{{ exercise.name }}</option>
          </select>
        </label>
        <p v-if="!options.length">По выбранным условиям ничего не найдено.</p>
        <div class="feature-row">
          <button type="button" class="feature-button" :disabled="!exerciseId" @click="apply(day)">Выбрать</button>
          <button type="button" class="feature-button secondary" @click="target = null">Отменить выбор</button>
        </div>
      </fieldset>
      <div class="feature-row">
        <button type="button" class="feature-button secondary" @click="choose(day)">Добавить упражнение</button>
        <button type="button" class="feature-button secondary" @click="copyDay(day)">Копировать день</button>
        <button type="button" class="feature-button secondary" @click="removeDay(dayIndex)">Удалить день</button>
      </div>
    </article>
    <button type="button" class="feature-button secondary" @click="addDay">Добавить день</button>
  </section>
</template>
