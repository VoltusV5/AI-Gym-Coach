<script setup>
import { computed, ref, watch } from 'vue'
import { useRouter, onBeforeRouteLeave } from 'vue-router'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import DayEditor from '@/components/ui/DayEditor.vue'
import RequestState from '@/components/ui/RequestState.vue'
import { getReference, v2 } from '@/api/v2'
import { useRead, useWrite } from '@/composables/useRequest'
import { newDay, writeDays } from '@/utils/plans'
import { levels } from '@/utils/profile'
import { useAuthStore } from '@/stores/auth'
import { stashDraft, takeDraft } from '@/utils/reauthDrafts'
import { priceMinor, programLabels } from '@/utils/trainers'

const props = defineProps({ programId: { type: String, default: '' } })
const router = useRouter()
const auth = useAuthStore()
const path = computed(() => props.programId ? `/trainer/programs/${encodeURIComponent(props.programId)}` : '/trainer/programs')
const draft = ref(null)
const program = ref(null)
const baseline = ref('')
const saved = ref(false)
const { data, loading, error, status, load } = useRead(async (signal) => {
  // Query a role-protected endpoint before opening a create form.
  const record = await v2('GET', path.value, { signal, params: props.programId ? undefined : { limit: 1 } })
  const [exercises, equipment] = await Promise.all([getReference('/exercises'), getReference('/equipment')])
  return { record: props.programId ? record : null, exercises, equipment }
}, 'Программа не найдена или вам недоступна.')
const { saving, error: writeError, status: writeStatus, run } = useWrite()
const editable = computed(() => !program.value || ['draft', 'rejected'].includes(program.value.status))
const dirty = computed(() => Boolean(draft.value && JSON.stringify(draft.value) !== baseline.value))
const missingEquipment = computed(() => {
  const d = draft.value
  if (!d || !data.value) return []
  const exercises = new Map(data.value.exercises.map((item) => [item.id, item]))
  const required = new Set(d.days.flatMap((day) =>
    day.exercises.flatMap((slot) => exercises.get(slot.exercise_id)?.equipment_ids ?? []),
  ))
  return data.value.equipment.filter((item) => required.has(item.id) && !d.equipment_ids.includes(item.id))
})
const canPublish = computed(() => draft.value?.days.length && draft.value.days.every((day) => day.exercises.length) &&
  !missingEquipment.value.length)

/** Initialize a create form or copy the current server draft into the editor. */
function apply(record) {
  program.value = record
  const base = record?.draft ?? {
    title: '', description: '', goal: '', level: 'beginner', duration_weeks: 4, sessions_per_week: 3,
    equipment_ids: [], category_ids: [], price_minor: 0, currency: 'RUB', allow_public_remix: false, days: [newDay('День 1')],
  }
  draft.value = { ...structuredClone(base), price: String(base.price_minor / 100) }
  baseline.value = JSON.stringify(draft.value)
}
watch(() => props.programId, () => {
  draft.value = null
  load()
}, { immediate: true })
watch(data, (loaded) => {
  if (!loaded || (dirty.value && writeStatus.value === 409)) return
  apply(loaded.record)
  const previous = takeDraft(auth.accountId, path.value)
  if (previous) {
    draft.value = previous.draft
    baseline.value = previous.baseline
    if (previous.version !== loaded.record?.version) {
      writeStatus.value = 409
      writeError.value = 'Версия программы изменилась, пока сессия была закрыта.'
    }
  }
})
onBeforeRouteLeave((to) => {
  if (to.name === 'Login' && !auth.token && dirty.value) {
    stashDraft(auth.lastSubject, path.value, { draft: draft.value, baseline: baseline.value, version: program.value?.version })
    return true
  }
  return !dirty.value || window.confirm('Уйти без сохранения черновика?')
})

/** Persist only writable draft fields, preserving category IDs and currency supplied by the API. */
function save() {
  return run(async () => {
    saved.value = false
    const d = draft.value
    const body = {
      title: d.title, description: d.description, goal: d.goal, level: d.level, duration_weeks: d.duration_weeks,
      sessions_per_week: d.sessions_per_week, equipment_ids: d.equipment_ids, category_ids: d.category_ids,
      price_minor: priceMinor(d.price), currency: d.currency, allow_public_remix: d.allow_public_remix, days: writeDays(d.days),
    }
    if (program.value) body.version = program.value.version
    const record = await v2(program.value ? 'PATCH' : 'POST', path.value, { data: body })
    apply(record)
    saved.value = true
    if (!props.programId) await router.replace(`/trainer/programs/${record.id}/edit`)
  })
}

