<template>
  <section>
    <article v-for="(day, index) in modelValue" :key="day.id" class="platform-panel">
      <div class="platform-fields">
        <label
          >Название дня<input
            :value="day.title"
            required
            maxlength="200"
            @input="changeDay(index, 'title', $event.target.value)"
        /></label>
      </div>
      <div class="day-actions">
        <button type="button" :disabled="index === 0" @click="moveDay(index, -1)">
          Поднять день</button
        ><button
          type="button"
          :disabled="index === modelValue.length - 1"
          @click="moveDay(index, 1)"
        >
          Опустить день</button
        ><button type="button" @click="copyDay(index)">Копировать день</button
        ><button type="button" @click="removeDay(index)">Удалить день</button>
      </div>
      <div
        v-for="(exercise, exIndex) in day.exercises"
        :key="`${day.id}-${exIndex}-${exercise.exercise_id}`"
        class="prescription"
      >
        <h3>{{ names.get(exercise.exercise_id) || 'Упражнение из плана' }}</h3>
        <div class="platform-fields">
          <label
            >Подходы<input
              :value="exercise.sets"
              type="number"
              min="1"
              max="100"
              required
              @input="changeExercise(index, exIndex, 'sets', Number($event.target.value))"
          /></label>
          <label
            >Повторения от<input
              :value="exercise.reps_min"
              type="number"
              min="1"
              required
              @input="changeExercise(index, exIndex, 'reps_min', Number($event.target.value))"
          /></label>
          <label
            >Повторения до<input
              :value="exercise.reps_max"
              type="number"
              :min="exercise.reps_min"
              required
              @input="changeExercise(index, exIndex, 'reps_max', Number($event.target.value))"
          /></label>
          <label
            >Рабочий вес, кг<input
              :value="exercise.weight_kg"
              type="number"
              min="0"
              step="0.5"
              placeholder="Не задан"
              @input="
                changeExercise(
                  index,
                  exIndex,
                  'weight_kg',
                  $event.target.value === '' ? null : Number($event.target.value),
                )
              "
          /></label>
          <label
            >Отдых, сек.<input
              :value="exercise.rest_sec"
              type="number"
              min="0"
              required
              @input="changeExercise(index, exIndex, 'rest_sec', Number($event.target.value))"
          /></label>
        </div>
        <div class="day-actions">
          <button type="button" :disabled="exIndex === 0" @click="moveExercise(index, exIndex, -1)">
            Поднять упражнение</button
          ><button
            type="button"
            :disabled="exIndex === day.exercises.length - 1"
            @click="moveExercise(index, exIndex, 1)"
          >
            Опустить упражнение</button
          ><button type="button" @click="removeExercise(index, exIndex)">Убрать упражнение</button>
        </div>
      </div>
      <label
        >Добавить или заменить упражнение<select v-model="selected[day.id]">
          <option :value="null">Выберите упражнение</option>
          <option v-for="ex in exercises" :key="ex.id" :value="ex.id">{{ ex.name }}</option>
        </select></label
      >
      <div class="platform-actions">
        <button type="button" :disabled="!selected[day.id]" @click="addExercise(index)">
          Добавить упражнение</button
        ><label v-if="day.exercises.length"
          >Заменить позицию<select v-model="replacement[day.id]">
            <option :value="null">Выберите позицию</option>
            <option v-for="(_, pos) in day.exercises" :key="pos" :value="pos">{{ pos + 1 }}</option>
          </select></label
        ><button
          v-if="day.exercises.length"
          type="button"
          :disabled="!selected[day.id] || replacement[day.id] == null"
          @click="replaceExercise(index)"
        >
          Заменить
        </button>
      </div>
    </article>
    <div class="platform-actions">
      <button type="button" @click="addDay">Добавить тренировочный день</button>
    </div>
    <div class="platform-panel exercise-search">
      <label>Поиск упражнения<input v-model="query" type="search" @input="search" /></label>
      <p v-if="catalog.error.value" role="alert" class="platform-error">
        {{ catalog.error.value }}
      </p>
      <button
        v-if="catalog.nextCursor.value"
        type="button"
        :disabled="catalog.loading.value"
        @click="moreExercises"
      >
        Ещё упражнения
      </button>
    </div>
  </section>
