<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, onBeforeRouteLeave } from 'vue-router'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import DayEditor from '@/components/ui/DayEditor.vue'
import RequestState from '@/components/ui/RequestState.vue'
import { getReference, v2 } from '@/api/v2'
import { useRead, useWrite } from '@/composables/useRequest'
import { assignDay, rebasePlan, writeDays } from '@/utils/plans'
import { useAuthStore } from '@/stores/auth'
import { stashDraft, takeDraft } from '@/utils/reauthDrafts'
import { weekdays } from '@/utils/profile'

const props = defineProps({ planId: { type: String, required: true } })
const route = useRoute()
const auth = useAuthStore()
const path = computed(() => `/me/plans/${encodeURIComponent(props.planId)}`)
const draft = ref(null)
const baseline = ref('')
const serverPlan = ref(null)
const saved = ref(false)
const { data, loading, error, status, load } = useRead(async (signal) => {
  const [plan, exercises, equipment] = await Promise.all([
    v2('GET', path.value, { signal }), getReference('/exercises'), getReference('/equipment'),
  ])
  return { plan, exercises, equipment }
}, 'План не найден или вам недоступен.')
const { saving, error: writeError, status: writeStatus, run } = useWrite()
const dirty = computed(() => draft.value && JSON.stringify(draft.value) !== baseline.value)
onBeforeRouteLeave((to) => {
  if (to.name === 'Login' && !auth.token && dirty.value) {
    stashDraft(auth.lastSubject, path.value, { draft: draft.value, baseline: baseline.value })
    return true
  }
  return !dirty.value || window.confirm('Уйти из редактора без сохранения изменений?')
})

watch(() => props.planId, () => {
  draft.value = null
  serverPlan.value = null
  load()
}, { immediate: true })
watch(data, (loaded) => {
  if (!loaded) return
  if (draft.value && writeStatus.value === 409) {
    serverPlan.value = loaded.plan
  } else {
    const previous = takeDraft(auth.accountId, path.value)
    draft.value = previous?.draft ?? structuredClone(loaded.plan)
    baseline.value = previous?.baseline ?? JSON.stringify(draft.value)
    if (previous && draft.value.version !== loaded.plan.version) {
      writeStatus.value = 409
      writeError.value = 'Версия плана изменилась, пока сессия была закрыта.'
      serverPlan.value = loaded.plan
    }
  }
})

/** Preserve schedules that reference surviving days. */
function removeScheduleDay(dayId) {
  draft.value.schedule.entries = draft.value.schedule.entries.filter((entry) => entry.plan_day_id !== dayId)
}

/** Update the day assigned to the ISO weekday. */
function schedule(weekday, dayId) {
  draft.value.schedule.entries = assignDay(draft.value.schedule.entries, weekday, dayId)
}

/** Apply only the sections edited against the previous version to the latest plan. */
function acceptLatest() {
  const rebased = rebasePlan(JSON.parse(baseline.value), draft.value, serverPlan.value)
  baseline.value = JSON.stringify(serverPlan.value)
  draft.value = rebased
  serverPlan.value = null
  writeStatus.value = 0
  writeError.value = ''
}

/** Save an aggregate atomically. A conflict leaves every editable field intact. */
function save() {
  return run(async () => {
    saved.value = false
    const { title, version, days, schedule } = draft.value
    draft.value = await v2('PATCH', path.value, { data: { title, version, days: writeDays(days), schedule } })
    baseline.value = JSON.stringify(draft.value)
    saved.value = true
  })
}
</script>

<template>
  <FeaturePage eyebrow="Конструктор" title="Редактор плана" subtitle="Настройте порядок, нагрузку и расписание тренировок." :back="`/plans/${planId}`">
    <div class="feature-stack">
      <RequestState :loading="loading" :error="error" :status="status" :retry="load" />
      <form v-if="draft && status !== 401" class="feature-stack" @submit.prevent="save">
        <fieldset class="feature-stack feature-card" :disabled="saving || loading || writeStatus === 401">
          <legend>Тренировочный план</legend>
          <label>Название плана <input v-model.trim="draft.title" maxlength="200" required /></label>
          <fieldset v-if="route.query.section !== 'schedule'" class="feature-stack">
            <legend>Тренировки</legend>
            <DayEditor v-model="draft.days" :exercises="data?.exercises ?? []" :equipment="data?.equipment ?? []" @removed-day="removeScheduleDay" />
          </fieldset>
          <fieldset class="feature-stack">
            <legend>Расписание</legend>
            <div class="feature-fields">
              <label>Дата начала<input v-model="draft.schedule.start_date" required type="date" /></label>
              <label>Часовой пояс<input v-model.trim="draft.schedule.time_zone" required placeholder="Europe/Moscow" /></label>
            </div>
            <div class="feature-fields">
              <label v-for="(day, index) in weekdays" :key="day">{{ day }}
                <select :value="draft.schedule.entries.find((entry) => entry.weekday === index + 1)?.plan_day_id ?? ''" @change="schedule(index + 1, $event.target.value)">
                  <option value="">Отдых</option>
                  <option v-for="planDay in draft.days" :key="planDay.id" :value="planDay.id">{{ planDay.title }}</option>
                </select>
              </label>
            </div>
          </fieldset>
        </fieldset>
        <div class="feature-row">
          <button class="feature-button" :disabled="saving || loading || !dirty || writeStatus === 409 || writeStatus === 401">
            {{ saving ? 'Сохраняем…' : 'Сохранить план' }}
          </button>
          <RouterLink v-if="route.query.section" class="feature-link" :to="`/plans/${planId}/edit`">Все настройки</RouterLink>
          <p v-if="saved && !dirty" class="feature-success" role="status">План сохранён</p>
        </div>
      </form>
      <RequestState :error="writeError" :status="writeStatus" />
      <section v-if="writeStatus === 409" class="feature-card">
        <p>Ваши правки остались в форме. Проверьте версию на сервере перед повторным сохранением.</p>
        <button v-if="!serverPlan" type="button" class="feature-button secondary" :disabled="loading" @click="load">Перечитать план</button>
        <div v-else class="feature-stack">
          <p>Ваши изменения будут перенесены на новую версию. Если вы и другой редактор меняли состав тренировок, сохранится ваш состав.</p>
          <RouterLink class="feature-link" :to="`/plans/${planId}`" target="_blank" rel="noopener">Открыть версию {{ serverPlan.version }} в новой вкладке</RouterLink>
          <button type="button" class="feature-button secondary" @click="acceptLatest">Применить мой ввод к версии {{ serverPlan.version }}</button>
        </div>
      </section>
    </div>
  </FeaturePage>
</template>
