<template>
  <platform-layout title="Мои тренеры" subtitle="Авторы, за публикациями которых вы следите."
    ><async-state
      :loading="page.loading.value"
      :error="page.error.value"
      :empty="!page.items.value.length"
      @retry="page.load()"
      ><template #empty
        >Найдите автора в <router-link to="/catalog">каталоге программ</router-link> и подпишитесь
        на него.</template
      >
      <div class="platform-grid">
        <article v-for="trainer in page.items.value" :key="trainer.id" class="platform-card">
          <h2>
            <router-link :to="`/trainers/${trainer.id}`">{{ trainer.display_name }}</router-link>
          </h2>
          <span v-if="trainer.verified" class="platform-badge">Проверенный тренер</span>
          <p>{{ trainer.bio }}</p>
          <button :disabled="action.busy.value" @click="unfollow(trainer.id)">Отписаться</button>
        </article>
      </div></async-state
    ><button v-if="page.nextCursor.value" :disabled="page.loading.value" @click="page.load(true)">
      Ещё тренеры
    </button>
    <p v-if="action.error.value" role="alert" class="platform-error">
      {{ action.error.value }}
    </p></platform-layout
  >
</template>
<script setup>
import { usePage, useAction } from '@/composables/useAsync'
import { useViewLoad } from '@/composables/useViewLoad'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const page = usePage('/api/v2/me/trainers'),
  action = useAction()
async function unfollow(id) {
  if (
    await action.run(`/api/v2/trainers/${id}/follow`, undefined, {
      method: 'delete',
      message: 'Подписка отменена',
    })
  )
    await page.load()
}
useViewLoad(() => page.load())
</script>
