<template>
  <ion-page>
    <ion-content class="sportik-generating-content ion-padding ion-text-center" fullscreen>
      <div
        v-if="workoutBgUrl"
        class="generating-bg"
        :style="{ backgroundImage: `url(${workoutBgUrl})` }"
        aria-hidden="true"
      ></div>
      <div class="generating-container">
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
import { ref, onMounted, onUnmounted } from 'vue'
import { IonPage, IonContent, IonSpinner, IonProgressBar } from '@ionic/vue'
import { useAuthStore } from '@/stores/auth'
import { useWorkoutPlanStore } from '@/stores/workoutPlan'
import { useRouter } from 'vue-router'
import { getWorkoutBackgroundImageUrl } from '@/utils/localImages'

const authStore = useAuthStore()
const workoutPlanStore = useWorkoutPlanStore()
const router = useRouter()

const workoutBgUrl = getWorkoutBackgroundImageUrl()

const progress = ref(0)
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

onMounted(async () => {
  // Интервал для прогресс-бара
  progressInterval = setInterval(() => {
    if (progress.value < 0.95) {
      progress.value += 0.01 + Math.random() * 0.02
    }
  }, 100)

  // Интервал для смены цитат
  quoteInterval = setInterval(() => {
    currentQuoteIndex.value = (currentQuoteIndex.value + 1) % quotes.length
  }, 2500)

  try {
    const planPayload = await authStore.generatePlan()
    workoutPlanStore.setPlanFromApi(planPayload)

    // Дождемся пока прогресс дойдет до 100% для красоты
    progress.value = 1
    setTimeout(() => {
      router.replace('/home')
    }, 500)
  } catch (err) {
    console.error('Failed to generate plan:', err)
    // Если ошибка, то вернем на шаг выбора дней
    router.replace('/training-days')
  }
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
  gap: 2rem;
}

.animation-wrapper {
  position: relative;
  width: 120px;
  height: 120px;
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
  font-size: 1.5rem;
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
  color: var(--ion-color-primary);
}

.footer-hint {
  position: absolute;
  bottom: 3rem;
  padding: 0 2rem;
  color: var(--ion-color-medium);
  font-size: 0.9rem;
}

/* Анимация перехода текста */
.fade-enter-active, .fade-leave-active {
  transition: opacity 0.5s ease;
}
.fade-enter-from, .fade-leave-to {
  opacity: 0;
}
</style>
