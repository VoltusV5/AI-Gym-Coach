<template>
  <platform-layout
    title="Мои авторские программы"
    subtitle="Черновик редактируется отдельно от опубликованной версии."
  >
    <template #actions
      ><router-link class="platform-button primary" to="/author/programs/new"
        >Создать программу</router-link
      ></template
    >
    <async-state
      :loading="page.loading.value"
      :error="page.error.value"
      :empty="!page.items.value.length"
      @retry="page.load()"
      ><template #empty>Создайте первую программу и отправьте её на проверку.</template>
      <div class="platform-grid">
        <article v-for="item in page.items.value" :key="item.id" class="platform-card">
          <span class="platform-badge">{{ statuses[item.status] || item.status }}</span>
          <h2>{{ item.draft.title }}</h2>
          <p>Черновик и настройки: версия {{ item.version }}</p>
          <p v-if="item.published_version_id">У программы есть опубликованная версия.</p>
          <router-link :to="`/author/programs/${item.id}`">Открыть программу</router-link>
        </article>
      </div></async-state
    >
    <button v-if="page.nextCursor.value" :disabled="page.loading.value" @click="page.load(true)">
      Ещё программы
    </button>
  </platform-layout>
</template>
<script setup>
import { useViewLoad } from '@/composables/useViewLoad'

import { usePage } from '@/composables/useAsync'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const page = usePage('/api/v2/trainer/programs')
const statuses = {
  draft: 'Черновик',
  in_review: 'На проверке',
  published: 'Опубликовано',
  rejected: 'Нужны исправления',
  hidden: 'Скрыто',
}
useViewLoad(() => page.load())
</script>
