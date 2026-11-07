<script setup>
import { computed, ref, watch } from 'vue'
import { onIonViewWillEnter } from '@ionic/vue'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import { v2 } from '@/api/v2'
import { usePages, useWrite } from '@/composables/useRequest'
import { useAuthStore } from '@/stores/auth'
import { verificationLabels, programLabels } from '@/utils/trainers'

const auth = useAuthStore()
const props = defineProps({ programs: Boolean })
const path = computed(() => props.programs ? '/admin/programs' : '/admin/verifications')
const { items, cursor, loading, error, status, reset, more, retry } = usePages((cursor, signal) =>
  v2('GET', path.value, { params: { cursor }, signal }),
)
const { saving, error: decisionError, status: decisionStatus, run } = useWrite()
const selected = ref(null)
const reason = ref('')
const decision = ref('')
const saved = ref(false)
const canSee = computed(() => auth.token && ![401, 403].includes(status.value) && ![401, 403].includes(decisionStatus.value))
onIonViewWillEnter(reset)
watch(canSee, (allowed) => {
  if (!allowed) {
    items.value = []
    selected.value = null
    reason.value = ''
  }
})

/** Open a decision form for the chosen record and clear prior form state. */
function choose(item) {
  selected.value = item
  reason.value = ''
  decision.value = ''
  decisionError.value = ''
  decisionStatus.value = 0
  saved.value = false
}

/** Apply a decision to the exact record or program revision displayed to the moderator. */
function submit() {
  return run(async () => {
    const item = selected.value
    const data = { status: decision.value, reason: reason.value }
    if (props.programs) data.program_version_id = item.version_id
    const response = await v2('PATCH', `${path.value}/${item.id}${props.programs ? '/moderation' : ''}`, { data })
    items.value = items.value.map((entry) => entry.id === item.id ? { ...entry, ...response } : entry)
    selected.value = null
    saved.value = true
  })
}
</script>

<template>
  <FeaturePage eyebrow="Администрирование" :title="programs ? 'Модерация программ' : 'Проверка тренеров'" back="/settings">
    <div class="feature-stack">
      <p v-if="saved" class="feature-state feature-success" role="status">Решение сохранено</p>
      <article v-for="item in canSee ? items : []" :key="item.id" class="feature-card">
        <span class="feature-badge">{{ (programs ? programLabels : verificationLabels)[item.status] ?? item.status }}</span>
        <template v-if="programs">
          <h2>{{ item.draft.title }}</h2>
          <p class="feature-text">{{ item.draft.description }}</p>
          <p>{{ item.draft.goal }} · {{ item.draft.level }} · {{ item.draft.duration_weeks }} нед.</p>
          <details>
            <summary>Состав программы</summary>
            <div v-for="day in item.draft.days" :key="day.id" class="feature-stack">
              <strong>{{ day.title }}</strong>
              <p v-for="exercise in day.exercises" :key="exercise.position">
                Упражнение №{{ exercise.exercise_id }} · {{ exercise.sets }} × {{ exercise.reps_min }}–{{ exercise.reps_max }}
              </p>
            </div>
          </details>
        </template>
        <template v-else>
          <h2>Тренер №{{ item.trainer_id }}</h2>
          <p class="feature-text">{{ item.qualification }}</p>
          <details><summary>Документы проверки</summary><ul><li v-for="doc in item.document_refs" :key="doc">{{ doc }}</li></ul></details>
          <p v-if="item.reason" class="feature-text">Решение: {{ item.reason }}</p>
        </template>
        <button v-if="item.status === (programs ? 'in_review' : 'pending') && selected?.id !== item.id" type="button"
          class="feature-button secondary" :disabled="saving" @click="choose(item)">Рассмотреть</button>
        <form v-if="selected?.id === item.id" class="feature-stack" @submit.prevent="submit">
          <fieldset class="feature-stack" :disabled="saving">
            <legend>Решение</legend>
            <label>Статус<select v-model="decision" required>
              <option value="">Выберите решение</option><option value="approved">Подтвердить</option><option value="rejected">Отклонить</option>
            </select></label>
            <label>Причина<textarea v-model.trim="reason" required maxlength="4000" /></label>
          </fieldset>
          <div class="feature-row">
            <button class="feature-button" :disabled="saving || !decision || !reason || decisionStatus === 409">Сохранить решение</button>
            <button type="button" class="feature-button secondary" :disabled="saving" @click="selected = null">Отмена</button>
          </div>
          <RequestState :error="decisionError" :status="decisionStatus" :retry="async () => { selected = null; await reset() }" />
        </form>
      </article>
      <RequestState v-if="!canSee" :error="decisionError" :status="decisionStatus" />
      <RequestState :loading="loading" :error="error" :status="status" :retry="retry" :empty="items.length ? '' : 'Очередь пока пуста.'" />
      <button v-if="cursor && !error" class="feature-button secondary" type="button" :disabled="loading" @click="more">Показать ещё</button>
    </div>
  </FeaturePage>
</template>
