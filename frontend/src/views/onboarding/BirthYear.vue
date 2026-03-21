<template>
  <onboarding-layout
    title="Год рождения"
    :progress="20"
    :disabled="!birthYear || birthYear < 1900 || birthYear > currentYear"
    :loading="isSubmitting"
    @next="submit"
  >
    <div class="birth-year-step">
      <h2>Твой год рождения</h2>
      <p>Нам это нужно, чтобы рассчитать твою нагрузку.</p>

      <ion-item fill="outline" mode="md" class="input-item">
        <ion-label position="stacked">Введи год</ion-label>
        <ion-input
          v-model="birthYear"
          type="number"
          placeholder="Напр: 1995"
          min="1900"
          :max="currentYear"
        ></ion-input>
      </ion-item>

      <p class="age-hint" v-if="calculatedAge">
        Твой возраст: {{ calculatedAge }}
      </p>
    </div>
  </onboarding-layout>
</template>

<script setup>
import { ref, computed } from 'vue'
import { IonItem, IonLabel, IonInput } from '@ionic/vue'
import OnboardingLayout from '@/components/layout/OnboardingLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const currentYear = new Date().getFullYear()
const birthYear = ref(null)
const isSubmitting = ref(false)

const calculatedAge = computed(() => {
  if (!birthYear.value || birthYear.value < 1900 || birthYear.value > currentYear) return null
  return currentYear - parseInt(birthYear.value)
})

const submit = async () => {
  if (!calculatedAge.value) return

  isSubmitting.value = true
  try {
    // отправляем кол-во полных лет в поле age
    await authStore.updateProfile({ age: calculatedAge.value })
    router.push('/body-metrics')
  } catch (error) {
    console.error('Submit error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.birth-year-step h2 {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 8px;
}

.birth-year-step p {
  color: var(--ion-color-medium);
  margin-bottom: 2rem;
}

.input-item {
  margin-bottom: 1rem;
}

.age-hint {
  font-size: 1.1rem;
  font-weight: 600;
  color: var(--ion-color-primary) !important;
}
</style>
