<template>
  <platform-layout
    :title="share?.title || 'Программа по ссылке'"
    subtitle="Снимок программы, которым с вами поделились."
  >
    <async-state :loading="load.loading.value" :error="load.error.value" @retry="fetchShare"
      ><template v-if="share"
        ><section class="platform-panel">
          <p>Автор: {{ share.author.display_name }}</p>
          <p v-if="share.requires_access">
            По этой ссылке доступно только превью. Полное содержимое зависит от доступа к источнику
            и разрешения автора.
          </p>
          <router-link v-if="share.source_program_id" :to="`/programs/${share.source_program_id}`"
            >Открыть исходную программу</router-link
          >
          <div class="platform-actions">
            <router-link
              v-if="!auth.isRegistered"
              class="platform-button primary"
              :to="{ path: '/login', query: { redirect: route.fullPath } }"
              >Войти для сохранения</router-link
            ><template v-else
              ><button v-if="route.params.token" :disabled="action.busy.value" @click="save">
                Сохранить ссылку</button
              ><button v-if="share.can_revoke" :disabled="action.busy.value" @click="revoke">
                Отозвать ссылку</button
              ><button
                v-if="share.can_fork"
                class="primary"
                :disabled="action.busy.value"
                @click="fork"
              >
                Создать свою версию
              </button></template
            >
          </div>
          <p v-if="action.error.value" role="alert" class="platform-error">
            {{ action.error.value }}
          </p>
          <p v-if="action.success.value" role="status" class="platform-success">
            {{ action.success.value }}
          </p>
        </section>
        <section v-for="day in share.days" :key="day.id" class="platform-panel">
          <h2>{{ day.title }}</h2>
          <ul>
            <li v-for="exercise in day.exercises" :key="exercise.position">
              {{ exerciseNames.name(exercise.exercise_id) }} · {{ exercise.sets }} ×
              {{ exercise.reps_min }}–{{ exercise.reps_max }} · отдых {{ exercise.rest_sec }} сек.
            </li>
          </ul>
        </section></template
      ></async-state
    >
  </platform-layout>
</template>
<script setup>
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import api from '@/api/api'
import { useAuthStore } from '@/stores/auth'
import { useAction, useRequest } from '@/composables/useAsync'
import { useExerciseNames } from '@/composables/useExerciseNames'
import { useViewLoad } from '@/composables/useViewLoad'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const exerciseNames = useExerciseNames()
const route = useRoute(),
  router = useRouter(),
  auth = useAuthStore(),
  load = useRequest(),
  action = useAction(),
  share = computed(() => load.data.value)
const endpoint = computed(() =>
  route.params.token
    ? `/api/v2/shares/${route.params.token}`
    : `/api/v2/me/saved-shares/${route.params.id}`,
)
async function fetchShare() {
  const value = await load.run(async () => (await api.get(endpoint.value)).data)
  if (value?.days?.length)
    void exerciseNames.load(
      value.days.flatMap((day) => day.exercises.map((item) => item.exercise_id)),
    )
}
async function revoke() {
  if (
    confirm('Отозвать свою ссылку?') &&
    (await action.run(`/api/v2/shares/${share.value.id}`, undefined, { method: 'delete' }))
  )
    await fetchShare()
}
async function save() {
  const saved = await action.run(`${endpoint.value}/save`, undefined, {
    message: 'Ссылка сохранена в вашей библиотеке',
  })
  if (saved) await router.push(`/saved-shares/${saved.library_item_id}`)
}
async function fork() {
  const value = await action.run(`${endpoint.value}/fork`, undefined, {
    idempotent: true,
    message: 'Личная версия создана',
  })
  if (value) await router.push(`/plans/${value.plan_id}`)
}
useViewLoad(fetchShare)
</script>
