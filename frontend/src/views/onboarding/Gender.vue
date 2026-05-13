<template>
  <onboarding-layout
    question="Ваш пол?"
    :progress="22"
    :disabled="!gender"
    :loading="isSubmitting"
    @next="submit"
  >
    <div class="gender-tiles">
      <button
        type="button"
        class="gender-tile"
        :class="{ selected: gender === 'Мужчина' }"
        @click="gender = 'Мужчина'"
      >
        <span class="gender-letter">М</span>
        <span class="gender-caption">Мужской</span>
      </button>
      <button
        type="button"
        class="gender-tile"
        :class="{ selected: gender === 'Женщина' }"
        @click="gender = 'Женщина'"
      >
        <span class="gender-letter">Ж</span>
        <span class="gender-caption">Женский</span>
      </button>
    </div>
  </onboarding-layout>
</template>

<script setup>
import { ref } from 'vue'
import OnboardingLayout from '@/components/layout/OnboardingLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

function normalizeGenderFromProfile(g) {
  if (g == null || g === '') return null
  if (g === 'Мужчина' || g === 'male' || g === 'м') return 'Мужчина'
  if (g === 'Женщина' || g === 'female' || g === 'ж') return 'Женщина'
  return null
}

const authStore = useAuthStore()
const router = useRouter()

const gender = ref(normalizeGenderFromProfile(authStore.profile?.gender))
const isSubmitting = ref(false)

const submit = async () => {
  if (!gender.value) return

  isSubmitting.value = true
  try {
    await authStore.updateProfile({ gender: gender.value })
    router.push('/age')
  } catch (error) {
    console.error('Submit error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.gender-tiles {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 1.25rem;
  margin-top: 0.5rem;
}

.gender-tile {
  flex: 1 1 140px;
  max-width: 200px;
  min-height: 168px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  padding: 1.25rem;
  border: 1px solid var(--sportik-border);
  border-radius: var(--sportik-radius-lg);
  background: var(--sportik-surface);
  box-shadow: var(--sportik-shadow-md);
  cursor: pointer;
  transition:
    border-color 0.2s,
    transform 0.2s,
    box-shadow 0.2s;
}

.gender-tile:active {
  transform: scale(0.98);
}

.gender-tile.selected {
  border-color: color-mix(in srgb, var(--sportik-brand) 70%, transparent);
  box-shadow: var(--sportik-shadow-lg);
}

.gender-letter {
  font-size: clamp(3.5rem, 12vw, 5rem);
  font-weight: 600;
  line-height: 1;
  color: var(--sportik-text);
}

.gender-caption {
  font-size: 0.95rem;
  font-weight: 500;
  color: var(--sportik-text-muted);
}
</style>