<template>
  <ion-app>
    <!-- Лоадер поверх всего приложения -->
    <div v-if="authStore.isLoading" class="loader-container">
      <div class="loader-card sportik-glass-card">
        <ion-spinner color="primary"></ion-spinner>
        <p>Загрузка профиля...</p>
      </div>
    </div>

    <template v-else>
      <div v-if="authStore.error" class="session-message" role="alert">
        <p>{{ authStore.error }}</p>
        <router-link to="/login">Войти</router-link>
        <button type="button" @click="authStore.initialize()">Повторить подключение</button>
      </div>
      <ion-router-outlet :key="authStore.sessionEpoch" />
    </template>
  </ion-app>
</template>

<script setup>
import { IonApp, IonSpinner, IonRouterOutlet } from '@ionic/vue'
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter, useRoute } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

onMounted(() => {
  if (route.path === '/') {
    const nextStep = authStore.nextOnboardingStep
    if (nextStep === 'Home') {
      router.replace('/home')
    }
  }
})
</script>

<style>
.session-message { padding: 12px 18px; background: var(--sportik-surface-soft); color: var(--sportik-danger); z-index: 10; position: relative; }
.session-message p { margin: 0 0 6px; }
.session-message a { margin-right: 16px; }

.loader-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background:
    radial-gradient(circle at 0% 0%, rgba(var(--sportik-brand-rgb), 0.22), transparent 36%),
    radial-gradient(circle at 100% 100%, rgba(109, 91, 255, 0.2), transparent 40%),
    var(--sportik-bg);
  z-index: 9999;
  gap: 1rem;
  text-align: center;
  font-family: var(--ion-font-family, 'Roboto', sans-serif);
}

.loader-card {
  width: min(92vw, 420px);
  padding: 1.35rem;
  text-align: center;
}

.loader-container p {
  color: var(--sportik-text-muted, #6c768f);
  margin: 0;
}

.error-container h2 {
  margin: 0;
  font-weight: 700;
}

.error-container p {
  color: var(--ion-color-medium);
  margin-bottom: 1.5rem;
  max-width: 80%;
}
</style>