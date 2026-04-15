<template>
  <ion-page class="tools-page">
    <ion-content class="tools-content" fullscreen>
      <div class="tools-frame ion-padding">
        <div class="tools-head">
          <h1 class="sportik-title">Инструменты тренировки</h1>
          <p class="sportik-subtitle">Соберите персональную систему тренировок</p>
        </div>
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
  --background: var(--sportik-bg);
}

.tools-frame {
  min-height: calc(100svh - 110px - env(safe-area-inset-bottom, 0px));
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.tools-head {
  display: flex;
  flex-direction: column;
  gap: 2px;
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
  border-radius: 20px;
  background: linear-gradient(
    165deg,
    color-mix(in srgb, var(--sportik-brand) 22%, var(--sportik-surface)) 0%,
    var(--sportik-surface) 100%
  );
  border: 1px solid color-mix(in srgb, var(--sportik-brand) 16%, var(--sportik-border));
  box-shadow: var(--sportik-shadow-md);
  color: var(--sportik-text);
  font-size: 1rem;
  font-weight: 600;
  text-align: left;
  padding: 16px;
  min-height: 18vh;
  display: flex;
  align-items: flex-end;
  cursor: pointer;
  transition:
    transform 0.18s ease,
    box-shadow 0.18s ease;
}

.tools-action-btn:active {
  transform: translateY(1px) scale(0.99);
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
  background: var(--sportik-surface-glass);
  box-shadow: 0 -8px 22px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(12px);
  padding-bottom: env(safe-area-inset-bottom, 0px);
}
</style>
