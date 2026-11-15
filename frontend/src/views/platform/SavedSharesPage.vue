<template>
  <platform-layout
    title="Сохранённые ссылки"
    subtitle="Ссылки на планы и тренировочные дни. Доступ зависит от владельца ссылки и исходной программы."
  >
    <async-state
      :loading="page.loading.value"
      :error="page.error.value"
      :empty="!page.items.value.length"
      @retry="page.load()"
      ><template #empty>Пока нет сохранённых ссылок.</template>
      <article v-for="item in page.items.value" :key="item.id" class="platform-panel">
        <h2>{{ item.title }}</h2>
        <p v-if="!item.available">Ссылка отозвана или срок её действия закончился.</p>
        <router-link v-else :to="`/saved-shares/${item.id}`">Открыть</router-link>
        <div class="platform-actions">
          <button :disabled="action.busy.value" @click="remove(item.id)">
            Удалить из сохранённых
          </button>
        </div>
      </article></async-state
    ><button v-if="page.nextCursor.value" :disabled="page.loading.value" @click="page.load(true)">
      Ещё ссылки
    </button>
    <p v-if="action.error.value" role="alert">{{ action.error.value }}</p>
  </platform-layout>
</template>
<script setup>
import { usePage, useAction } from '@/composables/useAsync'
import { useViewLoad } from '@/composables/useViewLoad'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const page = usePage('/api/v2/me/saved-shares'),
  action = useAction()
async function remove(id) {
  if (await action.run(`/api/v2/me/saved-shares/${id}`, undefined, { method: 'delete' }))
    await page.load()
}
useViewLoad(() => page.load())
</script>
