<template>
  <onboarding-layout
    question="Сколько тебе лет?"
    :progress="28"
    :disabled="age == null || age < 10 || age > 100"
    :loading="isSubmitting"
    @next="submit"
  >
    <div class="age-block">
      <ion-item lines="none" class="age-field">
        <ion-input
          v-model.number="age"
          type="number"
          inputmode="numeric"
          placeholder="25"
          :min="10"
          :max="100"
          class="age-input"
        ></ion-input>
      </ion-item>
      <p class="age-hint">Укажи возраст числом (полных лет).</p>
    </div>
  </onboarding-layout>
</template>

<script setup>
defineOptions({ name: 'OnboardingAgePage' })

import { ref } from 'vue'
import { IonItem, IonInput } from '@ionic/vue'
import OnboardingLayout from '@/components/layout/OnboardingLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const existing = authStore.profile?.age
const age = ref(existing != null && existing > 0 ? existing : null)
const isSubmitting = ref(false)

const submit = async () => {
  if (age.value == null || age.value < 10 || age.value > 100) return

  isSubmitting.value = true
  try {
    await authStore.updateProfile({ age: Math.round(Number(age.value)) })
    router.push('/activity-type')
  } catch (error) {
    console.error('Submit error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.age-block {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 1rem;
  margin-top: 1rem;
}

.age-field {
  --background: var(--sportik-surface);
  --border-radius: var(--sportik-radius-lg);
  width: 100%;
  max-width: 280px;
  border: 1px solid var(--sportik-border);
  box-shadow: var(--sportik-shadow-md);
}

.age-input {
  font-weight: 600;
  font-size: clamp(2.5rem, 10vw, 4rem);
  text-align: center;
  --padding-top: 1rem;
  --padding-bottom: 1rem;
}

.age-hint {
  font-size: 0.95rem;
  color: var(--sportik-text-muted);
  text-align: center;
  margin: 0;
  max-width: 280px;
  line-height: 1.4;
}
</style>
