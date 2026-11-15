<template>
  <platform-layout
    :title="reports ? 'Жалобы и обращения' : 'Модерация отзывов'"
    subtitle="Каждое решение требует причины и сохраняется в журнале действий."
  >
    <async-state
      :loading="page.loading.value"
      :error="page.error.value"
      :empty="!page.items.value.length"
      @retry="page.load()"
      ><template #empty>Обращений пока нет.</template>
      <article v-for="item in page.items.value" :key="item.id" class="platform-panel">
        <h2>{{ reports ? targetNames[item.target_type] : `${item.overall} из 5` }}</h2>
        <p>{{ reports ? item.reason : item.text }}</p>
        <p>Статус: {{ statusNames[item.status] || item.status }}</p>
        <section v-if="reports && item.target_type === 'review'" class="platform-panel">
          <h3>Отзыв, на который пожаловались</h3>
          <template v-if="item.target_review"
            ><p>
              {{ item.target_review.overall }} из 5 ·
              {{ statusNames[item.target_review.status] || item.target_review.status }}
            </p>
            <blockquote>{{ item.target_review.text }}</blockquote>
            <router-link :to="`/programs/${item.target_review.program_id}/reviews`"
              >Отзывы этой программы</router-link
            ></template
          >
          <p v-else>Отзыв больше недоступен.</p>
        </section>
        <p v-if="item.resolution">Решение: {{ item.resolution }}</p>
        <router-link
          v-if="reports && item.target_type === 'program'"
          :to="`/programs/${item.target_id}`"
          >Открыть программу</router-link
        ><router-link
          v-if="reports && item.target_type === 'trainer'"
          :to="`/trainers/${item.target_id}`"
          >Открыть тренера</router-link
        ><router-link v-if="!reports" :to="`/programs/${item.program_id}/reviews`"
          >Открыть отзывы программы</router-link
        >
        <form
          v-if="!reports || !['resolved', 'dismissed'].includes(item.status)"
          @submit.prevent="decide(item)"
        >
          <fieldset :disabled="action.busy.value" class="moderation-fields">
            <label
              >Решение<select v-model="decisions[item.id]" required>
                <option disabled value="">Выберите решение</option>
                <template v-if="reports"
                  ><option value="in_progress">Взять в работу</option>
                  <option value="resolved">Решено</option>
                  <option value="dismissed">Отклонить жалобу</option></template
                ><template v-else
                  ><option value="published">Опубликовать</option>
                  <option value="hidden">Скрыть</option></template
                >
              </select></label
            ><label>Причина<textarea v-model="reasons[item.id]" maxlength="4000" required /></label>
            <div class="platform-actions"><button class="primary">Подтвердить решение</button></div>
          </fieldset>
        </form>
      </article></async-state
    ><button v-if="page.nextCursor.value" :disabled="page.loading.value" @click="page.load(true)">
      Ещё записи
    </button>
    <p v-if="action.error.value" role="alert" class="platform-error">{{ action.error.value }}</p>
    <p v-if="action.success.value" role="status">{{ action.success.value }}</p>
  </platform-layout>
</template>
<script setup>
import { computed, reactive } from 'vue'
import { useRoute } from 'vue-router'
import { usePage, useAction } from '@/composables/useAsync'
import { useViewLoad } from '@/composables/useViewLoad'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const route = useRoute(),
  reports = computed(() => route.path.endsWith('/reports')),
  page = usePage(() => `/api/v2/admin/${reports.value ? 'reports' : 'reviews'}`),
  action = useAction(),
  reasons = reactive({}),
  decisions = reactive({})
const targetNames = {
    program: 'Жалоба на программу',
    trainer: 'Жалоба на тренера',
    review: 'Жалоба на отзыв',
  },
  statusNames = {
    open: 'Открыта',
    in_progress: 'В работе',
    resolved: 'Решена',
    dismissed: 'Отклонена',
    published: 'Опубликован',
    in_review: 'Ожидает проверки',
    hidden: 'Скрыт',
  }
async function decide(item) {
  if (!confirm('Сохранить решение модерации?')) return
  const path = reports.value
    ? `/api/v2/admin/reports/${item.id}`
    : `/api/v2/admin/reviews/${item.id}/moderation`
  if (
    await action.run(
      path,
      { status: decisions[item.id], reason: reasons[item.id] },
      { method: 'patch' },
    )
  ) {
    delete reasons[item.id]
    delete decisions[item.id]
    await page.load()
  }
}
useViewLoad(() => page.load())
</script>
<style scoped>
.moderation-fields {
  border: 0;
  margin: 18px 0 0;
  padding: 0;
  min-width: 0;
}
</style>
