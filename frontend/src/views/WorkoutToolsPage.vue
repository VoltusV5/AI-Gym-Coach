<template>
  <ion-page class="tools-page">
    <ion-content class="tools-content" fullscreen>
      <div class="tools-frame ion-padding">
        <div class="tools-grid">
          <button
            v-for="action in actions"
            :key="action"
            type="button"
            class="tools-action-btn"
            @click="showSoon(action)"
          >
            <span class="tools-action-text">{{ action }}</span>
          </button>
        </div>
      </div>

      <div class="tools-footer-stack">
        <app-tab-bar active-key="workout" />
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup>
defineOptions({ name: 'WorkoutToolsPage' })

import { IonPage, IonContent, toastController } from '@ionic/vue'
import AppTabBar from '@/components/navigation/AppTabBar.vue'

const actions = [
  'Смена плана',
  'Составление тренировки',
  'Расписание тренировок',
  'Чат с AI тренером',
  'Поменять параметры тела',
  'Собственные тренировки'
]

async function showSoon(title) {
  const toast = await toastController.create({
    message: `${title}: функционал будет добавлен позже`,
    duration: 1500,
    color: 'medium'
  })
  await toast.present()
}
</script>

<style scoped>
.tools-content {
  --background: var(--sportik-cream);
}

.tools-frame {
  min-height: calc(100svh - 110px - env(safe-area-inset-bottom, 0px));
  display: flex;
  align-items: stretch;
}

.tools-grid {
  width: 100%;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  align-content: stretch;
}

.tools-action-btn {
  border: none;
  border-radius: 18px;
  background: linear-gradient(165deg, #c6f8ff 0%, #a8f0f8 100%);
  box-shadow: 0 8px 20px rgba(0, 0, 0, 0.08);
  color: var(--sportik-text);
  font-family: 'Roboto', sans-serif;
  font-size: 1rem;
  font-weight: 600;
  text-align: left;
  padding: 16px 14px;
  min-height: 19vh;
  display: flex;
  align-items: flex-end;
  cursor: pointer;
}

.tools-action-text {
  line-height: 1.25;
}

.tools-footer-stack {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 10;
  background: var(--sportik-cream);
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.06);
  padding-bottom: env(safe-area-inset-bottom, 0px);
}
</style>
