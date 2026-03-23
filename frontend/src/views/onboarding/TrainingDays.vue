<template>
  <onboarding-layout
    title="График"
    :progress="95"
    :disabled="!hasSelectedDays"
    :loading="isSubmitting"
    next-label="Завершить"
    @next="submit"
  >
    <div class="days-step">
      <h2>Дни для тренировок</h2>
      <p>Выбери дни недели, когда тебе удобно заниматься.</p>

      <div class="days-grid">
        <div
          v-for="day in days"
          :key="day.id"
          class="day-card"
          :class="{ 'selected': selectedDays[day.id] }"
          @click="toggleDay(day.id)"
        >
          <span class="day-label">{{ day.label }}</span>
        </div>
      </div>
    </div>
  </onboarding-layout>
</template>

<script setup>
import { ref, computed } from 'vue'
import OnboardingLayout from '@/components/layout/OnboardingLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const days = [
  { id: 'mon', label: 'Пн' },
  { id: 'tue', label: 'Вт' },
  { id: 'wed', label: 'Ср' },
  { id: 'thu', label: 'Чт' },
  { id: 'fri', label: 'Пт' },
  { id: 'sat', label: 'Сб' },
  { id: 'sun', label: 'Вс' }
]

const selectedDays = ref(authStore.profile?.training_days_map || {
  mon: false, tue: false, wed: false, thu: false, fri: false, sat: false, sun: false
})

const isSubmitting = ref(false)

const hasSelectedDays = computed(() => {
  return Object.values(selectedDays.value).some(val => !!val)
})

const toggleDay = (id) => {
  selectedDays.value[id] = !selectedDays.value[id]
}

const submit = async () => {
  if (!hasSelectedDays.value) return

  isSubmitting.value = true
  try {
    await authStore.updateProfile({ training_days_map: selectedDays.value })

    // Переходим на страницу генерации плана
    router.replace('/plan-generating')
  } catch (error) {
    console.error('Submit error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.days-step h2 {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 8px;
}

.days-step p {
  color: var(--ion-color-medium);
  margin-bottom: 2rem;
}

.days-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.day-card {
  aspect-ratio: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--ion-color-light);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  border: 2px solid transparent;
}

.day-card.selected {
  background: var(--ion-color-primary-contrast);
  border-color: var(--ion-color-primary);
  color: var(--ion-color-primary);
  font-weight: 700;
}

.day-label {
  font-size: 1.1rem;
}
</style>
