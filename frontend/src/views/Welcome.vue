<template>
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
</template>

<script setup>

import { IonContent, IonButton } from '@ionic/vue'
import { useRouter } from 'vue-router'

const router = useRouter()

const startOnboarding = () => {
  router.push('/gender')
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
  min-height: 100%;
  height: 100%;
  padding: 2rem;
}
</style>