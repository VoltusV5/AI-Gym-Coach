<template>
  <ion-page class="success-page">
    <ion-content class="success-content ion-padding" fullscreen>
      <div class="success-container">
        <div class="confetti-wrapper" aria-hidden="true">
          <div v-for="n in 12" :key="n" class="confetti-piece"></div>
        </div>

        <div class="hero-section">
          <div class="success-icon-wrap">
            <ion-icon :icon="checkmarkCircleOutline" class="main-icon" />
          </div>
          <h1 class="congrats-title">Тренировка завершена!</h1>
          <p class="congrats-subtitle">Отличная работа! Вы стали сильнее.</p>
          <p class="build-ver-success">Build: {{ buildId }}</p>
        </div>

        <div v-if="newAchievements.length > 0" class="achievements-section">
          <h2 class="section-title">Новые достижения! 🎊</h2>
          <div class="ach-grid">
            <div v-for="ach in newAchievements" :key="ach.id" class="ach-card">
              <div class="ach-badge">{{ getAchEmoji(ach.title) }}</div>
              <div class="ach-info">
                <div class="ach-name">{{ formatAchTitle(ach.title) }}</div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="planUpdates.length > 0" class="progress-section">
          <h2 class="section-title">Ваш прогресс (цели на следующую тренировку)</h2>
          <div class="updates-list">
            <div v-for="up in planUpdates" :key="up.exercise_id" class="update-card">
              <div class="exercise-name">{{ up.exercise_name }}</div>
              <div class="weight-progression">
                <span class="old-weight">{{ up.old_weight }} кг</span>
                <ion-icon :icon="arrowForwardOutline" class="arrow-icon" />
                <span class="new-weight">{{ up.new_weight }} кг</span>
              </div>
              <div class="delta-badge" :class="{ 'positive': up.new_weight > up.old_weight }">
                {{ up.new_weight > up.old_weight ? '+' : '' }}{{ (up.new_weight - up.old_weight).toFixed(1) }} кг
              </div>
            </div>
          </div>
        </div>

        <div v-if="highlights.length > 0" class="progress-section">
          <h2 class="section-title">Ваши достижения сегодня</h2>

          <div v-if="!showDetailedStats" class="highlights-preview">
            <div class="update-card highlight-main">
              <div class="exercise-name">{{ highlights[0].exercise_name }}</div>
              <div class="delta-badge positive" v-if="highlights[0].delta_percent > 0">
                +{{ highlights[0].delta_percent.toFixed(0) }}% {{ highlights[0].metric === 'max_weight_kg' ? 'в весе' : 'в объёме' }}
              </div>
              <div class="delta-badge" v-else>
                Зафиксировано
              </div>
            </div>
            <ion-button fill="clear" size="small" class="details-toggle-btn" @click="showDetailedStats = true">
              Подробнее о прогрессе
            </ion-button>
          </div>

          <div v-else class="updates-list detailed-stats">
            <div v-for="h in highlights" :key="h.exercise_id + h.metric" class="update-card">
              <div class="stats-content">
                <div class="exercise-name">{{ h.exercise_name }}</div>
                <p class="stats-message">{{ formatHighlightMessage(h) }}</p>
                <div class="weight-progression">
                  <span class="old-weight">{{ h.previous }} {{ getUnit(h.metric) }}</span>
                  <ion-icon :icon="arrowForwardOutline" class="arrow-icon" />
                  <span class="new-weight">{{ h.current }} {{ getUnit(h.metric) }}</span>
                </div>
              </div>
              <div class="delta-badge positive">
                +{{ h.delta_percent.toFixed(0) }}%
              </div>
            </div>
            <ion-button fill="clear" size="small" class="details-toggle-btn" @click="showDetailedStats = false">
              Свернуть
            </ion-button>
          </div>
        </div>

        <div class="footer-actions">
          <ion-button expand="block" mode="ios" class="sportik-footer-btn" @click="goHome">
            Вернуться на главную
          </ion-button>
        </div>
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { IonPage, IonContent, IonIcon, IonButton, IonSpinner } from '@ionic/vue'
import { checkmarkCircleOutline, arrowForwardOutline } from 'ionicons/icons'

