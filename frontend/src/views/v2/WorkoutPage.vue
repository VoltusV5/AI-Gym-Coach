<script setup>
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import { apiStatus, getReference, v2 } from '@/api/v2'
import { useRead, useWrite } from '@/composables/useRequest'
import { useAuthStore } from '@/stores/auth'
import { completion, readTab, workoutDraft, writeTab } from '@/utils/workoutV2'

const props = defineProps({
  planId: { type: String, required: true },
  dayId: { type: String, required: true },
})
const router = useRouter()
const auth = useAuthStore()
const storageKey = computed(() => `sportik-workout:${auth.accountId}:${props.planId}:${props.dayId}`)
const draft = ref(null)
const pending = ref(null)
const pendingConflict = ref(false)
const { data, error, loading, status, load } = useRead(async (signal) => {
  const [plan, exercises] = await Promise.all([
    v2('GET', `/me/plans/${encodeURIComponent(props.planId)}`, { signal }),
    getReference('/exercises'),
  ])
  return { plan, day: plan.days.find(({ id }) => id === props.dayId), exercises }
}, 'Этот план недоступен.')
const { saving, error: submitError, status: submitStatus, run } = useWrite()
const names = computed(() => new Map((data.value?.exercises ?? []).map(({ id, name }) => [id, name])))
const hasSets = computed(() => draft.value?.slots.some((slot) => slot.sets.some((set) => set.reps > 0)))
const changedRevision = computed(() => draft.value && data.value && draft.value.plan_version !== data.value.plan.version)
watch(storageKey, (key) => {
  const previous = auth.accountId ? readTab(key) : null
  draft.value = previous?.draft ?? null
  pending.value = previous?.pending ?? null
  pendingConflict.value = previous?.conflict ?? false
  load()
}, { immediate: true })
watch(data, (loaded) => {
  if (!draft.value && loaded?.day?.exercises.length) draft.value = workoutDraft(loaded.plan, loaded.day)
})
watch([draft, pending, pendingConflict], ([current, request, conflict]) => {
  if (auth.accountId && current) writeTab(storageKey.value, { draft: current, pending: request, conflict })
}, { deep: true, flush: 'sync' })

/** Retry the exact same payload and key until the server confirms completion. */
function finish() {
  return run(async () => {
    const key = storageKey.value
    const account = auth.accountId
    pending.value ??= completion(draft.value)
    try {
      const result = await v2('POST', '/workouts/complete', {
        data: pending.value, key: pending.value.client_session_id,
      })
      writeTab(`sportik-result:${account}:${result.session_id}`, result)
      draft.value = null
      pending.value = null
      writeTab(key, null)
      await router.replace(`/workout/success/${result.session_id}`)
    } catch (err) {
      if ([400, 422].includes(apiStatus(err))) pending.value = null
      if (apiStatus(err) === 409) pendingConflict.value = true
      throw err
    }
  })
}

/** Start a distinct workout only when the user explicitly discards the conflicted attempt. */
function restart() {
  const current = data.value
  if (!current?.day?.exercises.length || !window.confirm('Начать отдельную тренировку? Ввод из конфликтующего черновика будет удалён.')) return
  draft.value = workoutDraft(current.plan, current.day)
  pending.value = null
  pendingConflict.value = false
  submitStatus.value = 0
  submitError.value = ''
}
</script>

<template>
  <FeaturePage eyebrow="Тренировка" :title="data?.day?.title ?? 'Фактические подходы'"
    subtitle="Запишите выполненные повторения и рабочий вес. Для упражнений с собственным весом оставьте 0 кг." :back="`/plans/${planId}`">
    <div class="feature-stack">
      <RequestState :loading="loading" :error="error" :status="status" :retry="load"
        :empty="data && !draft ? 'В этом дне нет упражнений. Добавьте их в редакторе плана.' : ''" />
      <form v-if="draft && auth.token" class="feature-stack" @submit.prevent="finish">
        <p v-if="changedRevision" class="feature-state">Вы начали занятие по версии {{ draft.plan_version }}. Итог сохранится с этой версией.</p>
        <fieldset v-for="slot in draft.slots" :key="slot.position" class="feature-card" :disabled="saving || !!pending">
          <legend>{{ names.get(slot.exercise_id) ?? `Упражнение №${slot.exercise_id}` }}</legend>
          <div v-for="(set, index) in slot.sets" :key="index" class="feature-fields">
            <label>Подход {{ index + 1 }} · повторы<input v-model.number="set.reps" type="number" min="1" /></label>
            <label>Подход {{ index + 1 }} · вес, кг<input v-model.number="set.weight_kg" type="number" min="0" step="any" /></label>
            <label class="feature-choice"><input v-model="set.is_warmup" type="checkbox" /> Разминка</label>
            <button type="button" class="feature-button secondary" @click="slot.sets.splice(index, 1)">Удалить подход</button>
          </div>
          <button type="button" class="feature-button secondary" @click="slot.sets.push({ reps: '', weight_kg: slot.sets.at(-1)?.weight_kg ?? 0, is_warmup: false })">Добавить подход</button>
        </fieldset>
        <button type="submit" class="feature-button" :disabled="saving || pendingConflict || (!pending && !hasSets)">
          {{ saving ? 'Сохраняем…' : pending ? 'Повторить сохранение' : 'Завершить тренировку' }}
        </button>
      </form>
      <RequestState v-if="!pendingConflict" :error="submitError" :status="submitStatus" />
      <section v-else class="feature-state feature-error" role="alert">
        <p>Эта сессия уже обработана с другими данными. Проверьте историю перед началом новой тренировки.</p>
        <div class="feature-row">
          <RouterLink class="feature-button secondary" to="/progress/history">Открыть историю</RouterLink>
          <button type="button" class="feature-button secondary" :disabled="!data?.day?.exercises.length" @click="restart">Начать новую тренировку</button>
        </div>
      </section>
    </div>
  </FeaturePage>
</template>
