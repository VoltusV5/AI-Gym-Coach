<template>
  <ion-page>
    <ion-content class="sportik-generating-content ion-padding ion-text-center" fullscreen>
      <div
        v-if="workoutBgUrl"
        class="generating-bg"
        :style="{ backgroundImage: `url(${workoutBgUrl})` }"
        aria-hidden="true"
      ></div>
        <div v-if="errorMsg" class="error-content">
          <div class="error-icon-wrapper">
            <ion-icon :icon="alertCircleOutline" color="danger"></ion-icon>
          </div>
          <h3>Ошибка генерации</h3>
          <p>{{ errorMsg }}</p>
          <ion-button expand="block" mode="ios" class="sportik-footer-btn" @click="retry">
            Попробовать ещё раз
          </ion-button>
          <ion-button fill="clear" color="medium" @click="router.replace('/profile')">
            Открыть тренировочный профиль
          </ion-button>
        </div>

        <div v-else class="generating-container">
          <!-- Анимация загрузки -->
          <div class="animation-wrapper">
            <ion-spinner name="crescent" color="primary" class="main-spinner"></ion-spinner>
            <div class="pulse-ring"></div>
          </div>

          <div class="text-content">
            <transition name="fade" mode="out-in">
              <h2 :key="currentQuoteIndex">{{ quotes[currentQuoteIndex] }}</h2>
            </transition>
            <p>Это займёт всего несколько секунд...</p>
          </div>

          <div class="progress-bar-wrapper">
            <ion-progress-bar :value="progress"></ion-progress-bar>
            <span class="progress-text">{{ Math.round(progress * 100) }}%</span>
          </div>

          <div class="footer-hint">
            <p>Учитываем ваши цели и уровень... Осталось совсем немного!</p>
          </div>
        </div>
    </ion-content>
  </ion-page>
</template>

<script setup>
import { onIonViewWillLeave } from '@ionic/vue'
import { ref, onMounted, onUnmounted } from 'vue'
import { IonPage, IonContent, IonSpinner, IonProgressBar, IonIcon, IonButton } from '@ionic/vue'
import { alertCircleOutline } from 'ionicons/icons'
import { useAuthStore } from '@/stores/auth'
import api from '@/api/api'

import { useRouter } from 'vue-router'
import { getWorkoutBackgroundImageUrl } from '@/utils/localImages'

const authStore = useAuthStore()
let generationKey = crypto.randomUUID()
let generationVersion = null
let generating = false
let alive = true
const controller = new AbortController()
onIonViewWillLeave(() => { alive = false; controller.abort() })
const router = useRouter()

const workoutBgUrl = getWorkoutBackgroundImageUrl()

const progress = ref(0)
const errorMsg = ref(null)
const currentQuoteIndex = ref(0)
const quotes = [
  'Анализируем ваши данные...',
  'Подбираем оптимальные упражнения...',
  'Генерируем ваш идеальный план тренировок...',
  'Рассчитываем нагрузку для лучшего результата...',
  'Почти готово!'
]

let progressInterval
let quoteInterval

const startGeneration = async () => {
  if (generating) return
  generating = true
  errorMsg.value = null
  progress.value = 0

  try {
    const profile = await authStore.fetchProfileV2()
    if (!alive) return
    if (!profile.session_duration_min || !profile.level || !profile.goal || !profile.training_days?.length || !profile.age || !profile.height_cm || !profile.weight_kg) {
      await router.replace('/profile')
      return
    }
    if (generationVersion !== profile.version) { generationVersion = profile.version; generationKey = crypto.randomUUID() }
    const { data } = await api.post('/api/v2/plans/generate', { profile_version: profile.version }, { headers: { 'Idempotency-Key': generationKey }, timeout: 90000, signal: controller.signal })
    if (!alive) return
    progress.value = 1
    await router.replace(`/plans/${data.plan_id}`)
  } catch (err) {
    if (!alive) return
    console.error('Failed to generate plan:', err)
    const status = err?.response?.status
    if (status === 502 || status === 504) {
      errorMsg.value = 'Сервис генерации временно недоступен (502/504). Пожалуйста, попробуйте позже.'
    } else if (status === 500) {
      errorMsg.value = 'Ошибка на сервере при создании плана. Мы уже работаем над исправлением.'
    } else if (err?.code === 'ERR_NETWORK') {
      errorMsg.value = 'Проблема с сетью. Проверьте соединение с интернетом.'
    } else {
      errorMsg.value = err?.response?.data?.error?.message || 'Не удалось создать план. Проверьте профиль, оборудование и тренировочные дни.'
    }
  } finally { generating = false }
}

const retry = () => {
  startGeneration()
}

onMounted(async () => {
  progressInterval = setInterval(() => {
    if (!errorMsg.value && progress.value < 0.95) {
      progress.value += 0.01 + Math.random() * 0.02
    }
  }, 100)
  quoteInterval = setInterval(() => {
    currentQuoteIndex.value = (currentQuoteIndex.value + 1) % quotes.length
  }, 2500)

  await startGeneration()
})

onUnmounted(() => {
  clearInterval(progressInterval)
  clearInterval(quoteInterval)
})
</script>

<style scoped>
.generating-bg {
  position: fixed;
  inset: 0;
  z-index: 0;
  background-size: cover;
  background-position: center top;
  opacity: 0.22;
  pointer-events: none;
}

.generating-container {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 100%;
  padding: 2rem;
  gap: 1.4rem;
}

.animation-wrapper {
  position: relative;
  width: 112px;
  height: 112px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.main-spinner {
  width: 80px;
  height: 80px;
  z-index: 2;
}

.pulse-ring {
  position: absolute;
  width: 100%;
  height: 100%;
  border-radius: 50%;
  background: var(--ion-color-primary);
  opacity: 0.2;
  animation: pulse 2s infinite ease-out;
}

@keyframes pulse {
  0% { transform: scale(0.5); opacity: 0.5; }
  100% { transform: scale(1.5); opacity: 0; }
}

.text-content h2 {
  font-size: 1.35rem;
  font-weight: 700;
  margin-bottom: 0.5rem;
  min-height: 4rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.text-content p {
  color: var(--ion-color-medium);
}

.progress-bar-wrapper {
  width: 100%;
  max-width: 300px;
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.progress-text {
  font-weight: 700;
  color: var(--sportik-brand);
}

.footer-hint {
  padding: 0 2rem;
  color: var(--sportik-text-muted);
  font-size: 0.9rem;
  text-align: center;
}


.fade-enter-active, .fade-leave-active {
  transition: opacity 0.5s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}


.error-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1.5rem;
  padding: 2.5rem;
  min-height: 100%;
  position: relative;
  z-index: 2;
}

.error-icon-wrapper {
  font-size: 5rem;
  margin-bottom: -0.5rem;
  animation: shake 0.5s ease-in-out;
}

.error-content h3 {
  font-size: 1.6rem;
  font-weight: 800;
  margin: 0;
  color: var(--sportik-text);
}

.error-content p {
  font-size: 1.05rem;
  line-height: 1.5;
  color: var(--sportik-text-muted);
  max-width: 320px;
  margin-bottom: 0.5rem;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-8px); }
  50% { transform: translateX(8px); }
  75% { transform: translateX(-4px); }
}
</style>