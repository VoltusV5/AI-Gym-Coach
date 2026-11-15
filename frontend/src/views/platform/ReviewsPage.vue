<template>
  <platform-layout
    title="Отзывы о программе"
    subtitle="Оценки участников, завершивших необходимое число тренировок."
  >
    <template #actions
      ><router-link :to="`/programs/${route.params.id}`">К программе</router-link></template
    >
    <p v-if="page.data.value?.average != null">
      Средняя оценка: {{ page.data.value.average.toFixed(1) }} из 5 ·
      {{ page.data.value.count }} отзывов
    </p>
    <async-state
      :loading="page.loading.value"
      :error="page.error.value"
      :empty="!page.items.value.length"
      @retry="page.load()"
      ><template #empty>Опубликованных отзывов пока нет.</template>
      <article v-for="review in page.items.value" :key="review.id" class="platform-panel">
        <h2>{{ review.overall }} из 5</h2>
        <p>{{ review.text }}</p>
        <p class="platform-muted">{{ new Date(review.created_at).toLocaleDateString('ru-RU') }}</p>
        <report-form
          v-if="auth.isRegistered"
          target-type="review"
          :target-id="review.id"
        /></article
    ></async-state>
    <button v-if="page.nextCursor.value" :disabled="page.loading.value" @click="page.load(true)">
      Ещё отзывы
    </button>
    <section class="platform-panel">
      <h2>Ваш отзыв</h2>
      <p v-if="!auth.isRegistered">
        <router-link :to="{ path: '/login', query: { redirect: route.fullPath } }"
          >Войдите</router-link
        >, чтобы проверить право на отзыв.
      </p>
      <p v-else-if="eligibility.error.value" role="alert" class="platform-error">
        {{ eligibility.error.value }}
      </p>
      <template v-else-if="eligibility.data.value"
        ><p>
          Завершено тренировок: {{ eligibility.data.value.completed_workouts }} из
          {{ eligibility.data.value.required_workouts }}.
        </p>
        <p v-if="!eligibility.data.value.eligible" class="platform-muted">
          Отзыв станет доступен после необходимого числа тренировок по программе и при действующем
          доступе.
        </p>
        <p v-if="own?.status === 'hidden' || own?.status === 'in_review'" class="platform-muted">
          Ваш отзыв ожидает решения администрации и не участвует в публичном рейтинге.
        </p>
        <form v-if="eligibility.data.value.eligible" @submit.prevent="save">
          <fieldset :disabled="action.busy.value" class="review-fields">
            <div class="platform-fields">
              <label
                >Общая оценка<select v-model.number="form.overall" required>
                  <option v-for="n in 5" :key="n" :value="n">{{ n }}</option>
                </select></label
              ><label v-for="(label, key) in dimensions" :key="key"
                >{{ label
                }}<select v-model="form[key]">
                  <option value="">Без оценки</option>
                  <option v-for="n in 5" :key="n" :value="n">{{ n }}</option>
                </select></label
              >
            </div>
            <label>Ваш опыт<textarea v-model="form.text" maxlength="10000" /></label>
            <div class="platform-actions">
              <button class="primary">{{ own ? 'Сохранить отзыв' : 'Опубликовать отзыв' }}</button>
            </div>
          </fieldset>
        </form>
        <div v-if="own" class="platform-actions">
          <p>{{ own.text }}</p>
          <button type="button" :disabled="action.busy.value" @click="remove">
            Удалить свой отзыв
          </button>
        </div></template
      >
      <p v-if="action.error.value" role="alert" class="platform-error">{{ action.error.value }}</p>
      <p v-if="action.success.value" role="status" class="platform-success">
        {{ action.success.value }}
      </p>
    </section>
  </platform-layout>
</template>
<script setup>
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/api/api'
import { useAuthStore } from '@/stores/auth'
import { usePage, useRequest, useAction } from '@/composables/useAsync'
import { useViewLoad } from '@/composables/useViewLoad'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
import ReportForm from '@/components/platform/ReportForm.vue'
const route = useRoute(),
  auth = useAuthStore(),
  page = usePage(() => `/api/v2/programs/${route.params.id}/reviews`),
  eligibility = useRequest(),
  action = useAction()
const own = computed(() => eligibility.data.value?.own_review)
const dimensions = {
  difficulty: 'Соответствие сложности',
  structure: 'Структура',
  description_match: 'Соответствие описанию',
  usability: 'Удобство выполнения',
  result: 'Результат',
}
const form = ref({
  overall: 5,
  difficulty: '',
  structure: '',
  description_match: '',
  usability: '',
  result: '',
  text: '',
})
async function refresh() {
  await page.load()
  if (!auth.isRegistered) return
  const value = await eligibility.run(
    async () => (await api.get(`/api/v2/programs/${route.params.id}/review-eligibility`)).data,
  )
  if (value?.own_review) {
    const review = value.own_review
    form.value = {
      overall: review.overall,
      text: review.text,
      ...Object.fromEntries(Object.keys(dimensions).map((key) => [key, review[key] ?? ''])),
    }
  }
}
async function save() {
  const payload = Object.fromEntries(Object.entries(form.value).filter(([, value]) => value !== ''))
  if (
    await action.run(
      own.value ? `/api/v2/reviews/${own.value.id}` : `/api/v2/programs/${route.params.id}/reviews`,
      payload,
      { method: own.value ? 'patch' : 'post' },
    )
  )
    await refresh()
}
async function remove() {
  if (
    confirm('Удалить свой отзыв?') &&
    (await action.run(`/api/v2/reviews/${own.value.id}`, undefined, { method: 'delete' }))
  ) {
    form.value.text = ''
    await refresh()
  }
}
useViewLoad(refresh)
</script>
<style scoped>
.review-fields {
  border: 0;
  margin: 0;
  padding: 0;
  min-width: 0;
}
</style>
