<template>
  <platform-layout
    title="Пользователи"
    subtitle="Роли и блокировки проверяются сервером. Каждое изменение сохраняется в аудит."
  >
    <template #actions><router-link to="/admin">Управление</router-link></template>
    <form class="platform-actions" @submit.prevent="page.load()">
      <label>Поиск по почте<input v-model.trim="query" type="search" /></label
      ><button class="primary">Найти</button>
    </form>
    <async-state
      :loading="page.loading.value"
      :error="page.error.value"
      :empty="!page.items.value.length"
      @retry="page.load()"
      ><div class="platform-grid">
        <article v-for="user in page.items.value" :key="user.id" class="platform-card">
          <h2>{{ user.email || `Гость ${user.id}` }}</h2>
          <p>Роли: {{ user.roles.join(', ') || 'Гость' }}</p>
          <p>{{ user.blocked ? 'Заблокирован' : 'Активен' }}</p>
          <button :disabled="user.id === auth.identity?.id" @click="edit(user)">
            Изменить доступ
          </button>
        </article>
      </div></async-state
    >
    <button v-if="page.nextCursor.value" :disabled="page.loading.value" @click="page.load(true)">
      Ещё пользователи
    </button>
    <form v-if="selected" class="platform-panel" @submit.prevent="save">
      <h2>{{ selected.email || `Пользователь ${selected.id}` }}</h2>
      <label class="check-label"><input v-model="blocked" type="checkbox" />Заблокировать</label>
      <fieldset v-if="!selected.is_anonymous">
        <legend>Роли</legend>
        <label v-for="role in ['user', 'trainer', 'admin']" :key="role" class="check-label"
          ><input v-model="roles" type="checkbox" :value="role" />{{ labels[role] }}</label
        >
      </fieldset>
      <label>Причина изменения<textarea v-model.trim="reason" required maxlength="4000" /></label>
      <div class="platform-actions">
        <button class="primary" :disabled="action.busy.value">Подтвердить изменение</button
        ><button type="button" @click="selected = null">Отмена</button>
      </div>
    </form>
    <p v-if="action.error.value" role="alert" class="platform-error">{{ action.error.value }}</p>
    <p v-if="action.success.value" role="status" class="platform-success">
      {{ action.success.value }}
    </p>
  </platform-layout>
</template>
<script setup>
import { useViewLoad } from '@/composables/useViewLoad'
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { usePage, useAction } from '@/composables/useAsync'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const auth = useAuthStore(),
  query = ref(''),
  selected = ref(null),
  roles = ref([]),
  blocked = ref(false),
  reason = ref(''),
  action = useAction()
const page = usePage('/api/v2/admin/users', () => ({ query: query.value }))
const labels = { user: 'Пользователь', trainer: 'Тренер', admin: 'Администратор' }
function edit(user) {
  selected.value = user
  roles.value = [...user.roles]
  blocked.value = user.blocked
  reason.value = ''
}
async function save() {
  if (!confirm('Изменить права этой учётной записи?')) return
  const result = await action.run(
    `/api/v2/admin/users/${selected.value.id}`,
    {
      blocked: blocked.value,
      reason: reason.value,
      ...(!selected.value.is_anonymous ? { roles: roles.value } : {}),
    },
    { method: 'patch' },
  )
  if (result) {
    selected.value = null
    await page.load()
  }
}
useViewLoad(() => page.load())
</script>
