<template>
  <workout-chrome active-tab-key="workout">
    <div class="tools-head">
      <h1 class="sportik-title">Инструменты тренировки</h1>
      <p class="sportik-subtitle">Соберите персональную систему тренировок</p>
    </div>
    <div class="tools-grid">
      <button
        v-for="action in actions"
        :key="action.title"
        type="button"
        class="tools-action-btn"
        :class="'tools-btn--' + action.type"
        @click="onActionClick(action)"
      >
        <!-- Real Image Illustration from user assets -->
        <div class="action-illustration">
          <img
            v-if="action.imageUrl"
            :src="action.imageUrl"
            :alt="action.title"
            class="tool-img"
          />
        </div>

        <span class="tools-action-text">{{ action.title }}</span>
      </button>
    </div>
  </workout-chrome>
</template>

<script setup>
defineOptions({ name: 'WorkoutToolsPage' })

import { useRouter } from 'vue-router'
import { toastController } from '@ionic/vue'
import WorkoutChrome from '@/components/workout/WorkoutChrome.vue'
import {
  getChangePlanImageUrl,
  getAiChatImageUrl,
  getChangeParamsImageUrl
} from '@/utils/localImages'

const router = useRouter()

const actions = [
  {
    title: 'Смена плана',
    path: '/body-metrics',
    query: { regenerate: '1' },
    type: 'plan',
    imageUrl: getChangePlanImageUrl()
  },
  {
    title: 'Чат с AI тренером',
    path: '/ai-chat',
    type: 'ai',
    imageUrl: getAiChatImageUrl()
  },
  {
    title: 'Поменять параметры тела',
    path: '/body-metrics',
    type: 'params',
    imageUrl: getChangeParamsImageUrl()
  }
]

async function onActionClick(action) {
  if (action.path) {
    router.push({ path: action.path, query: action.query || {} })
  } else {
    const toast = await toastController.create({
      message: `${action.title}: функционал будет добавлен позже`,
      duration: 1500,
      color: 'medium'
    })
    await toast.present()
  }
}
</script>

<style scoped>
.tools-head {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin-bottom: 20px;
}

.tools-grid {
  width: 100%;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  align-content: stretch;
}

.tools-action-btn {
  border: none;
  border-radius: 24px;
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--sportik-brand) 12%, var(--sportik-surface)) 0%,
    var(--sportik-surface-soft) 100%
  );
  border: 1px solid var(--sportik-border);
  box-shadow: var(--sportik-shadow-md);
  color: var(--sportik-text);
  font-size: 0.95rem;
  font-weight: 700;
  text-align: center;
  padding: 20px;
  min-height: 170px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
  transition:
    transform 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    box-shadow 0.2s cubic-bezier(0.4, 0, 0.2, 1),
    border-color 0.2s ease;
}

.tools-action-btn:hover {
  transform: translateY(-2px);
  border-color: color-mix(in srgb, var(--sportik-brand) 40%, var(--sportik-border));
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.tools-action-btn:active {
  transform: translateY(0) scale(0.98);
}

.tools-btn--params {
  grid-column: span 2;
  flex-direction: row;
  align-items: center;
  justify-content: flex-start;
  gap: 24px;
  min-height: 90px;
  text-align: left;
}

.tools-btn--params .action-illustration {
  margin-bottom: 0;
  width: 52px;
  height: 52px;
}

.action-illustration {
  width: 60px;
  height: 60px;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border-radius: 12px;
  overflow: hidden;
  background: rgba(255, 255, 255, 0.03);
}

.tool-img {
  width: 100%;
  height: 100%;
  object-fit: cover;
  display: block;
}

.tools-action-text {
  line-height: 1.35;
}
</style>