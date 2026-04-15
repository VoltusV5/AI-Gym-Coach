<template>
  <workout-chrome>
    <ion-header class="ion-no-border alt-header">
      <ion-toolbar class="alt-toolbar">
        <ion-buttons slot="start">
          <ion-back-button :default-href="backHref" text="" color="dark"></ion-back-button>
        </ion-buttons>
        <ion-title class="alt-title">Поменять упражнение</ion-title>
      </ion-toolbar>
    </ion-header>

    <div v-if="!slot" class="empty-msg">
      <p>Слот не найден.</p>
      <ion-button @click="router.back()">Назад</ion-button>
    </div>

    <template v-else>
      <p class="alt-sub">
        Текущее: <strong>{{ currentName }}</strong>
      </p>
      <p class="alt-hint">Выберите вариант — вернёмся к тренировке с новым упражнением.</p>

      <div class="alt-grid">
        <button
          v-for="alt in slot.alternatives"
          :key="String(alt.id)"
          type="button"
          class="alt-card"
          :class="{ 'alt-card--active': Number(alt.id) === Number(slot.selectedId) }"
          @click="pick(alt.id)"
        >
          <div class="alt-thumb" aria-hidden="true"></div>
          <p class="alt-card-name">{{ alt.name }}</p>
          <p class="alt-card-desc">{{ alt.description }}</p>
        </button>
      </div>

      <ion-button expand="block" fill="outline" class="back-btn" @click="router.back()">Вернуться</ion-button>
    </template>
  </workout-chrome>
</template>

<script setup>
defineOptions({ name: 'WorkoutAlternativesScreen' })

import { computed, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { IonHeader, IonToolbar, IonTitle, IonButtons, IonBackButton, IonButton } from '@ionic/vue'
import WorkoutChrome from '@/components/workout/WorkoutChrome.vue'
import { useWorkoutSessionStore } from '@/stores/workoutSession'

const props = defineProps({
  slotIndex: {
    type: String,
    required: true
  }
})

const route = useRoute()
const router = useRouter()
const session = useWorkoutSessionStore()

const idx = computed(() => {
  const n = Number.parseInt(props.slotIndex ?? route.params.slotIndex, 10)
  return Number.isFinite(n) ? n : -1
})

const slot = computed(() => {
  const i = idx.value
  if (i < 0 || i >= session.slots.length) return null
  return session.slots[i]
})

const currentName = computed(() => {
  const s = slot.value
  if (!s) return ''
  const ex = s.alternatives.find((a) => Number(a.id) === Number(s.selectedId))
  return ex?.name ?? ''
})

const backHref = computed(() => '/workout/session')

onMounted(async () => {
  session.hydrate()
  await nextTick()
  const i = idx.value
  if (i < 0 || i >= session.slots.length) return
  const s = session.slots[i]
  if (s && s.alternatives.length <= 1) {
    router.replace(backHref.value)
  }
})

function pick(alternativeId) {
  const i = idx.value
  if (i < 0) return
  if (slot.value && slot.value.selectedId !== alternativeId) {
    session.selectAlternative(i, alternativeId)
  }
  router.back()
}
</script>

<style scoped>
.alt-header {
  box-shadow: none;
}

.alt-toolbar {
  --background: transparent;
  --border-width: 0;
}

.alt-title {
  font-weight: 700;
  font-size: 1rem;
  color: var(--sportik-text);
}

.empty-msg {
  text-align: center;
  padding: 2rem 0;
  color: var(--sportik-text-muted);
  font-family: 'Roboto', sans-serif;
}

.alt-sub {
  font-size: 0.92rem;
  color: var(--sportik-text-muted);
  margin: 0 0 0.35rem;
}

.alt-hint {
  font-size: 0.85rem;
  color: var(--sportik-text-soft);
  margin: 0 0 1rem;
  line-height: 1.4;
}

.alt-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 1.25rem;
}

.alt-card {
  text-align: left;
  border: none;
  border-radius: 16px;
  padding: 10px;
  background: var(--sportik-surface-soft);
  border: 1px solid var(--sportik-border);
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 6px;
  min-height: 140px;
  transition:
    box-shadow 0.15s ease,
    transform 0.15s ease;
}

.alt-card:active {
  transform: scale(0.98);
}

.alt-card--active {
  box-shadow: 0 0 0 2px color-mix(in srgb, var(--sportik-brand) 60%, transparent);
}

.alt-thumb {
  width: 100%;
  aspect-ratio: 4 / 3;
  border-radius: 10px;
  background: linear-gradient(
    135deg,
    color-mix(in srgb, var(--sportik-brand) 28%, var(--sportik-surface)) 0%,
    color-mix(in srgb, var(--sportik-brand-2) 20%, var(--sportik-surface)) 100%
  );
}

.alt-card-name {
  font-weight: 700;
  font-size: 0.82rem;
  margin: 0;
  color: var(--sportik-text);
  line-height: 1.25;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.alt-card-desc {
  font-weight: 300;
  font-size: 0.72rem;
  margin: 0;
  color: var(--sportik-text-muted);
  line-height: 1.35;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  overflow: hidden;
  flex: 1;
}

.back-btn {
  margin-top: 0.25rem;
  font-family: 'Roboto', sans-serif;
}
</style>
