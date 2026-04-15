<template>
  <ion-page class="home-page">
    <ion-content class="home-content" fullscreen>
      <!-- Классика смартфона: сверху полоса с Аполлоном, ниже — лист на весь экран со скруглением только сверху -->
      <div class="home-scroll">
        <div class="home-frame">
          <div v-if="workoutApolloImg" class="home-apollo-strip" aria-hidden="true">
            <img class="home-apollo-strip-img" :src="workoutApolloImg" alt="" />
          </div>

          <div class="home-sheet">
            <div class="home-sheet-inner ion-padding">
              <p class="home-hero-title">Упражнения</p>
              <p v-if="splitLabel" class="home-split-hint">{{ splitLabel }}</p>

              <div class="home-stats">
                <div class="stat-pill">
                  <span class="stat-label">Упражнения</span>
                  <span class="stat-value">{{ exerciseCount || '—' }}</span>
                </div>
                <div class="stat-pill">
                  <span class="stat-label">Длительность (оценка)</span>
                  <span class="stat-value">{{ durationLabel }}</span>
                </div>
              </div>

              <p v-if="!hasPlan && mocksOn" class="home-empty-hint">
                Пройдите онбординг и сгенерируйте план — здесь появятся упражнения из ответа сервера. Пока плана
                нет, показан демо-список (моки включены).
              </p>
              <p v-else-if="!hasPlan && !mocksOn" class="home-empty-hint">
                Сгенерируйте план после онбординга. Демо-план отключён
                <code class="home-code">VITE_USE_WORKOUT_MOCKS=false</code>
                — без ответа API список пуст.
              </p>
              <p v-else-if="isDemoPlan" class="home-empty-hint">
                Сейчас показан демо-план (мок) для превью экрана тренировки. После генерации плана с сервера
                список обновится.
              </p>

              <div class="exercise-list">
                <article
                  v-for="(ex, idx) in exercises"
                  :key="`row-${idx}`"
                  class="exercise-row"
                >
                  <div class="exercise-thumb" aria-hidden="true"></div>
                  <div class="exercise-meta">
                    <p class="exercise-name">{{ ex.exercise_name }}</p>
                    <p class="exercise-time">{{ formatRowMeta(ex) }}</p>
                  </div>
                </article>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="home-footer-stack">
        <div class="home-bottom ion-padding">
          <ion-button
            class="sportik-footer-btn start-btn"
            expand="block"
            :disabled="!canStartWorkout"
            :title="startButtonTitle"
            @click="onStart"
          >
            Начать тренировку
          </ion-button>
          <ion-button fill="clear" size="small" class="logout-btn" @click="resetSession">
            Заново пройти онбординг (тест)
          </ion-button>
        </div>

        <app-tab-bar active-key="main" />
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup>
defineOptions({ name: 'HomePage' })

import { computed, onMounted } from 'vue'
import { IonPage, IonContent, IonButton } from '@ionic/vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useWorkoutPlanStore } from '@/stores/workoutPlan'
import { useWorkoutSessionStore } from '@/stores/workoutSession'
import { workoutMocksEnabled } from '@/config/workoutMocks'
import { getWorkoutBackgroundImageUrl } from '@/utils/localImages'
import AppTabBar from '@/components/navigation/AppTabBar.vue'

const router = useRouter()
const authStore = useAuthStore()
const workoutPlanStore = useWorkoutPlanStore()
const workoutSessionStore = useWorkoutSessionStore()

onMounted(() => {
  workoutPlanStore.hydrateFromStorage()
  workoutSessionStore.hydrate()
  workoutSessionStore.syncWithPlanStore(workoutPlanStore)
})

const exercises = computed(() => workoutSessionStore.homeRows)
const exerciseCount = computed(() => workoutSessionStore.slotCount)
const hasPlan = computed(() => exerciseCount.value > 0)
const mocksOn = computed(() => workoutMocksEnabled())
const isDemoPlan = computed(() => mocksOn.value && workoutSessionStore.source === 'mock')
const splitLabel = computed(() => workoutSessionStore.split || workoutPlanStore.splitLabel)

const canStartWorkout = computed(() => workoutSessionStore.slotCount > 0)
const startButtonTitle = computed(() =>
  canStartWorkout.value
    ? 'Открыть экран тренировки: упражнения, подходы, смена варианта'
    : 'Нет упражнений в сессии — сгенерируйте план или включите моки (VITE_USE_WORKOUT_MOCKS=true)'
)

const durationLabel = computed(() => {
  const n = workoutSessionStore.slotCount
  if (n === 0) return '—'
  return `${Math.max(20, Math.round(n * 6))} мин.`
})

function formatRowMeta(ex) {
  const day = ex.day_name || (ex.day != null ? `День ${ex.day}` : '')
  const w = ex.weight != null ? `${ex.weight} кг` : 'вес из плана'
  return [day, w].filter(Boolean).join(' · ')
}

const workoutApolloImg = getWorkoutBackgroundImageUrl()
const onStart = () => {
  if (!canStartWorkout.value) return
  workoutSessionStore.setCurrentIndex(0)
  router.push({ name: 'WorkoutSession' })
}

