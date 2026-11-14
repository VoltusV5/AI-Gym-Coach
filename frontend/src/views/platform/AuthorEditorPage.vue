<template>
  <platform-layout
    :title="isNew ? 'Новая авторская программа' : form.title || 'Программа тренера'"
    subtitle="После публикации изменения выпускаются новой версией."
  >
    <async-state :loading="load.loading.value" :error="load.error.value" @retry="fetchProgram">
      <p v-if="moderationReason" class="platform-muted">
        Решение администрации: {{ moderationReason }}
      </p>
      <p v-if="status" class="platform-badge">{{ statuses[status] || status }}</p>
      <p v-if="published" class="platform-muted">
        Опубликованная версия сохранена отдельно. Редактирование нового черновика её не меняет.
      </p>
      <form @submit.prevent="save">
        <fieldset :disabled="!editable || action.busy.value" class="author-fields">
          <section class="platform-panel">
            <div class="platform-fields">
              <label>Название<input v-model.trim="form.title" required maxlength="200" /></label
              ><label
                >Цель<select v-model="form.goal" required>
                  <option v-for="(label, key) in goalNames" :key="key" :value="key">
                    {{ label }}
                  </option>
                </select></label
              ><label
                >Уровень<select v-model="form.level" required>
                  <option v-for="(label, key) in levelNames" :key="key" :value="key">
                    {{ label }}
                  </option>
                </select></label
              ><label
                >Длительность, нед.<input
                  v-model.number="form.duration_weeks"
                  type="number"
                  min="1"
                  required /></label
              ><label
                >Занятий в неделю<input
                  v-model.number="form.sessions_per_week"
                  type="number"
                  min="1"
                  max="7"
                  required /></label
              ><label
                >Цена, руб.<input v-model="rubles" type="number" min="0" step="0.01" required
              /></label>
            </div>
            <label>Описание<textarea v-model="form.description" required maxlength="10000" /></label
            ><label class="check-label"
              ><input v-model="form.allow_public_remix" type="checkbox" />Разрешить публичное
              распространение личных копий</label
            >
          </section>
          <section class="platform-panel">
            <h2>Оборудование</h2>
            <label v-for="item in equipment" :key="item.id" class="check-label"
              ><input v-model="form.equipment_ids" type="checkbox" :value="item.id" />{{
                item.name
              }}</label
            >
          </section>
          <section class="platform-panel">
            <h2>Категории</h2>
            <p v-if="!categories.length" class="platform-muted">
              В каталоге пока нет активных категорий.
            </p>
            <label v-for="item in categories" :key="item.id" class="check-label"
              ><input v-model="form.category_ids" type="checkbox" :value="item.id" />{{
                item.name
              }}</label
            >
          </section>
          <training-days-editor v-model="form.days" />
          <div class="platform-actions"><button class="primary">Сохранить черновик</button></div>
        </fieldset>
      </form>
      <p v-if="action.error.value" class="platform-error" role="alert">{{ action.error.value }}</p>
      <p v-if="action.success.value" class="platform-success" role="status">
        {{ action.success.value }}
      </p>
      <div v-if="!isNew" class="platform-actions">
        <button v-if="editable" :disabled="dirty || action.busy.value" @click="publish">
          Отправить на публикацию</button
        ><button
          v-if="status === 'published' || status === 'hidden'"
          :disabled="action.busy.value"
          @click="newVersion"
        >
          Создать новую версию</button
        ><router-link v-if="published" :to="`/programs/${route.params.id}`"
          >Открыть опубликованную программу</router-link
        >
      </div>
      <p v-if="status === 'in_review'" class="platform-muted">
        Программа проверяется администрацией. Правки доступны после решения.
      </p>
    </async-state>
  </platform-layout>
