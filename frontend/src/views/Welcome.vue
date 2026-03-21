<template>
  <ion-page>
    <ion-content
      fullscreen
      class="ion-padding"
    >
      <div class="welcome-container">
        <h1>Добро пожаловать в Спортик!</h1>
        <p>Создай свой план тренировок за 2 минуты</p>

        <ion-button expand="block" color="primary" @click="startOnboarding">
          Начать
        </ion-button>
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup>
import { IonPage, IonContent, IonButton } from '@ionic/vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const startOnboarding = () => {
  const step = authStore.nextOnboardingStep
  if (step === 'Welcome') {
    router.push('/gender')
  } else if (step === 'Home') {
    router.push('/home')
  } else {
    const resolvedRoute = router.resolve({ name: step })
    router.push(resolvedRoute.path)
  }
}
</script>

<style scoped>
ion-content {
  --background: var(--ion-color-light);
}

.welcome-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  height: 100%;
  padding: 2.5rem;
  gap: 1.5rem;
}

h1 {
  font-size: 2rem;
  font-weight: 800;
  margin-bottom: 0.5rem;
  color: var(--ion-color-dark);
}

p {
  font-size: 1.1rem;
  color: var(--ion-color-medium);
  margin-bottom: 2rem;
}

ion-button {
  width: 100%;
  max-width: 320px;
  --border-radius: 12px;
  --padding-top: 1.25rem;
  --padding-bottom: 1.25rem;
  font-weight: 700;
}
</style>
