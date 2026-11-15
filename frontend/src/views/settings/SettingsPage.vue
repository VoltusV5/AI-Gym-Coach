<script setup>
import { ref } from 'vue'
import { alertController, onIonViewWillEnter } from '@ionic/vue'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import api from '@/api/api'
import { apiStatus, v2 } from '@/api/v2'
import { useRead } from '@/composables/useRequest'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const themeOn = ref(document.documentElement.classList.contains('dark-theme'))
const trainer = ref(false)
const admin = ref(false)
const { data: achievements, loading, error, status, load } = useRead(async (signal) => {
  const { data } = await api.get('/api/v1/achievements/get_achievements', { signal })
  return data.achievements ?? []
})

/** Probe role-protected endpoints, then show only sections the server allows. */
async function checkRoles() {
  trainer.value = false
  admin.value = false
  if (!auth.token || auth.isGuest) return
  const token = auth.token
  const [trainerAccess, adminAccess] = await Promise.all([
    v2('GET', '/trainer/programs', { params: { limit: 1 } }).then(() => true, () => false),
    v2('GET', '/admin/trainers', { params: { limit: 1 } }).then(() => true, (err) => apiStatus(err) === 404),
  ])
  if (token === auth.token) {
    trainer.value = trainerAccess
    admin.value = adminAccess
  }
}
onIonViewWillEnter(() => {
  load()
  checkRoles()
})

/** Toggle theme for the current application instance. */
function toggleTheme() {
  document.documentElement.classList.toggle('dark-theme', themeOn.value)
}

/** Show the server-supplied description for the selected achievement. */
async function showAchievement(achievement) {
  const alert = await alertController.create({
    header: achievement.title.replaceAll('_', ' '),
    message: achievement.description,
    buttons: ['Закрыть'],
  })
  await alert.present()
}
</script>

<template>
  <FeaturePage eyebrow="Аккаунт" title="Профиль и настройки" subtitle="Анкета, достижения и дополнительные инструменты." tab="settings">
    <div class="feature-stack">
      <nav class="feature-card" aria-label="Настройки и инструменты">
        <div class="feature-row">
          <RouterLink class="feature-button" to="/settings/profile">Мои данные</RouterLink>
          <RouterLink class="feature-button secondary" to="/auth/login">{{ auth.isGuest || !auth.token ? 'Войти' : 'Управление аккаунтом' }}</RouterLink>
        </div>
        <div class="feature-row">
          <RouterLink class="feature-link" to="/workout-tools">Инструменты тренировки</RouterLink>
          <RouterLink class="feature-link" to="/ai-chat">AI-чат</RouterLink>
          <RouterLink class="feature-link" to="/nutrition">Питание</RouterLink>
          <RouterLink class="feature-link" to="/notes">Заметки</RouterLink>
        </div>
        <div v-if="trainer || admin" class="feature-row">
          <RouterLink v-if="trainer" class="feature-link" to="/trainer/profile">Кабинет тренера</RouterLink>
          <template v-if="admin">
            <RouterLink class="feature-link" to="/admin/verifications">Проверка тренеров</RouterLink>
            <RouterLink class="feature-link" to="/admin/programs">Модерация программ</RouterLink>
          </template>
        </div>
      </nav>
      <section class="feature-card">
        <h2>Достижения</h2>
        <div v-if="achievements?.length" class="feature-row">
          <button v-for="achievement in achievements" :key="achievement.id" class="feature-button secondary" type="button" @click="showAchievement(achievement)">
            {{ achievement.title.replaceAll('_', ' ') }}
          </button>
        </div>
        <RequestState :loading="loading" :error="error" :status="status" :retry="load" :empty="achievements?.length === 0 ? 'Достижения появятся по мере тренировок.' : ''" />
      </section>
      <section class="feature-card">
        <h2>Оформление</h2>
        <label class="feature-choice"><input v-model="themeOn" type="checkbox" @change="toggleTheme" /> Тёмная тема</label>
      </section>
    </div>
  </FeaturePage>
</template>