</template>
<script setup>
import { useViewLoad } from '@/composables/useViewLoad'
import { ref, computed } from 'vue'
import { useRoute, useRouter, onBeforeRouteLeave } from 'vue-router'
import api from '@/api/api'
import { useAction, useRequest } from '@/composables/useAsync'
import { goalNames, levelNames, clone } from '@/utils/platform'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
import TrainingDaysEditor from '@/components/platform/TrainingDaysEditor.vue'
const route = useRoute(),
  router = useRouter(),
  load = useRequest(),
  action = useAction()
const form = ref({
  title: '',
  description: '',
  goal: 'muscle_gain',
  level: 'beginner',
  duration_weeks: 4,
  sessions_per_week: 3,
  equipment_ids: [],
  category_ids: [],
  price_minor: 0,
  currency: 'RUB',
  allow_public_remix: false,
  days: [],
})
const rubles = ref('0'),
  equipment = ref([]),
  categories = ref([]),
  status = ref('draft'),
  moderationReason = ref(null),
  published = ref(null),
  baseline = ref('')
const statuses = {
  draft: 'Черновик',
  in_review: 'На проверке',
  published: 'Опубликовано',
  rejected: 'Нужны исправления',
  hidden: 'Скрыто',
}
const isNew = computed(() => !route.params.id || route.params.id === 'new'),
  editable = computed(() => ['draft', 'rejected'].includes(status.value))
const dirty = computed(() => JSON.stringify([form.value, rubles.value]) !== baseline.value)
function minor(value) {
  const match = String(value).match(/^(\d+)(?:[.,](\d{1,2}))?$/)
  return match ? Number(match[1]) * 100 + Number((match[2] || '').padEnd(2, '0')) : null
}
async function fetchProgram() {
  await load.run(async () => {
    equipment.value = (await api.get('/api/v2/equipment', { params: { limit: 100 } })).data.items
    categories.value = (
      await api.get('/api/v2/catalog/categories', { params: { limit: 100 } })
    ).data.items
    if (!isNew.value) {
      const { data } = await api.get(`/api/v2/trainer/programs/${route.params.id}`)
      form.value = { ...clone(data.draft), version: data.version }
      status.value = data.status
      moderationReason.value = data.moderation_reason
      published.value = data.published_version_id
      rubles.value = (data.draft.price_minor / 100).toFixed(2)
    }
    baseline.value = JSON.stringify([form.value, rubles.value])
    return true
  })
}
async function save() {
  const amount = minor(rubles.value)
  if (amount == null || !Number.isSafeInteger(amount)) {
    action.error.value = 'Укажите стоимость с точностью до копеек.'
    return
  }
  const payload = { ...clone(form.value), price_minor: amount }
  const value = await action.run(
    isNew.value ? '/api/v2/trainer/programs' : `/api/v2/trainer/programs/${route.params.id}`,
    payload,
    { method: isNew.value ? 'post' : 'patch' },
  )
  if (value) {
    form.value = { ...value.draft, version: value.version }
    status.value = value.status
    baseline.value = JSON.stringify([form.value, rubles.value])
    if (isNew.value) await router.replace(`/author/programs/${value.id}`)
  }
}
async function publish() {
  if (!confirm('Отправить сохранённую версию программы на проверку?')) return
  const value = await action.run(
    `/api/v2/trainer/programs/${route.params.id}/publish`,
    { version: form.value.version },
    { message: 'Программа отправлена на проверку' },
  )
  if (value) await fetchProgram()
}
async function newVersion() {
  if (
    await action.run(`/api/v2/trainer/programs/${route.params.id}/versions`, undefined, {
      message: 'Новый черновик создан',
    })
  )
    await fetchProgram()
}
onBeforeRouteLeave(
  () => !dirty.value || confirm('В черновике есть несохранённые изменения. Выйти?'),
)
useViewLoad(fetchProgram, () => !dirty.value)
</script>
<style scoped>
.author-fields {
  border: 0;
  padding: 0;
  margin: 16px 0;
  min-width: 0;
}
</style>