</template>
<script setup>
import { ref, computed, onMounted, watch } from 'vue'
import { clone } from '@/utils/platform'
import { useExerciseNames } from '@/composables/useExerciseNames'
import { usePage } from '@/composables/useAsync'
const props = defineProps({ modelValue: { type: Array, required: true } })
const emit = defineEmits(['update:modelValue'])
const selected = ref({}),
  replacement = ref({}),
  query = ref(''),
  exercises = ref([])
const catalog = usePage('/api/v2/exercises', () => ({ query: query.value }))
const known = useExerciseNames()
const names = computed(
  () =>
    new Map([
      ...Object.entries(known.names.value).map(([id, name]) => [Number(id), name]),
      ...exercises.value.map((ex) => [ex.id, ex.name]),
    ]),
)
watch(
  () => props.modelValue.flatMap((day) => day.exercises.map((ex) => ex.exercise_id)).join(','),
  () => {
    const missing = props.modelValue
      .flatMap((day) => day.exercises.map((ex) => ex.exercise_id))
      .filter((id) => !names.value.has(id))
    if (missing.length) void known.load(missing)
  },
  { immediate: true },
)
function update(fn) {
  const days = clone(props.modelValue)
  fn(days)
  days.forEach((day, i) => {
    day.position = i
    day.exercises.forEach((exercise, p) => {
      exercise.position = p
    })
  })
  emit('update:modelValue', days)
}
function addDay() {
  update((days) =>
    days.push({
      id: crypto.randomUUID(),
      title: `День ${days.length + 1}`,
      position: days.length,
      exercises: [],
    }),
  )
}
function changeDay(i, key, value) {
  update((days) => {
    days[i][key] = value
  })
}
function changeExercise(i, e, key, value) {
  update((days) => {
    days[i].exercises[e][key] = value
  })
}
function moveDay(i, delta) {
  update((days) => {
    ;[days[i], days[i + delta]] = [days[i + delta], days[i]]
  })
}
function removeDay(i) {
  if (confirm('Удалить тренировочный день?')) update((days) => days.splice(i, 1))
}
function copyDay(i) {
  update((days) => {
    const day = clone(days[i])
    day.id = crypto.randomUUID()
    day.title += ' — копия'
    days.splice(i + 1, 0, day)
  })
}
function moveExercise(i, e, delta) {
  update((days) => {
    const list = days[i].exercises
    ;[list[e], list[e + delta]] = [list[e + delta], list[e]]
  })
}
function removeExercise(i, e) {
  update((days) => days[i].exercises.splice(e, 1))
}
function addExercise(i) {
  update((days) =>
    days[i].exercises.push({
      exercise_id: selected.value[props.modelValue[i].id],
      position: days[i].exercises.length,
      sets: 3,
      reps_min: 8,
      reps_max: 12,
      weight_kg: null,
      rest_sec: 90,
    }),
  )
}
function replaceExercise(i) {
  changeExercise(
    i,
    replacement.value[props.modelValue[i].id],
    'exercise_id',
    selected.value[props.modelValue[i].id],
  )
}
function mergeExercises() {
  const map = new Map(exercises.value.map((ex) => [ex.id, ex]))
  for (const ex of catalog.items.value) map.set(ex.id, ex)
  exercises.value = [...map.values()]
}
async function search() {
  await catalog.load()
  mergeExercises()
}
async function moreExercises() {
  await catalog.load(true)
  mergeExercises()
}
onMounted(search)
</script>
<style scoped>
.prescription {
  padding: 18px 0;
  border-top: 1px solid var(--sportik-border);
  margin-top: 18px;
}
.prescription h3 {
  font-size: 1rem;
  margin: 0 0 14px;
}
.day-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.day-actions button {
  font-size: 0.8rem;
  padding: 8px 10px;
}
.exercise-search {
  margin-top: 20px;
}
</style>
