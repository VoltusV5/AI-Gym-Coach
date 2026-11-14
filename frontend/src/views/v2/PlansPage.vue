<script setup>
import { onIonViewWillEnter } from '@ionic/vue'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import { v2 } from '@/api/v2'
import { usePages } from '@/composables/useRequest'
import { sourceLabels } from '@/utils/plans'

const { items, cursor, loading, error, status, reset, more, retry } = usePages((cursor, signal) =>
  v2('GET', '/me/plans', { params: { cursor }, signal }),
)
onIonViewWillEnter(reset)
</script>

<template>
  <FeaturePage eyebrow="Тренировки" title="Мои планы" subtitle="Создавайте тренировки с нуля или адаптируйте план от Спортика." tab="plans">
    <template #actions>
      <div class="feature-row">
        <RouterLink class="feature-button secondary" to="/plans/new">Создать план</RouterLink>
        <RouterLink class="feature-button secondary" to="/plans/generate">Создать AI-план</RouterLink>
      </div>
    </template>
    <div class="feature-stack">
      <div v-if="items.length" class="feature-grid">
        <article v-for="plan in items" :key="plan.id" class="feature-card">
          <div class="feature-row">
            <span class="feature-badge">{{ sourceLabels[plan.source_type] ?? plan.source_type }}</span>
            <span v-if="plan.is_active" class="feature-badge">Активный</span>
          </div>
          <h2>{{ plan.title }}</h2>
          <p class="feature-muted">Версия {{ plan.version }} · Дней в плане: {{ plan.days.length }}</p>
          <RouterLink class="feature-button" :to="`/plans/${plan.id}`">Открыть план</RouterLink>
        </article>
      </div>
      <RequestState :loading="loading" :error="error" :status="status" :retry="retry" :empty="items.length ? '' : 'Планов пока нет. Создайте свой первый план.'" />
      <button v-if="cursor && !error" type="button" class="feature-button secondary" :disabled="loading" @click="more">Показать ещё</button>
    </div>
  </FeaturePage>
</template>
