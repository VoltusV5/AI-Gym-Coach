<template>
  <platform-layout
    :title="program?.title || 'Тренировочная программа'"
    subtitle="Авторский подход к вашим тренировкам."
  >
    <async-state :loading="load.loading.value" :error="load.error.value" @retry="fetchProgram">
      <template v-if="program">
        <section class="platform-panel">
          <router-link :to="`/trainers/${program.trainer.id}`">{{
            program.trainer.display_name
          }}</router-link
          ><span v-if="program.trainer.verified" class="platform-badge">Проверенный тренер</span>
          <p>{{ program.description }}</p>
          <p>
            {{ goalNames[program.goal] || program.goal }} ·
            {{ levelNames[program.level] || program.level }}
          </p>
          <p>
            {{ program.duration_weeks }} недель · {{ program.sessions_per_week }} тренировок в
            неделю
          </p>
          <p>
            Оборудование:
            {{ program.equipment.map((item) => item.name).join(', ') || 'Собственный вес' }}
          </p>
          <report-form v-if="auth.isRegistered" target-type="program" :target-id="program.id" />
          <h2>
            {{ program.price_minor ? money(program.price_minor, program.currency) : 'Бесплатно' }}
          </h2>
          <div class="platform-actions">
            <router-link
              v-if="!auth.isRegistered"
              class="platform-button primary"
              :to="{ path: '/login', query: { redirect: route.fullPath } }"
              >Войти для получения доступа</router-link
            >
            <button
              v-else-if="program.access === 'preview' && program.price_minor === 0"
              class="primary"
              :disabled="action.busy.value"
              @click="grant"
            >
              Получить бесплатно
            </button>
            <button
              v-else-if="program.access === 'preview'"
              class="primary"
              :disabled="action.busy.value"
              @click="buy"
            >
              К тестовой оплате
            </button>
            <button v-else class="primary" :disabled="action.busy.value" @click="fork">
              Создать свою версию
            </button>
            <button v-if="auth.isRegistered" :disabled="action.busy.value" @click="favorite">
              {{ program.is_favorite ? 'Убрать из избранного' : 'В избранное' }}
            </button>
          </div>
          <p v-if="action.error.value" role="alert" class="platform-error">
            {{ action.error.value }}
          </p>
          <p v-if="action.success.value" role="status" class="platform-success">
            {{ action.success.value }}
          </p>
        </section>
        <section class="platform-panel">
          <h2>{{ program.access === 'preview' ? 'Превью программы' : 'Тренировочные дни' }}</h2>
          <p v-if="program.access === 'preview'" class="platform-muted">
            Состав упражнений откроется после получения доступа.
          </p>
          <article v-for="day in program.days || []" :key="day.id">
            <h3>{{ day.title }}</h3>
            <ul>
              <li v-for="exercise in day.exercises" :key="exercise.position">
                {{ exerciseNames.name(exercise.exercise_id) }} · {{ exercise.sets }} ×
                {{ exercise.reps_min }}–{{ exercise.reps_max }} · отдых {{ exercise.rest_sec }} сек.
              </li>
            </ul>
          </article>
        </section>
        <section class="platform-panel">
          <h2>Отзывы</h2>
          <router-link :to="`/programs/${program.id}/reviews`"
            >Оценки и опыт участников</router-link
          >
        </section>
      </template>
    </async-state>
  </platform-layout>
</template>
<script setup>
import { useExerciseNames } from '@/composables/useExerciseNames'
import { useViewLoad } from '@/composables/useViewLoad'
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api/api'
import { useAuthStore } from '@/stores/auth'
import { useRequest, useAction } from '@/composables/useAsync'
import { sendProgramEvent } from '@/utils/events'
import ReportForm from '@/components/platform/ReportForm.vue'
import { money, goalNames, levelNames } from '@/utils/platform'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const exerciseNames = useExerciseNames()
const route = useRoute(),
  router = useRouter(),
  auth = useAuthStore(),
  load = useRequest(),
  action = useAction()
const program = computed(() => load.data.value)
let viewed = false
async function fetchProgram() {
  const value = await load.run(
    async () => (await api.get(`/api/v2/catalog/programs/${route.params.id}`)).data,
  )
  if (value?.days?.length)
    void exerciseNames.load(
      value.days.flatMap((day) => day.exercises.map((item) => item.exercise_id)),
    )
  if (value && !viewed) {
    viewed = true
    sendProgramEvent('view', value.id, 'program')
  }
}
async function grant() {
  if (
    await action.run(`/api/v2/programs/${program.value.id}/access`, undefined, {
      message: 'Программа доступна',
    })
  )
    await fetchProgram()
}
async function favorite() {
  const add = !program.value.is_favorite
  if (
    await action.run(`/api/v2/me/favorites/${program.value.id}`, undefined, {
      method: add ? 'put' : 'delete',
      message: add ? 'Программа сохранена' : 'Программа удалена из избранного',
    })
  ) {
    sendProgramEvent(add ? 'save' : 'unsave', program.value.id, 'program')
    await fetchProgram()
  }
}
async function buy() {
  const result = await action.run(
    '/api/v2/orders',
    { program_id: program.value.id, program_version_id: program.value.published_version_id },
    { idempotent: true, message: 'Тестовый заказ создан' },
  )
  if (result) await router.push(`/checkout/${result.id}`)
}
async function fork() {
  const result = await action.run(
    `/api/v2/programs/${program.value.id}/fork`,
    {
      source_version_id: program.value.published_version_id,
      title: `${program.value.title} — моя версия`,
    },
    { idempotent: true, message: 'Личная копия создана' },
  )
  if (result) await router.push(`/plans/${result.plan_id}`)
}
useViewLoad(fetchProgram)
</script>
