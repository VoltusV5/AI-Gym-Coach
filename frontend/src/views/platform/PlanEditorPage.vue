<template>
  <platform-layout
    :title="isNew ? 'Новый тренировочный план' : draft.title || 'Редактор плана'"
    subtitle="Дни, упражнения и расписание сохраняются вместе."
  >
    <async-state :loading="load.loading.value" :error="load.error.value" @retry="fetchPlan">
      <p v-if="source" class="platform-muted">
        Личная копия<span v-if="source.title">
          программы «{{ source.title }}»<span v-if="source.trainer_name">
            · {{ source.trainer_name }}</span
          ></span
        >. Изменения этой копии не меняют авторскую программу.
      </p>
      <form @submit.prevent="save">
        <fieldset :disabled="action.busy.value" class="editor-fields">
          <section class="platform-panel">
            <label
              >Название плана<input v-model.trim="draft.title" required maxlength="200"
            /></label>
          </section>
          <training-days-editor v-model="draft.days" />
          <schedule-editor v-model="draft.schedule" :days="draft.days" />
          <p v-if="action.error.value" class="platform-error" role="alert">
            {{ action.error.value }}
          </p>
          <p v-if="action.success.value" class="platform-success" role="status">
            {{ action.success.value }}
          </p>
          <div class="platform-actions">
            <button class="primary" :disabled="action.busy.value">Сохранить план</button
            ><button
              v-if="!isNew"
              type="button"
              :disabled="action.busy.value || dirty"
              @click="activate"
            >
              {{ active ? 'План активен' : 'Сделать активным' }}</button
            ><router-link v-if="!isNew && active" :to="`/training/${route.params.id}`"
              >Начать тренировку</router-link
            ><button v-if="!isNew" type="button" :disabled="action.busy.value" @click="compare">
              Сравнить с сервером</button
            ><button v-if="!isNew" type="button" :disabled="action.busy.value" @click="archive">
              Архивировать
            </button>
          </div>
          <p v-if="dirty && !isNew" class="platform-muted">
            Сохраните правки перед активацией или началом тренировки.
          </p>
        </fieldset>
      </form>
      <share-controls
        v-if="!isNew && auth.isRegistered"
        :plan-id="String(route.params.id)"
        :days="draft.days"
        :dirty="dirty"
      />
      <section v-if="current" class="platform-panel">
        <h2>Текущие данные на сервере</h2>
        <p>{{ current.title }} · {{ current.days.length }} дней</p>
        <ul>
          <li v-for="day in current.days" :key="day.id">
            {{ day.title }}: {{ day.exercises.length }} упражнений
          </li>
        </ul>
        <button type="button" @click="acceptVersion">
          Использовать текущую версию для моих правок
        </button>
      </section>
    </async-state>
  </platform-layout>
</template>
<script setup>
import { useViewLoad } from '@/composables/useViewLoad'
import { ref, computed, watch } from 'vue'
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router'
import api, { errorMessage } from '@/api/api'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
import TrainingDaysEditor from '@/components/platform/TrainingDaysEditor.vue'
import ShareControls from '@/components/platform/ShareControls.vue'
import { useAuthStore } from '@/stores/auth'
import ScheduleEditor from '@/components/platform/ScheduleEditor.vue'
import { useRequest, useAction } from '@/composables/useAsync'
import { emptyPlan, planPayload } from '@/utils/platform'
const auth = useAuthStore(),
  route = useRoute(),
  router = useRouter(),
  load = useRequest(),
  action = useAction()
const draft = ref(emptyPlan()),
  baseline = ref(''),
  current = ref(null),
  active = ref(false),
  source = ref(null)
const isNew = computed(() => !route.params.id || route.params.id === 'new')
const dirty = computed(() => JSON.stringify(planPayload(draft.value)) !== baseline.value)
watch(
  () => draft.value.days.map((day) => day.id),
  (ids) => {
    draft.value.schedule.entries = draft.value.schedule.entries.filter((entry) =>
      ids.includes(entry.plan_day_id),
    )
  },
)
async function fetchPlan() {
  if (isNew.value) {
    baseline.value = JSON.stringify(planPayload(draft.value))
    return
  }
  const plan = await load.run(
    async () => (await api.get(`/api/v2/me/plans/${route.params.id}`)).data,
  )
  if (!plan) return
  draft.value = planPayload(plan)
  active.value = plan.is_active
  source.value = plan.source_program || (plan.source_program_version_id ? {} : null)
  baseline.value = JSON.stringify(planPayload(draft.value))
}
async function save() {
  const saved = await action.run(
    isNew.value ? '/api/v2/me/plans' : `/api/v2/me/plans/${route.params.id}`,
    planPayload(draft.value),
    { method: isNew.value ? 'post' : 'patch', idempotent: isNew.value },
  )
  if (!saved) return
  draft.value = planPayload(saved)
  baseline.value = JSON.stringify(draft.value)
  active.value = saved.is_active
  if (isNew.value) await router.replace(`/plans/${saved.id}`)
}
async function activate() {
  if (active.value) return
  const saved = await action.run(`/api/v2/me/plans/${route.params.id}/activate`, {
    version: draft.value.version,
  })
  if (saved) {
    draft.value.version = saved.version
    baseline.value = JSON.stringify(planPayload(draft.value))
    active.value = saved.is_active
  }
}
async function compare() {
  try {
    current.value = (await api.get(`/api/v2/me/plans/${route.params.id}`)).data
  } catch (cause) {
    action.error.value = errorMessage(cause)
  }
}
function acceptVersion() {
  if (confirm('Ваши правки могут заменить изменения на сервере. Продолжить после сравнения?')) {
    draft.value.version = current.value.version
    current.value = null
  }
}
async function archive() {
  if (!confirm('Архивировать план? Его нельзя будет использовать для новых тренировок.')) return
  const result = await action.run(`/api/v2/me/plans/${route.params.id}`, undefined, {
    method: 'delete',
  })
  if (result) {
    baseline.value = JSON.stringify(planPayload(draft.value))
    await router.replace('/plans')
  }
}
onBeforeRouteLeave(
  () => !dirty.value || confirm('В плане есть несохранённые изменения. Выйти из редактора?'),
)
useViewLoad(fetchPlan, () => !dirty.value)
</script>

<style scoped>
.editor-fields {
  border: 0;
  padding: 0;
  margin: 0;
  min-width: 0;
}
</style>
