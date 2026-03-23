<template>
  <onboarding-layout
    question="Удобные дни?"
    :progress="88"
    :disabled="!hasSelectedDays"
    :loading="isSubmitting"
    next-label="Завершить"
    @next="submit"
  >
    <p class="hint">Выбери дни тренировок — например понедельник, среда и суббота.</p>
    <div class="days-grid">
      <button
        v-for="day in days"
        :key="day.id"
        type="button"
        class="day-square"
        :class="{ selected: selectedDays[day.id] }"
        @click="toggleDay(day.id)"
      >
        {{ day.label }}
      </button>
    </div>
  </onboarding-layout>
</template>

<script setup>
import { ref, computed } from 'vue'
import OnboardingLayout from '@/components/layout/OnboardingLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'
import { trainingDaysToSelection, selectionToTrainingDaysArray } from '@/utils/trainingDays'

const authStore = useAuthStore()
const router = useRouter()

const days = [
  { id: 'mon', label: 'ПН' },
  { id: 'tue', label: 'ВТ' },
  { id: 'wed', label: 'СР' },
  { id: 'thu', label: 'ЧТ' },
  { id: 'fri', label: 'ПТ' },
  { id: 'sat', label: 'СБ' },
  { id: 'sun', label: 'ВС' }
]

const selectedDays = ref(trainingDaysToSelection(authStore.profile?.training_days_map))

const isSubmitting = ref(false)

const hasSelectedDays = computed(() => Object.values(selectedDays.value).some(Boolean))

const toggleDay = (id) => {
  selectedDays.value = {
    ...selectedDays.value,
    [id]: !selectedDays.value[id]
  }
}

const submit = async () => {
  if (!hasSelectedDays.value) return

  isSubmitting.value = true
  try {
    const arr = selectionToTrainingDaysArray(selectedDays.value)
    await authStore.updateProfile({ training_days_map: arr })
    router.replace('/plan-generating')
  } catch (error) {
    console.error('Submit error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.hint {
  font-family: 'Roboto', sans-serif;
  font-size: 0.95rem;
  color: var(--sportik-text-muted);
  text-align: center;
  margin: 0 0 1.5rem;
}

.days-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: clamp(4px, 2vw, 12px);
  width: 100%;
  max-width: min(100%, 299px);
  margin: 0 auto;
}

/* 7-й элемент (ВС) — по центру третьей строки */
.days-grid .day-square:nth-child(7) {
  grid-column: 2;
}

/* Ещё −10% */
.day-square {
  aspect-ratio: 1;
  width: 100%;
  min-width: 0;
  min-height: clamp(56px, 15vw, 77px);
  border-radius: 9px;
  border: 3px solid var(--sportik-card-gray);
  background: var(--sportik-cream);
  font-family: 'Roboto', sans-serif;
  font-weight: 700;
  font-size: clamp(0.7rem, 2.43vw, 0.9rem);
  color: var(--sportik-text);
  cursor: pointer;
  transition:
    border-color 0.2s,
    background 0.2s,
    transform 0.15s;
}

.day-square:active {
  transform: scale(0.97);
}

.day-square.selected {
  border-color: var(--sportik-cyan);
  background: #e6ffff;
  box-shadow: 0 0 0 2px rgba(102, 255, 255, 0.35);
}
</style>