const route = useRoute()
const router = useRouter()
const buildId = __SPORTIK_BUILD_ID__

const resultData = computed(() => {
  try {
    return JSON.parse(route.query.result || '{}')
  } catch (e) {
    return {}
  }
})

const planUpdates = computed(() => resultData.value.plan_updates || [])
const highlights = computed(() => resultData.value.session_highlights || [])
const newAchievements = computed(() => resultData.value.new_achievements || [])

const showDetailedStats = ref(false)

function getUnit(metric) {
  return metric === 'volume_kg' ? 'кг (объём)' : 'кг'
}

function formatHighlightMessage(h) {
  if (h.message_key === 'new_record') {
    return 'Первая фиксация результата в этом упражнении. Так держать!'
  }
  if (h.message_key === 'weight_pr') {
    const diff = (h.current - h.previous).toFixed(1)
    return `Новый максимум: +${diff} кг к прошлой тренировке!`
  }
  if (h.message_key === 'weight_up_percent') {
    return `Ваш прогресс повысился на ${h.delta_percent.toFixed(0)}%! Так держать!`
  }
  if (h.message_key === 'weight_maintained') {
    return 'Рабочий вес сохранен. Стабильность — залог успеха!'
  }
  if (h.metric === 'volume_kg') {
    return `Вы сделали больше объёма на ${h.delta_percent.toFixed(0)}%.`
  }
  return `Результат зафиксирован: ${h.current} ${getUnit(h.metric)}.`
}

function getAchEmoji(title) {
  const map = {
    first_workout: '🏆',
    five_workouts: '⭐',
    ten_workouts: '🌟',
    twenty_five_workouts: '👑',
    week_warrior: '🔥',
    consistent_month: '📅',
    comeback: '🔄',
    early_bird: '🌅',
    night_owl: '🦉',
    volume_session_5k: '💪',
    volume_session_10k: '🏋️',
    double_digit_sets: '🔢',
    profile_ready: '✅'
  }
  return map[title] || '🏅'
}

