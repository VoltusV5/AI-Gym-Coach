<script setup>
import { onIonViewWillEnter } from '@ionic/vue'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import { v2 } from '@/api/v2'
import { usePages } from '@/composables/useRequest'
import { programLabels } from '@/utils/trainers'

const { items, cursor, loading, error, status, reset, more, retry } = usePages((cursor, signal) =>
  v2('GET', '/trainer/programs', { params: { cursor }, signal }),
)
onIonViewWillEnter(reset)
</script>

<template>
  <FeaturePage eyebrow="Кабинет тренера" title="Мои программы" subtitle="Готовьте черновики и отправляйте версии на модерацию." back="/trainer/profile">
    <template #actions>
      <RouterLink v-if="!loading && !error" class="feature-button secondary" to="/trainer/programs/new">Новая программа</RouterLink>
    </template>
    <div class="feature-stack">
      <div v-if="items.length" class="feature-grid">
        <article v-for="item in items" :key="item.id" class="feature-card">
          <span class="feature-badge">{{ programLabels[item.status] ?? item.status }}</span>
          <h2>{{ item.draft.title }}</h2>
          <p v-if="item.published_version_id" class="feature-muted">В каталоге доступна опубликованная версия.</p>
          <RouterLink class="feature-button" :to="`/trainer/programs/${item.id}/edit`">Открыть программу</RouterLink>
        </article>
      </div>
      <RequestState :loading="loading" :error="error" :status="status" :retry="retry" :empty="items.length ? '' : 'Авторских программ пока нет. Создайте черновик.'" />
      <button v-if="cursor && !error" class="feature-button secondary" type="button" :disabled="loading" @click="more">Показать ещё</button>
    </div>
  </FeaturePage>
</template>
