<template>
  <platform-layout
    :title="plan?.title || 'Тренировка'"
    subtitle="Запишите фактически выполненные подходы. Разминочные подходы не входят в рабочий объём."
  >
    <async-state :loading="load.loading.value" :error="load.error.value" @retry="refresh">
      <template v-if="result"
        ><section class="platform-panel">
          <h2>Тренировка сохранена</h2>
          <p>
            Рабочий объём: {{ result.metrics.volume_kg.toLocaleString('ru-RU') }} кг · подходов:
            {{ result.metrics.working_sets }} ·
            {{ Math.round(result.metrics.duration_sec / 60) }} мин.
          </p>
          <p v-for="(item, index) in result.highlights" :key="index">{{ item.message }}</p>
          <router-link to="/progress">Открыть прогресс</router-link>
          <div class="platform-actions">
            <button @click="nextTraining">Начать следующее занятие</button>
          </div>
        </section></template
      >
      <template v-else-if="plan || draft"
        ><p v-if="names.error.value" role="alert">{{ names.error.value }}</p>
        <template v-if="!draft"
          ><section class="platform-panel">
            <h2>Выберите тренировочный день</h2>
            <p v-if="!plan.days.length">В этом плане пока нет дней. Добавьте их в редакторе.</p>
            <div class="platform-actions">
              <button
                v-for="day in plan.days"
                :key="day.id"
                :disabled="!day.exercises.length"
                @click="start(day)"
              >
                {{ day.title }}
              </button>
            </div>
            <router-link :to="`/plans/${plan.id}`">Изменить план</router-link>
          </section></template
        >
        <form v-else @submit.prevent="complete">
          <p v-if="draft.pending" class="platform-muted">
            Результат подготовлен к отправке. Повторная попытка отправит те же данные и не создаст
            вторую тренировку.
          </p>
          <fieldset :disabled="action.busy.value || Boolean(draft.pending)" class="workout-fields">
            <section v-for="slot in draft.input.slots" :key="slot.position" class="platform-panel">
              <h2>{{ names.name(slot.exercise_id) }}</h2>
              <div v-for="(set, index) in slot.sets" :key="index" class="platform-fields">
                <label
                  >Подход {{ index + 1 }} · повторения<input
                    v-model.number="set.reps"
                    type="number"
                    min="1"
                    max="10000"
                    required /></label
                ><label
                  >Вес, кг<input
                    v-model.number="set.weight_kg"
                    type="number"
                    min="0"
                    max="10000"
                    step="0.1"
                    required /></label
                ><label class="check-label"
                  ><input v-model="set.is_warmup" type="checkbox" />Разминка</label
                ><button
                  type="button"
                  :disabled="slot.sets.length === 1"
                  @click="slot.sets.splice(index, 1)"
                >
                  Убрать подход
                </button>
              </div>
              <button
                type="button"
                @click="slot.sets.push({ reps: 8, weight_kg: 0, is_warmup: false })"
              >
                Добавить подход
              </button>
            </section>
          </fieldset>
          <p v-if="action.error.value" role="alert" class="platform-error">
            {{ action.error.value }}
          </p>
          <div class="platform-actions">
            <button class="primary" :disabled="action.busy.value">
              {{
                draft.pending ? 'Повторить отправку результата' : 'Завершить и сохранить тренировку'
              }}</button
            ><button
              v-if="!draft.pending"
              type="button"
              :disabled="action.busy.value"
              @click="discard"
            >
              Отменить тренировку
            </button>
          </div>
        </form></template
      >
    </async-state>
  </platform-layout>
</template>
<script setup>
import { ref, computed, watch } from 'vue'
import { useRoute, onBeforeRouteLeave } from 'vue-router'
import api from '@/api/api'
import { useAuthStore } from '@/stores/auth'
import { useRequest, useAction } from '@/composables/useAsync'
import { useViewLoad } from '@/composables/useViewLoad'
import { useExerciseNames } from '@/composables/useExerciseNames'
import {
  createWorkoutDraft,
  freezeWorkout,
  readWorkoutDraft,
  storeWorkoutDraft,
  removeWorkoutDraft,
} from '@/utils/workoutDraft'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const route = useRoute(),
  auth = useAuthStore(),
  load = useRequest(),
  action = useAction(),
  names = useExerciseNames()
const plan = computed(() => load.data.value),
  draft = ref(null),
  result = ref(null)
watch(
  draft,
  (value) => {
    if (value && value.account_id === auth.identity?.id && auth.token) storeWorkoutDraft(value)
  },
  { deep: true },
)
async function refresh() {
  result.value = null
  const existing =
    draft.value?.input.plan_id === route.params.id
      ? draft.value
      : readWorkoutDraft(auth.identity?.id, route.params.id)
  if (existing) {
    draft.value = existing
    load.error.value = ''
    await names.load(existing.input.slots.map((slot) => slot.exercise_id))
    return
  }
  draft.value = null
  const value = await load.run(
    async () => (await api.get(`/api/v2/me/plans/${route.params.id}`)).data,
  )
  if (value)
    await names.load(value.days.flatMap((day) => day.exercises.map((item) => item.exercise_id)))
}
function nextTraining() {
  result.value = null
  void refresh()
}

function start(day) {
  const previous = readWorkoutDraft(auth.identity?.id, plan.value.id)
  draft.value = previous || createWorkoutDraft(plan.value, day, auth.identity?.id)
}
function discard() {
  if (confirm('Отменить занятие без сохранения результата?')) {
    removeWorkoutDraft(auth.identity?.id, draft.value.input.plan_id)
    draft.value = null
  }
}
async function complete() {
  freezeWorkout(draft.value)
  storeWorkoutDraft(draft.value)
  const value = await action.run('/api/v2/workouts/complete', draft.value.pending, {
    idempotent: true,
    message: 'Тренировка сохранена',
  })
  if (value) {
    result.value = value
    removeWorkoutDraft(auth.identity?.id, draft.value.input.plan_id)
    draft.value = null
  }
}
onBeforeRouteLeave(
  () =>
    !draft.value ||
    confirm('Тренировка ещё не сохранена на сервере. Черновик останется в этой вкладке. Выйти?'),
)
useViewLoad(refresh)
</script>
<style scoped>
.workout-fields {
  border: 0;
  margin: 0;
  padding: 0;
  min-width: 0;
}
</style>