function formatAchTitle(title) {
  return title
    .split('_')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

function goHome() {
  router.replace('/home')
}
</script>

<style scoped>
.success-page {
  --background: var(--sportik-bg);
}

.success-content {
  --background: transparent;
  display: flex;
  align-items: center;
  justify-content: center;
}

.success-container {
  max-width: 480px;
  margin: 0 auto;
  padding: 2rem 1rem;
  display: flex;
  flex-direction: column;
  gap: 2rem;
  min-height: 100%;
}

.hero-section {
  text-align: center;
  margin-top: 1rem;
}

.success-icon-wrap {
  width: 100px;
  height: 100px;
  background: var(--sportik-brand-soft);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1.5rem;
  animation: scaleIn 0.6s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.main-icon {
  font-size: 4rem;
  color: var(--sportik-brand);
}

.congrats-title {
  font-size: 2rem;
  font-weight: 800;
  margin: 0 0 0.5rem;
  color: var(--sportik-text);
}

.congrats-subtitle {
  font-size: 1.1rem;
  color: var(--sportik-text-soft);
  margin: 0 0 4px;
}

.build-ver-success {
  font-size: 0.65rem;
  color: var(--sportik-text-muted);
  opacity: 0.7;
}

.section-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: var(--sportik-text);
  margin-bottom: 1rem;
  padding-left: 0.5rem;
}

.updates-list {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.update-card {
  background: var(--sportik-surface);
  border: 1px solid var(--sportik-border);
  border-radius: var(--sportik-radius-lg);
  padding: 1.25rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  position: relative;
  box-shadow: var(--sportik-shadow-md);
  animation: slideUp 0.5s ease-out backwards;
}

.update-card:nth-child(2) { animation-delay: 0.1s; }
.update-card:nth-child(3) { animation-delay: 0.2s; }
.update-card:nth-child(4) { animation-delay: 0.3s; }

.exercise-name {
  flex: 1;
  font-weight: 600;
  color: var(--sportik-text);
  font-size: 1rem;
}

.weight-progression {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  background: var(--sportik-bg);
  padding: 6px 12px;
  border-radius: 20px;
}

.old-weight {
  color: var(--sportik-text-muted);
  font-size: 0.9rem;
}

.new-weight {
  font-weight: 700;
  color: var(--sportik-text);
  font-size: 1rem;
}

.arrow-icon {
  font-size: 0.9rem;
  color: var(--sportik-brand);
}

.delta-badge {
  font-size: 0.85rem;
  font-weight: 700;
  padding: 4px 8px;
  border-radius: 8px;
  background: var(--sportik-surface-soft);
  color: var(--sportik-text-muted);
}

.delta-badge.positive {
  background: #e6f7ed;
  color: #1a8754;
}

.footer-actions {
  margin-top: auto;
  padding-bottom: 2rem;
}

.achievements-section {
  animation: bounceIn 0.8s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.ach-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
}

.ach-card {
  background: linear-gradient(135deg, var(--sportik-surface), var(--sportik-surface-soft));
  border: 2px solid var(--sportik-brand);
  border-radius: 16px;
  padding: 1rem;
  display: flex;
  align-items: center;
  gap: 1rem;
  box-shadow: 0 4px 15px rgba(var(--sportik-brand-rgb), 0.15);
}

.ach-badge {
  font-size: 2.2rem;
  width: 60px;
  height: 60px;
  background: var(--sportik-bg);
  border-radius: 14px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.ach-name {
  font-weight: 800;
  color: var(--sportik-text);
  font-size: 1.1rem;
}

@keyframes bounceIn {
  from { transform: scale(0.8); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}

.details-toggle-btn {
  --color: var(--sportik-brand);
  font-weight: 600;
  margin-top: 0.5rem;
  width: 100%;
}

.highlight-main {
  border-left: 4px solid #1a8754;
}

.stats-content {
  flex: 1;
}

.stats-message {
  margin: 4px 0 8px;
  font-size: 0.85rem;
  color: var(--sportik-text-soft);
  line-height: 1.3;
}

.detailed-stats {
  animation: fadeIn 0.4s ease-out;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(10px); }
  to { opacity: 1; transform: translateY(0); }
}


@keyframes scaleIn {
  from { transform: scale(0); opacity: 0; }
  to { transform: scale(1); opacity: 1; }
}

@keyframes slideUp {
  from { transform: translateY(20px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}


.confetti-wrapper {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 50%;
  overflow: hidden;
  pointer-events: none;
}

.confetti-piece {
  position: absolute;
  width: 10px;
  height: 10px;
  background: var(--sportik-brand);
  top: -20px;
  opacity: 0;
}

.confetti-piece:nth-child(1) { left: 10%; background: #ff7043; animation: drop 3s infinite; animation-delay: 0s; }
.confetti-piece:nth-child(2) { left: 20%; background: #42a5f5; animation: drop 3.5s infinite; animation-delay: 0.5s; }
.confetti-piece:nth-child(3) { left: 30%; background: #66bb6a; animation: drop 4s infinite; animation-delay: 1.2s; }
.confetti-piece:nth-child(4) { left: 40%; background: #ffee58; animation: drop 3s infinite; animation-delay: 0.2s; }
.confetti-piece:nth-child(5) { left: 50%; background: #ab47bc; animation: drop 3.8s infinite; animation-delay: 0.8s; }
.confetti-piece:nth-child(6) { left: 60%; background: #26a69a; animation: drop 3.2s infinite; animation-delay: 0.3s; }
.confetti-piece:nth-child(7) { left: 70%; background: #ec407a; animation: drop 4.2s infinite; animation-delay: 1.5s; }
.confetti-piece:nth-child(8) { left: 80%; background: #78909c; animation: drop 3.4s infinite; animation-delay: 0.1s; }
.confetti-piece:nth-child(9) { left: 90%; background: #ffa726; animation: drop 3.6s infinite; animation-delay: 0.7s; }

@keyframes drop {
  0% { transform: translateY(0) rotate(0); opacity: 1; }
  100% { transform: translateY(100vh) rotate(720deg); opacity: 0; }
}
</style>