<template>
  <platform-layout
    title="Избранные программы"
    subtitle="Вернитесь к программам, которые вы сохранили."
  >
    <p v-if="action.error.value" role="alert" class="platform-error">{{ action.error.value }}</p>
    <async-state
      :loading="page.loading.value"
      :error="page.error.value"
      :empty="!page.items.value.length"
      @retry="page.load()"
      ><template #empty
        >Пока ничего не сохранено.
        <router-link to="/catalog">Откройте каталог</router-link>.</template
      >
      <div class="platform-grid">
        <program-card v-for="program in page.items.value" :key="program.id" :program="program"
          ><button :disabled="action.busy.value" @click="remove(program.id)">
            Убрать из избранного
          </button></program-card
        >
      </div></async-state
    >
    <button v-if="page.nextCursor.value" :disabled="page.loading.value" @click="page.load(true)">
      Ещё программы
    </button>
  </platform-layout>
</template>
<script setup>
import { useViewLoad } from '@/composables/useViewLoad'

import { usePage, useAction } from '@/composables/useAsync'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
import ProgramCard from '@/components/platform/ProgramCard.vue'
const page = usePage('/api/v2/me/favorites'),
  action = useAction()
async function remove(id) {
  const result = await action.run(`/api/v2/me/favorites/${id}`, undefined, { method: 'delete' })
  if (result) await page.load()
}
useViewLoad(() => page.load())
</script>
