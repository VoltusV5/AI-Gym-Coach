<template>
  <platform-layout
    :title="exercise ? 'Справочник упражнений' : 'Категории программ'"
    subtitle="Изменения сохраняются с причиной в журнале действий. Использованные записи можно отключить."
  >
    <form class="platform-panel" @submit.prevent="page.load()">
      <label>Поиск по названию<input v-model="query" maxlength="200" /></label>
      <div class="platform-actions">
        <button>Найти</button><button type="button" @click="create">Добавить запись</button>
      </div>
    </form>
    <form v-if="form" class="platform-panel" @submit.prevent="save">
      <fieldset :disabled="action.busy.value" class="reference-fields">
        <h2>{{ form.id ? 'Изменить запись' : 'Новая запись' }}</h2>
        <div class="platform-fields">
          <label
            >Название<input v-model="form.name" :maxlength="exercise ? 100 : 200" required /></label
          ><template v-if="!exercise"
            ><label
              >Код категории<input
                v-model="form.slug"
                pattern="[a-z0-9]+(-[a-z0-9]+)*"
                maxlength="100"
                required /></label
            ><label
              >Порядок<input
                v-model.number="form.position"
                type="number"
                min="0"
                required /></label></template
          ><template v-else
            ><label>Группа мышц<input v-model="form.muscle_group" maxlength="100" required /></label
            ><label>Подгруппа мышц<input v-model="form.muscle_subgroup" maxlength="100" /></label
            ><label
              >Изображение, URL<input v-model="form.image_url" type="text" maxlength="100" /></label
            ><label>Видео, URL<input v-model="form.video_url" type="text" maxlength="100" /></label
          ></template>
        </div>
        <template v-if="exercise"
          ><fieldset>
            <legend>Необходимое оборудование</legend>
            <label v-for="item in equipment.data.value || []" :key="item.id" class="check-label"
              ><input v-model="form.equipment_ids" type="checkbox" :value="item.id" />{{
                item.name
              }}</label
            >
            <p v-if="equipment.error.value" role="alert">{{ equipment.error.value }}</p>
          </fieldset>
          <label class="check-label"
            ><input v-model="form.safe_for_injuries" type="checkbox" />Допустимо при ограничениях по
            здоровью</label
          ></template
        ><label class="check-label"
          ><input v-model="form.is_active" type="checkbox" />Запись активна</label
        ><label
          >Причина изменения<textarea v-model="form.reason" maxlength="4000" required />
        </label>
        <div class="platform-actions">
          <button class="primary">Сохранить запись</button
          ><button type="button" @click="form = null">Закрыть форму</button>
        </div>
      </fieldset>
    </form>
    <p v-if="action.error.value" role="alert" class="platform-error">{{ action.error.value }}</p>
    <p v-if="action.success.value" role="status">{{ action.success.value }}</p>
    <async-state
      :loading="page.loading.value"
      :error="page.error.value"
      :empty="!page.items.value.length"
      @retry="page.load()"
      ><template #empty>Записей пока нет.</template>
      <article v-for="item in page.items.value" :key="item.id" class="platform-panel">
        <h2>{{ item.name }}</h2>
        <p>{{ item.is_active ? 'Активна' : 'Отключена' }}</p>
        <button :disabled="action.busy.value" @click="edit(item)">Изменить</button>
      </article></async-state
    ><button v-if="page.nextCursor.value" :disabled="page.loading.value" @click="page.load(true)">
      Ещё записи
    </button>
  </platform-layout>
</template>
<script setup>
import { computed, ref } from 'vue'
import { useRoute, onBeforeRouteLeave } from 'vue-router'
import api from '@/api/api'
import { clone } from '@/utils/platform'
import { usePage, useRequest, useAction } from '@/composables/useAsync'
import { useViewLoad } from '@/composables/useViewLoad'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const route = useRoute(),
  exercise = computed(() => route.path.endsWith('/exercises')),
  query = ref(''),
  form = ref(null),
  baseline = ref(''),
  equipment = useRequest(),
  action = useAction()
const endpoint = computed(() => `/api/v2/admin/${exercise.value ? 'exercises' : 'categories'}`),
  page = usePage(
    () => endpoint.value,
    () => ({ query: query.value }),
  )
function edit(item) {
  if (
    form.value &&
    JSON.stringify(form.value) !== baseline.value &&
    !confirm('Отменить несохранённые изменения?')
  )
    return
  form.value = {
    ...clone(item),
    reason: '',
    muscle_subgroup: item.muscle_subgroup ?? '',
    image_url: item.image_url ?? '',
    video_url: item.video_url ?? '',
  }
  baseline.value = JSON.stringify(form.value)
}
function create() {
  edit(
    exercise.value
      ? { name: '', muscle_group: '', equipment_ids: [], safe_for_injuries: false, is_active: true }
      : { name: '', slug: '', position: 0, is_active: true },
  )
}
async function save() {
  const payload = exercise.value
    ? {
        name: form.value.name,
        muscle_group: form.value.muscle_group,
        muscle_subgroup: form.value.muscle_subgroup || null,
        image_url: form.value.image_url || null,
        video_url: form.value.video_url || null,
        equipment_ids: form.value.equipment_ids,
        safe_for_injuries: form.value.safe_for_injuries,
        is_active: form.value.is_active,
        reason: form.value.reason,
      }
    : {
        name: form.value.name,
        slug: form.value.slug,
        position: form.value.position,
        is_active: form.value.is_active,
        reason: form.value.reason,
      }
  if (!confirm('Сохранить изменение справочника?')) return
  const saved = await action.run(
    endpoint.value + (form.value.id ? `/${form.value.id}` : ''),
    payload,
    { method: form.value.id ? 'patch' : 'post' },
  )
  if (saved) {
    form.value = null
    await page.load()
  }
}
onBeforeRouteLeave(
  () =>
    !form.value ||
    JSON.stringify(form.value) === baseline.value ||
    confirm('В справочнике есть несохранённые изменения. Выйти?'),
)
useViewLoad(() => {
  page.load()
  if (exercise.value)
    equipment.run(
      async () => (await api.get('/api/v2/equipment', { params: { limit: 100 } })).data.items,
    )
})
</script>
<style scoped>
.reference-fields {
  border: 0;
  padding: 0;
  margin: 0;
  min-width: 0;
}
</style>
