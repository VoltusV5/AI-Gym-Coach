<template>
  <workout-chrome :active-tab-key="chromeTabKey" :show-apollo="false">
    <ion-header class="ion-no-border session-header">
      <ion-toolbar class="session-toolbar">
        <ion-buttons slot="start">
          <ion-back-button default-href="/home" text="" style="--color: #ffffff;"></ion-back-button>
        </ion-buttons>
      </ion-toolbar>
    </ion-header>

    <div v-if="!hasSlots" class="empty-msg">
      <p>Нет упражнений. Вернитесь на главную и сформируйте план.</p>
      <ion-button @click="router.replace('/home')">На главную</ion-button>
    </div>

    <template v-else>
      <div class="dots-row" aria-label="Прогресс по упражнениям">
        <button
          v-for="(st, i) in statusList"
          :key="i"
          type="button"
          class="dot"
          :class="{
            'dot--done': st.complete,
            'dot--current': st.isCurrent && !st.complete,
            'dot--current-done': st.isCurrent && st.complete,
            'dot--todo': !st.complete && !st.isCurrent
          }"
          :title="`Упражнение ${i + 1}`"
          @click="session.setCurrentIndex(i)"
        ></button>
      </div>

      <div class="exercise-layout">
        <div class="exercise-main">
          <h1 class="ex-name">{{ exercise?.name }}</h1>
          <p class="ex-desc">{{ exercise?.description }}</p>

          <!-- Demo animação do exercício -->
          <div class="video-demo-wrap" aria-label="Демонстрация упражнения">
            <img
              v-if="exerciseVideoUrl"
              :src="exerciseVideoUrl"
              :alt="exercise?.name"
              class="video-demo-media"
            />
            <div v-else class="video-placeholder">
              <span>Демонстрация упражнения</span>
            </div>
          </div>


          <div class="sets-block">
            <p class="sets-label">Подходы</p>
            <div v-for="(_, si) in currentSlot.sets" :key="si" class="set-row">
              <span class="set-num">{{ si + 1 }}</span>
              <ion-input
                :model-value="String(currentSlot?.sets?.[si]?.weightKg ?? '')"
                class="set-input"
                inputmode="decimal"
                placeholder="Вес, кг"
                @ion-input="onSetInput(si, 'weightKg', $event.detail.value)"
              />
              <ion-input
                :model-value="String(currentSlot?.sets?.[si]?.reps ?? '')"
                class="set-input"
                inputmode="numeric"
                placeholder="Повторы"
                @ion-input="onSetInput(si, 'reps', $event.detail.value)"
              />
            </div>
          </div>

          <div class="nav-row">
            <ion-button fill="clear" class="nav-arrow-btn" :disabled="currentIndex <= 0" @click="session.goPrev">
              <ion-icon slot="icon-only" :icon="arrowBack"></ion-icon>
            </ion-button>
            <div class="nav-center-wrap">
              <ion-button
                v-if="canSwapExercise"
                fill="clear"
                class="swap-btn"
                @click="goChangeExercise"
              >
                <ion-icon :icon="swapHorizontalOutline" slot="start" />
                Поменять упражнение
              </ion-button>
              <p v-else class="swap-hint">Для этого слота в плане только одна вариация.</p>
            </div>
            <ion-button
              fill="clear"
              class="nav-arrow-btn"
              :disabled="currentIndex >= slotCount - 1"
              @click="session.goNext"
            >
              <ion-icon slot="icon-only" :icon="arrowForward"></ion-icon>
            </ion-button>
          </div>

          <ion-button
            expand="block"
            class="finish-workout-btn"
            :disabled="isSubmitting"
            @click="finishWorkout"
          >
            <ion-spinner v-if="isSubmitting" name="crescent" slot="start"></ion-spinner>
            Завершить тренировку
          </ion-button>
        </div>
      </div>
    </template>

    <template #footer>
      <!-- Footer content if needed later -->
    </template>
  </workout-chrome>
</template>

<script setup>
defineOptions({ name: 'WorkoutExerciseScreen' })

import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  IonHeader,
  IonToolbar,
  IonTitle,
  IonButtons,
  IonBackButton,
  IonButton,
  IonInput,
  IonIcon,
  toastController
} from '@ionic/vue'
import { arrowBack, arrowForward, swapHorizontalOutline } from 'ionicons/icons'
import WorkoutChrome from '@/components/workout/WorkoutChrome.vue'
import { useWorkoutSessionStore } from '@/stores/workoutSession'
import { staticUrl } from '@/api/api'

const route = useRoute()
const router = useRouter()
const session = useWorkoutSessionStore()
const isSubmitting = ref(false)


const chromeTabKey = computed(() => (route.query.context === 'home' ? 'main' : 'workout'))
const sessionBackHref = computed(() => (route.query.context === 'home' ? '/home' : '/workout-tools'))

async function finishWorkout() {
  if (isSubmitting.value) return
  if (!session.isReadyForComplete) {
    alert('Пожалуйста, заполните хотя бы один подход в любом упражнении.')
    return
  }

  isSubmitting.value = true
  try {
    const result = await session.submitCompleteWorkout()
    router.replace({
      name: 'WorkoutSuccess',
      query: { result: JSON.stringify(result) }
    })
  } catch (error) {
    console.error('Failed to finish workout:', error)
    const msg = error?.response?.data?.message || error?.message || 'неизвестная ошибка'
    alert('Не удалось завершить тренировку: ' + msg)
  } finally {
    isSubmitting.value = false
  }
}

onMounted(() => {
  session.hydrate()
  if (session.slots.length === 0) {
    router.replace('/home')
  }
})