/** Submit a saved version for moderation, or create an editable successor to a publication. */
function workflow(action) {
  return run(async () => {
    const data = action === 'publish' ? { version: program.value.version } : undefined
    await v2('POST', path.value + '/' + action, { data })
    await load()
  })
}
</script>

<template>
  <FeaturePage eyebrow="Кабинет тренера" :title="programId ? 'Авторская программа' : 'Новая программа'" subtitle="Подготовьте описание, условия и расписание упражнений для участников." back="/trainer/programs">
    <div class="feature-stack">
      <RequestState :loading="loading" :error="error" :status="status" :retry="load" />
      <template v-if="draft">
        <section v-if="program" class="feature-card">
          <span class="feature-badge" role="status">{{ programLabels[program.status] }}</span>
          <p v-if="program.status === 'rejected'" class="feature-text">{{ program.moderation_reason || 'Модератор отклонил эту версию. Отредактируйте черновик и отправьте повторно.' }}</p>
          <p v-if="program.status === 'published'">Эта версия доступна в каталоге. Для изменений создайте новую версию.</p>
          <p v-if="program.status === 'in_review'">Версия на модерации. Её можно будет изменить после решения.</p>
          <button v-if="program.status === 'published'" type="button" class="feature-button" :disabled="saving" @click="workflow('versions')">Создать новую версию</button>
          <button v-if="program.status === 'in_review'" type="button" class="feature-button secondary" :disabled="saving" @click="load">Обновить статус</button>
        </section>

        <form class="feature-stack" @submit.prevent="save">
          <fieldset class="feature-card" :disabled="!editable || saving || loading">
            <legend>Содержание программы</legend>
            <label>Название программы<input v-model.trim="draft.title" required maxlength="200" /></label>
            <label>Описание<textarea v-model.trim="draft.description" maxlength="50000" /></label>
            <div class="feature-fields">
              <label>Цель<input v-model.trim="draft.goal" required maxlength="200" /></label>
              <label>Уровень<select v-model="draft.level">
                <option v-for="(label, key) in levels" :key="key" :value="key">{{ label }}</option>
              </select></label>
              <label>Недель<input v-model.number="draft.duration_weeks" type="number" min="1" required /></label>
              <label>Занятий в неделю<input v-model.number="draft.sessions_per_week" type="number" min="1" max="7" required /></label>
              <label>Цена, {{ draft.currency }} (0 — бесплатно)<input v-model.trim="draft.price" inputmode="decimal" pattern="[0-9]{1,12}([.,][0-9]{1,2})?" required /></label>
            </div>
            <fieldset>
              <legend>Оборудование</legend>
              <div class="feature-row">
                <label v-for="item in data?.equipment ?? []" :key="item.id" class="feature-choice">
                  <input v-model="draft.equipment_ids" type="checkbox" :value="item.id" />{{ item.name }}
                </label>
              </div>
            </fieldset>
            <label class="feature-choice"><input v-model="draft.allow_public_remix" type="checkbox" /> Разрешить публичные переработки</label>
            <DayEditor v-model="draft.days" :exercises="data?.exercises ?? []" :equipment="data?.equipment ?? []" />
            <div v-if="editable && missingEquipment.length" class="feature-state">
              <p>Для выбранных упражнений требуется: {{ missingEquipment.map((item) => item.name).join(', ') }}.</p>
              <button type="button" class="feature-button secondary"
                @click="draft.equipment_ids.push(...missingEquipment.map((item) => item.id))">Добавить необходимое оборудование</button>
            </div>
          </fieldset>
          <div v-if="editable" class="feature-row">
            <button class="feature-button" :disabled="saving || loading || !dirty || writeStatus === 409">{{ saving ? 'Сохраняем…' : 'Сохранить черновик' }}</button>
            <button v-if="program && canPublish && !dirty" type="button" class="feature-button secondary" :disabled="saving || loading || writeStatus === 409" @click="workflow('publish')">
              Отправить на модерацию
            </button>
            <p v-if="saved && !dirty" class="feature-success" role="status">Черновик сохранён</p>
          </div>
        </form>
      </template>
      <RequestState :error="writeError" :status="writeStatus" />
      <section v-if="writeStatus === 409" class="feature-card">
        <p>Сервер изменил версию. Ваш ввод остался в форме. Скопируйте нужные изменения перед загрузкой версии с сервера.</p>
        <button type="button" class="feature-button secondary" :disabled="loading" @click="writeStatus = 0; writeError = ''; load()">Загрузить версию с сервера</button>
      </section>
    </div>
  </FeaturePage>
</template>
