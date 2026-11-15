<template>
  <platform-layout
    title="Уведомления"
    subtitle="Новые программы и обновления, которые касаются ваших тренировок."
    ><p v-if="page.data.value">Непрочитано: {{ page.data.value.unread_count }}</p>
    <async-state
      :loading="page.loading.value"
      :error="page.error.value"
      :empty="!page.items.value.length"
      @retry="page.load()"
      ><template #empty>Новых уведомлений пока нет.</template>
      <article v-for="item in page.items.value" :key="item.id" class="platform-panel">
        <span v-if="!item.read_at" class="platform-badge">Новое</span>
        <h2>{{ item.title }}</h2>
        <p>{{ item.message }}</p>
        <p class="platform-muted">{{ new Date(item.created_at).toLocaleString('ru-RU') }}</p>
        <div class="platform-actions">
          <button
            v-if="safeRoute(item.target_route)"
            class="primary"
            :disabled="action.busy.value"
            @click="open(item)"
          >
            Открыть</button
          ><button :disabled="action.busy.value" @click="mark(item, !item.read_at)">
            {{ item.read_at ? 'Отметить непрочитанным' : 'Прочитано' }}
          </button>
        </div>
      </article></async-state
    ><button v-if="page.nextCursor.value" :disabled="page.loading.value" @click="page.load(true)">
      Ещё уведомления
    </button>
    <p v-if="action.error.value" role="alert" class="platform-error">
      {{ action.error.value }}
    </p></platform-layout
  >
</template>
<script setup>
import { useRouter } from 'vue-router'
import { usePage, useAction } from '@/composables/useAsync'
import { useViewLoad } from '@/composables/useViewLoad'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const router = useRouter(),
  page = usePage('/api/v2/notifications'),
  action = useAction()
function safeRoute(value) {
  return typeof value === 'string' && value.startsWith('/') && !value.startsWith('//')
}
async function mark(item, read) {
  const result = await action.run(`/api/v2/notifications/${item.id}`, { read }, { method: 'patch' })
  if (result) await page.load()
  return result
}
async function open(item) {
  if (await mark(item, true)) await router.push(item.target_route)
}
useViewLoad(() => page.load())
</script>