const hasSlots = computed(() => session.slots.length > 0)
const slotCount = computed(() => session.slots.length)
const currentIndex = computed(() => session.currentIndex)
const currentSlot = computed(() => session.currentSlot)
const exercise = computed(() => session.selectedExercise)
const statusList = computed(() => session.slotStatusList)
const exerciseVideoUrl = computed(() => {
  const vUrl = exercise.value?.videoUrl
  return vUrl ? staticUrl(vUrl) : null
})

const canSwapExercise = computed(() => (currentSlot.value?.alternatives?.length ?? 0) > 1)

function onSetInput(setIndex, field, value) {
  session.updateSet(setIndex, field, value)
}

function goChangeExercise() {
  if (!canSwapExercise.value) return
  router.push({
    name: 'WorkoutAlternatives',
    params: { slotIndex: String(session.currentIndex) },
    query: route.query.context === 'home' ? { context: 'home' } : {}
  })
}
</script>

<style scoped>
.session-header {
  box-shadow: none;
}

.session-toolbar {
  --background: transparent;
  --border-width: 0;
}

.session-title {
  font-weight: 600;
  font-size: 1rem;
  color: var(--sportik-text);
}

.empty-msg {
  text-align: center;
  padding: 2rem 0;
  color: var(--sportik-text-muted);
}

.exercise-layout {
  display: flex;
  justify-content: center;
  align-items: flex-start;
  width: 100%;
}

.exercise-main {
  flex: 1;
  max-width: 500px;
  min-width: 0;
}

.dots-row {
  display: flex;
  flex-direction: row;
  justify-content: center;
  flex-wrap: wrap;
  gap: 12px;
  width: 100%;
  padding: 0 0 20px;
}

.ex-name {
  font-weight: 700;
  font-size: 1.35rem;
  margin: 0 0 0.5rem;
  color: var(--sportik-text);
  line-height: 1.25;
}

.ex-desc {
  font-size: 0.92rem;
  color: var(--sportik-text-muted);
  margin: 0 0 1rem;
  line-height: 1.45;
}

.video-demo-wrap {
  width: 100%;
  aspect-ratio: 4 / 3;
  border-radius: 16px;
  overflow: hidden;
  margin-bottom: 1rem;
  background: linear-gradient(
    145deg,
    color-mix(in srgb, var(--sportik-brand) 20%, var(--sportik-surface)) 0%,
    color-mix(in srgb, var(--sportik-brand-2) 12%, var(--sportik-surface-soft)) 100%
  );
  display: flex;
  align-items: center;
  justify-content: center;
}

.video-demo-media {
  width: 100%;
  height: 100%;
  object-fit: contain;
  display: block;
}

.video-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.reps-hint {
  font-size: 0.88rem;
  color: var(--sportik-text-soft);
  margin: 0 0 1.25rem;
  line-height: 1.45;
  background: var(--sportik-surface-soft);
  border: 1px solid var(--sportik-border);
  padding: 12px 14px;
  border-radius: 14px;
}

.sets-block {
  margin-bottom: 1.25rem;
}

.sets-label {
  font-weight: 600;
  font-size: 0.95rem;
  margin: 0 0 0.5rem;
  color: var(--sportik-text);
}

.set-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 10px;
}

.set-num {
  flex: 0 0 22px;
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--sportik-text-muted);
  text-align: center;
}

.set-input {
  flex: 1;
  --background: var(--sportik-surface-soft);
  --padding-start: 12px;
  --padding-end: 12px;
  border-radius: 12px;
  border: 1px solid var(--sportik-border);
  min-height: 44px;
}

.nav-row {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.nav-arrow-btn {
  --color: var(--sportik-text);
  --padding-start: 8px;
  --padding-end: 8px;
  margin: 0 !important;
  flex: 0 0 auto;
}

.nav-center-wrap {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-width: 0;
}

.nav-btn {
  margin: 0 !important;
  width: 100%;
  max-width: 240px;
  font-size: 0.9rem;
}

.nav-btn--primary {
  order: 2;
}

.nav-row .nav-btn:first-child {
  order: 1;
}

.nav-row .nav-btn:last-child {
  order: 3;
}

.nav-btn--primary {
  --background: var(--ion-color-primary);
  --color: var(--ion-color-primary-contrast);
}

.finish-temp-btn {
  margin-top: 1rem;
  font-size: 0.82rem;
}

.swap-hint {
  font-size: 0.82rem;
  color: var(--sportik-text-muted);
  text-align: center;
  margin: 0;
  line-height: 1.35;
}

.dot {
  width: 14px;
  height: 14px;
  border-radius: 50%;
  border: none;
  padding: 0;
  cursor: pointer;
  background: #bdbdbd;
  transition:
    transform 0.15s ease,
    box-shadow 0.15s ease;
}

.dot--todo {
  background: #bdbdbd;
}

.dot--done {
  background: #2dd36f;
}

.dot--current {
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--sportik-brand) 50%, transparent);
  background: #bdbdbd;
  transform: scale(1.15);
}

.dot--current-done {
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--sportik-brand) 50%, transparent);
  background: #2dd36f;
  transform: scale(1.15);
}

@media (max-width: 420px) {
  .nav-btn {
    font-size: 0.82rem;
  }
}

.finish-workout-btn {
  margin-top: 1.5rem;
  --background: var(--ion-color-primary);
  --color: #ffffff;
  --border-radius: 12px;
  font-weight: 700;
  width: 100%;
  height: 54px;
  font-size: 1.1rem;
}

.finish-container {
  background: var(--sportik-surface-glass);
  backdrop-filter: blur(12px);
  border-top: 1px solid var(--sportik-border);
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.05);
}

.finish-btn {
  margin: 0;
  --box-shadow: none;
}
</style>