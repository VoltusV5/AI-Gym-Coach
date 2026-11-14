<template>
  <platform-layout
    title="Мои тренировочные планы"
    subtitle="Создайте свой план или измените программу, которую подготовил AI."
  >
    <template #actions
      ><router-link class="platform-button primary" to="/plans/new"
        >Создать план</router-link
      ></template
    >
    <div class="platform-actions">
      <router-link to="/saved-shares">Сохранённые ссылки</router-link
      ><router-link to="/profile">Заполнить профиль для генерации AI</router-link>
    </div>
    <async-state
      :loading="page.loading.value"
      :error="page.error.value"
      :empty="!page.items.value.length"
      @retry="page.load()"
      ><template #empty>Пока нет планов. Начните с одного тренировочного дня.</template>
      <div class="platform-grid">
        <article v-for="plan in page.items.value" :key="plan.id" class="platform-card">
          <span v-if="plan.is_active" class="platform-badge">Активный план</span>
          <h2>
            <router-link :to="`/plans/${plan.id}`">{{ plan.title }}</router-link>
          </h2>
          <p class="platform-muted">
            {{ plan.days.length }} тренировочных дней ·
            {{ sources[plan.source_type] || 'Личный план' }}
          </p>
          <router-link :to="`/plans/${plan.id}`">Открыть и изменить</router-link>
        </article>
      </div></async-state
    >
    <div v-if="page.nextCursor.value" class="platform-actions">
      <button :disabled="page.loading.value" @click="page.load(true)">Ещё планы</button>
    </div>
  </platform-layout>
</template>
<script setup>
import { useViewLoad } from '@/composables/useViewLoad'

import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
import { usePage } from '@/composables/useAsync'
const page = usePage('/api/v2/me/plans')
const sources = {
  manual: 'Создан вручную',
  ai: 'Создан с AI',
  fork: 'Личная копия программы',
  legacy_ai: 'Ранее созданный AI-план',
}
useViewLoad(() => page.load())
</script>
