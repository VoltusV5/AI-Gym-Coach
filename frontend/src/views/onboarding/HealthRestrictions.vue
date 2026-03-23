<template>
  <onboarding-layout
    question="Травмы или болезни?"
    :progress="50"
    :disabled="choice === null"
    :loading="isSubmitting"
    next-label="Далее"
    @next="submit"
  >
    <p class="hint">Есть ли противопоказания, травмы или хронические заболевания?</p>
    <div class="yesno-tiles">
      <button
        type="button"
        class="yesno-tile"
        :class="{ selected: choice === false }"
        @click="choice = false"
      >
        <span class="yesno-title">Нет</span>
        <span class="yesno-sub">ограничений нет</span>
      </button>
      <button
        type="button"
        class="yesno-tile"
        :class="{ selected: choice === true }"
        @click="choice = true"
      >
        <span class="yesno-title">Да</span>
        <span class="yesno-sub">нужно учитывать</span>
      </button>
    </div>
  </onboarding-layout>
</template>

<script setup>
import { ref } from 'vue'
import OnboardingLayout from '@/components/layout/OnboardingLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const raw = authStore.profile?.injuries_notes
const choice = ref(
  typeof raw === 'boolean' ? raw : null
)
const isSubmitting = ref(false)

const submit = async () => {
  if (choice.value === null) return

  isSubmitting.value = true
  try {
    await authStore.updateProfile({ injuries_notes: choice.value })
    router.push('/goal-selection')
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
  line-height: 1.45;
}

.yesno-tiles {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 1rem;
  margin-top: 0.25rem;
}

.yesno-tile {
  flex: 1 1 140px;
  max-width: 200px;
  min-height: 120px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.35rem;
  padding: 1rem 1.25rem;
  border: 3px solid transparent;
  border-radius: var(--sportik-radius-xl);
  background: var(--sportik-cream);
  box-shadow: 0 6px 20px rgba(0, 0, 0, 0.07);
  cursor: pointer;
  font-family: 'Roboto', sans-serif;
  transition:
    border-color 0.2s,
    transform 0.2s;
}

.yesno-tile:active {
  transform: scale(0.98);
}

.yesno-tile.selected {
  border-color: var(--sportik-cyan);
  box-shadow:
    0 6px 24px rgba(102, 255, 255, 0.35),
    0 2px 8px rgba(0, 0, 0, 0.06);
}

.yesno-title {
  font-size: 1.75rem;
  font-weight: 700;
  color: var(--sportik-text);
}

.yesno-sub {
  font-size: 0.88rem;
  font-weight: 500;
  color: var(--sportik-text-muted);
  text-align: center;
  line-height: 1.3;
}
</style>
