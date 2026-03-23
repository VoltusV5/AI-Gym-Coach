<template>
  <ion-app>
    <!-- Лоадер поверх всего приложения -->
    <div v-if="authStore.isLoading" class="loader-container">
      <ion-spinner color="primary"></ion-spinner>
      <p>Загрузка профиля...</p>
    </div>

    <div v-else-if="authStore.error" class="loader-container error-container ion-padding">
      <ion-icon :icon="alertCircleOutline" color="danger" style="font-size: 64px;"></ion-icon>
      <h2>Ошибка подключения</h2>
      <p>{{ authStore.error }}</p>
      <ion-button class="sportik-footer-btn" expand="block" mode="ios" @click="authStore.initialize()">
        Повторить
      </ion-button>
    </div>

    <ion-router-outlet />
  </ion-app>
</template>

<script setup>
import { IonApp, IonSpinner, IonIcon, IonButton, IonRouterOutlet } from '@ionic/vue'
import { alertCircleOutline } from 'ionicons/icons'
import { onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter, useRoute } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

onMounted(async () => {
  // Выполняем логику запуска приложения
  await authStore.initialize()

  // Старт всегда с Welcome (как в макете): с `/` не уводим на шаги онбординга.
  // На главную — только если профиль уже полностью заполнен.
  if (route.path === '/') {
    const nextStep = authStore.nextOnboardingStep
    if (nextStep === 'Home') {
      router.replace('/home')
    }
  }
})
</script>

<style>
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
  background: linear-gradient(
    165deg,
    #ffffff 0%,
    var(--sportik-mint-soft, #d4f5ec) 55%,
    var(--sportik-mint, #aef3e5) 100%
  );
  z-index: 9999;
  gap: 1rem;
  text-align: center;
  font-family: var(--ion-font-family, 'Roboto', sans-serif);
}

.loader-container p {
  color: var(--sportik-text-muted, rgba(0, 0, 0, 0.6));
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
