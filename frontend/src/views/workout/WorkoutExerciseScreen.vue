<template>
  <workout-chrome>
    <ion-header class="ion-no-border session-header">
      <ion-toolbar class="session-toolbar">
        <ion-buttons slot="start">
          <ion-back-button default-href="/home" text="" color="dark"></ion-back-button>
        </ion-buttons>
        <ion-title class="session-title">Тренировка</ion-title>
      </ion-toolbar>
    </ion-header>

    <div v-if="!hasSlots" class="empty-msg">
      <p>Нет упражнений. Вернитесь на главную и сформируйте план.</p>
      <ion-button @click="router.replace('/home')">На главную</ion-button>
    </div>

    <template v-else>
      <div class="exercise-layout">
        <div class="exercise-main">
          <h1 class="ex-name">{{ exercise?.name }}</h1>
          <p class="ex-desc">{{ exercise?.description }}</p>

          <div class="video-placeholder" aria-hidden="true">
            <span>Видео упражнения (скоро)</span>
          </div>

          <p class="reps-hint">
            Рекомендуем подобрать веса так, чтобы вы делали 6–12 повторений. Рекомендуемое количество
            подходов: 3.
          </p>

          <div class="sets-block">
            <p class="sets-label">Подходы</p>
            <div v-for="(_, si) in 3" :key="si" class="set-row">
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
            <ion-button fill="outline" class="nav-btn" :disabled="currentIndex <= 0" @click="session.goPrev">
              Выше
            </ion-button>
            <ion-button
              v-if="canSwapExercise"
              class="nav-btn nav-btn--primary"
              @click="goChangeExercise"
            >
              Поменять упражнение
            </ion-button>
            <p v-else class="swap-hint">Для этого слота в плане только одна вариация.</p>
            <ion-button
              fill="outline"
              class="nav-btn"
              :disabled="currentIndex >= slotCount - 1"
              @click="session.goNext"
            >
              Ниже
            </ion-button>
          </div>

          <ion-button
            expand="block"
            class="finish-temp-btn"
            color="success"
            :disabled="!canFinish || finishing"
            @click="onFinishWorkoutTemp"
          >
            Закончить тренировку (временная кнопка для отладки API)
          </ion-button>
          <p v-if="!canFinish" class="finish-hint">
            Заполните вес и повторы хотя бы в одном подходе — на сервер уйдут только заполненные подходы
            (<code class="finish-code">POST /api/v1/workouts/complete</code>).
          </p>
          <p v-else class="finish-hint finish-hint--ok">
            Отправятся только упражнения с заполненными подходами; незаполненные слоты в JSON не попадут.
          </p>
        </div>

        <div class="dots-col" aria-label="Прогресс по упражнениям">
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
      </div>
    </template>
  </workout-chrome>
</template>

<script setup>
defineOptions({ name: 'WorkoutExerciseScreen' })

import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import {
  IonHeader,
  IonToolbar,
  IonTitle,
  IonButtons,
  IonBackButton,
  IonButton,
  IonInput,
  toastController
} from '@ionic/vue'
import WorkoutChrome from '@/components/workout/WorkoutChrome.vue'
import { useWorkoutSessionStore } from '@/stores/workoutSession'

const router = useRouter()
const session = useWorkoutSessionStore()

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

const canSwapExercise = computed(() => (currentSlot.value?.alternatives?.length ?? 0) > 1)
const canFinish = computed(() => session.isReadyForComplete)
const finishing = ref(false)

function onSetInput(setIndex, field, value) {
  session.updateSet(setIndex, field, value)
}

function goChangeExercise() {
  if (!canSwapExercise.value) return
  router.push({
    name: 'WorkoutAlternatives',
    params: { slotIndex: String(session.currentIndex) }
  })
}