const resetSession = async () => {
  try {
    await authStore.restartSessionForTesting()
    workoutSessionStore.clear()
    await router.replace('/')
  } catch (e) {
    console.error(e)
  }
}
</script>

<style scoped>
.home-content {
  --background: var(--sportik-bg);
}

/* Отступ под фикс. футер — внутри .home-sheet, чтобы лист визуально смыкался с кнопками */
.home-scroll {
  padding-bottom: 0;
  background: transparent;
}

.home-frame {
  --home-apollo-h: clamp(124px, 31vw, 176px);
  --home-footer-pad: calc(180px + env(safe-area-inset-bottom, 0px));
  display: flex;
  flex-direction: column;
  min-height: calc(100svh - env(safe-area-inset-bottom, 0px));
  width: 100%;
}

/* Зона над скруглённым листом — голубой фон + Аполлон */
.home-apollo-strip {
  flex: 0 0 var(--home-apollo-h);
  width: 100%;
  overflow: hidden;
  position: relative;
  background: linear-gradient(145deg, var(--sportik-brand), var(--sportik-brand-2));
}

.home-apollo-strip-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center 18%;
  display: block;
}

/* Лист ниже (меньше заезд на Аполлона) — Аполлон заметнее; низ листа = место под футер */
.home-sheet {
  flex: 1 1 auto;
  width: 100%;
  margin-top: -16px;
  background: var(--sportik-surface);
  border-radius: 28px 28px 0 0;
  box-shadow: var(--sportik-shadow-lg);
  min-height: calc(100svh - var(--home-apollo-h) - env(safe-area-inset-bottom, 0px) + 8px);
  padding-bottom: calc(var(--home-footer-pad) + 4px);
  position: relative;
  z-index: 1;
}

.home-sheet-inner {
  padding-top: 1rem;
  padding-bottom: 0.25rem;
}

.home-hero-title {
  font-weight: 700;
  font-size: clamp(1.55rem, 5vw, 2.25rem);
  margin: 0 0 0.8rem;
  color: var(--sportik-text);
}

.home-split-hint {
  font-family: 'Roboto', sans-serif;
  font-size: 0.95rem;
  color: var(--sportik-text-soft);
  text-align: center;
  margin: -0.5rem 0 1rem;
}

.home-empty-hint {
  font-family: 'Roboto', sans-serif;
  font-size: 0.95rem;
  color: var(--sportik-text-muted);
  text-align: center;
  margin: 0 0 1rem;
  line-height: 1.4;
}

.home-stats {
  display: flex;
  gap: 12px;
  justify-content: center;
  flex-wrap: wrap;
  margin-bottom: 1rem;
}

.stat-pill {
  flex: 1 1 140px;
  max-width: 200px;
  background: var(--sportik-surface-soft);
  border: 1px solid var(--sportik-border);
  border-radius: var(--sportik-radius-lg);
  padding: 1rem 1.25rem;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.stat-label {
  font-family: 'Roboto', sans-serif;
  font-weight: 200;
  font-size: 0.95rem;
  color: var(--sportik-text-soft);
  opacity: 0.9;
}

.stat-value {
  font-family: 'Roboto', sans-serif;
  font-weight: 600;
  font-size: 1.35rem;
  color: var(--sportik-text);
}

.exercise-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.exercise-row {
  display: flex;
  align-items: center;
  gap: 14px;
  background: var(--sportik-surface-soft);
  border: 1px solid var(--sportik-border);
  border-radius: 14px;
  padding: 10px 12px;
  box-shadow: var(--sportik-shadow-md);
}

.exercise-thumb {
  width: 72px;
  height: 72px;
  border-radius: 10px;
  flex-shrink: 0;
  background: linear-gradient(135deg, #e0e0e0, #bdbdbd);
}

.exercise-meta {
  flex: 1;
  min-width: 0;
}

.exercise-name {
  font-family: 'Roboto', sans-serif;
  font-weight: 600;
  font-size: 1rem;
  margin: 0 0 4px;
  color: var(--sportik-text);
}

.exercise-time {
  font-family: 'Roboto', sans-serif;
  font-weight: 300;
  font-size: 0.9rem;
  margin: 0;
  color: var(--sportik-text-muted);
}

.home-footer-stack {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 10;
  display: flex;
  flex-direction: column;
  padding-bottom: env(safe-area-inset-bottom, 0px);
  /* без прозрачности — иначе сверху просвечивает мятный фон ion-content */
  background: var(--sportik-surface-glass);
  box-shadow: 0 -8px 22px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(12px);
}

.home-bottom {
  padding-top: 0.5rem;
}

.start-btn {
  margin-bottom: 0.25rem;
}

.logout-btn {
  --color: var(--sportik-text-muted);
  font-size: 0.85rem;
}

.home-code {
  font-size: 0.8em;
  padding: 2px 6px;
  border-radius: 6px;
  background: rgba(0, 0, 0, 0.06);
}

.start-btn[disabled] {
  opacity: 0.55;
}
</style>
