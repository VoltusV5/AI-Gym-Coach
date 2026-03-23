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
      <ion-button expand="block" mode="ios" @click="authStore.initialize()">
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

  // Если мы на корневой странице, решаем куда перейти
  if (route.path === '/') {
    const nextStep = authStore.nextOnboardingStep
    console.log('Initial step detection:', nextStep)

    if (nextStep === 'Home') {
      router.replace('/home')
    } else if (nextStep !== 'Welcome') {
      const resolvedRoute = router.resolve({ name: nextStep })
      router.replace(resolvedRoute.path)
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
  background: var(--ion-background-color, #fff);
  z-index: 9999;
  gap: 1rem;
  text-align: center;
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
