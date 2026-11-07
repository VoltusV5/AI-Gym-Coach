<script setup>
import { computed, ref } from 'vue'
import { onIonViewWillEnter } from '@ionic/vue'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import { errorText, v2 } from '@/api/v2'
import { useRead, useWrite } from '@/composables/useRequest'
import { useAuthStore } from '@/stores/auth'
import { newId } from '@/utils/plans'

const auth = useAuthStore()
const { data: profile, loading, error, status, load } = useRead((signal) => v2('GET', '/me/profile', { signal }))
const { saving, error: generationError, status: generationStatus, run } = useWrite()
const result = ref(null)
const complete = computed(() => {
  const p = profile.value
  return p?.age && p.height_cm && p.weight_kg && p.level && p.goal && p.training_days.length && p.session_duration_min
})
onIonViewWillEnter(load)

/** Describe errors specific to the generation contract without rendering upstream text. */
function generationMessage(err) {
  const code = err.response?.data?.code
  if (code === 'PROFILE_INCOMPLETE') return 'Заполните анкету, чтобы Спортик мог подобрать план.'
  if (code === 'NO_FEASIBLE_PLAN') return 'Не получилось подобрать упражнения для этих условий. Проверьте оборудование и расписание.'
  if (err.response?.status >= 500) return 'Генератор временно недоступен. Ваши планы остались в библиотеке.'
  return errorText(err)
}

/** Send a stable key for retries of the same profile version. */
function generate() {
  return run(async () => {
    const version = profile.value.version
    const storageKey = `sportik-generation:${auth.accountId}:${version}`
    const key = sessionStorage.getItem(storageKey) ?? newId()
    sessionStorage.setItem(storageKey, key)
    result.value = await v2('POST', '/plans/generate', {
      data: { profile_version: version }, key, timeout: 25000,
    })
    sessionStorage.removeItem(storageKey)
  }, generationMessage)
}
</script>

<template>
  <FeaturePage eyebrow="Мои планы · AI" title="Персональный AI-план" subtitle="Спортик учтёт анкету, доступные упражнения и ваши тренировки." back="/plans">
    <div class="feature-stack">
      <RequestState :loading="loading" :error="error" :status="status" :retry="load" />
      <section v-if="result" class="feature-card">
        <h2 role="status">План готов</h2>
        <p>Проверьте дни и упражнения в редакторе. Сделать план активным можно после просмотра.</p>
        <p v-if="result.warnings?.length" class="feature-muted">Проверьте, подходит ли предложенная нагрузка.</p>
        <RouterLink class="feature-button" :to="`/plans/${result.plan_id}/edit`">Открыть редактор</RouterLink>
      </section>
      <section v-else-if="profile" class="feature-card">
        <h2>{{ complete ? 'Анкета готова' : 'Нужны данные анкеты' }}</h2>
        <p class="feature-muted">{{ complete ? 'Генерация займёт несколько секунд.' : 'Укажите цель, уровень, параметры тела, дни и длительность занятия.' }}</p>
        <RouterLink v-if="!complete" class="feature-button" to="/settings/profile">Открыть профиль</RouterLink>
        <button v-else type="button" class="feature-button" :disabled="saving || generationStatus === 409" @click="generate">
          {{ saving ? 'Готовим план…' : generationError ? 'Повторить генерацию' : 'Сгенерировать план' }}
        </button>
        <RequestState :loading="saving" :error="generationError" :status="generationStatus" />
        <button v-if="generationStatus === 409" type="button" class="feature-button secondary" @click="load().then(() => { generationStatus = 0; generationError = '' })">Перечитать профиль</button>
      </section>
    </div>
  </FeaturePage>
</template>