async function onFinishWorkoutTemp() {
  if (!canFinish.value || finishing.value) return
  finishing.value = true
  try {
    await session.submitCompleteWorkout()
    const toast = await toastController.create({
      message: 'Тренировка отправлена на сервер (тело — по ТЗ).',
      duration: 2500,
      color: 'success'
    })
    await toast.present()
    await router.replace('/home')
  } catch (e) {
    const msg =
      e?.response?.data?.message ||
      e?.message ||
      'Не удалось отправить (проверьте бэкенд и авторизацию).'
    const toast = await toastController.create({
      message: msg,
      duration: 3500,
      color: 'danger'
    })
    await toast.present()
  } finally {
    finishing.value = false
  }
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
  font-family: 'Roboto', sans-serif;
  font-weight: 600;
  font-size: 1rem;
  color: var(--sportik-text);
}

.empty-msg {
  text-align: center;
  padding: 2rem 0;
  font-family: 'Roboto', sans-serif;
  color: var(--sportik-text-muted);
}

.exercise-layout {
  display: flex;
  gap: 14px;
  align-items: flex-start;
}

.exercise-main {
  flex: 1;
  min-width: 0;
}

.ex-name {
  font-family: 'Roboto', sans-serif;
  font-weight: 700;
  font-size: 1.35rem;
  margin: 0 0 0.5rem;
  color: var(--sportik-text);
  line-height: 1.25;
}

.ex-desc {
  font-family: 'Roboto', sans-serif;
  font-size: 0.92rem;
  color: var(--sportik-text-muted);
  margin: 0 0 1rem;
  line-height: 1.45;
}

.video-placeholder {
  width: 100%;
  aspect-ratio: 16 / 9;
  border-radius: 14px;
  background: linear-gradient(145deg, #c4c4c4, #9e9e9e);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 1rem;
}

.video-placeholder span {
  font-family: 'Roboto', sans-serif;
  font-size: 0.9rem;
  color: rgba(255, 255, 255, 0.95);
  text-align: center;
  padding: 0 1rem;
}

.reps-hint {
  font-family: 'Roboto', sans-serif;
  font-size: 0.88rem;
  color: var(--sportik-text-soft);
  margin: 0 0 1.25rem;
  line-height: 1.45;
  background: var(--sportik-card-gray);
  padding: 12px 14px;
  border-radius: 12px;
}

.sets-block {
  margin-bottom: 1.25rem;
}

.sets-label {
  font-family: 'Roboto', sans-serif;
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
  font-family: 'Roboto', sans-serif;
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--sportik-text-muted);
  text-align: center;
}

.set-input {
  flex: 1;
  --background: var(--sportik-card-gray);
  --padding-start: 12px;
  --padding-end: 12px;
  border-radius: 10px;
  font-family: 'Roboto', sans-serif;
  min-height: 44px;
}

.nav-row {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.nav-btn {
  margin: 0;
  font-family: 'Roboto', sans-serif;
  font-size: 0.9rem;
}

.nav-btn--primary {
  --background: var(--ion-color-primary);
  --color: var(--ion-color-primary-contrast);
}

.finish-temp-btn {
  margin-top: 1rem;
  font-family: 'Roboto', sans-serif;
  font-size: 0.82rem;
}

.finish-hint {
  font-family: 'Roboto', sans-serif;
  font-size: 0.75rem;
  color: var(--sportik-text-muted);
  text-align: center;
  margin: 0.5rem 0 0;
  line-height: 1.4;
}

.finish-hint--ok {
  color: var(--sportik-text-soft);
}

.finish-code {
  font-size: 0.7rem;
  word-break: break-all;
}

.swap-hint {
  font-family: 'Roboto', sans-serif;
  font-size: 0.82rem;
  color: var(--sportik-text-muted);
  text-align: center;
  margin: 0;
  line-height: 1.35;
}

.dots-col {
  flex: 0 0 auto;
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding-top: 4px;
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
  box-shadow: 0 0 0 3px var(--sportik-cyan);
  background: #bdbdbd;
  transform: scale(1.15);
}

.dot--current-done {
  box-shadow: 0 0 0 3px var(--sportik-cyan);
  background: #2dd36f;
  transform: scale(1.15);
}

@media (min-width: 400px) {
  .nav-row {
    flex-direction: row;
    flex-wrap: wrap;
  }

  .nav-btn {
    flex: 1 1 30%;
  }
}
</style>
