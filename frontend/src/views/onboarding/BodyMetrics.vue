<template>
  <onboarding-layout
    title="Вес и рост"
    :progress="30"
    :disabled="!height || !weight"
    :loading="isSubmitting"
    @next="submit"
  >
    <div class="metrics-step">
      <h2>Твои параметры</h2>
      <p>Они необходимы для расчёта индекса массы тела (ИМТ).</p>

      <ion-item fill="outline" mode="md" class="input-item ion-margin-bottom">
        <ion-label position="stacked">Рост (см)</ion-label>
        <ion-input
          v-model="height"
          type="number"
          placeholder="Напр: 178"
          min="50"
          max="300"
        ></ion-input>
      </ion-item>

      <ion-item fill="outline" mode="md" class="input-item">
        <ion-label position="stacked">Вес (кг)</ion-label>
        <ion-input
          v-model="weight"
          type="number"
          placeholder="Напр: 75"
          min="20"
          max="300"
        ></ion-input>
      </ion-item>
    </div>
  </onboarding-layout>
</template>

<script setup>
import { ref } from 'vue'
import { IonItem, IonLabel, IonInput } from '@ionic/vue'
import OnboardingLayout from '@/components/layout/OnboardingLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const height = ref(authStore.profile?.height_cm || null)
const weight = ref(authStore.profile?.weight_kg || null)
const isSubmitting = ref(false)

const submit = async () => {
  if (!height.value || !weight.value) return

  isSubmitting.value = true
  try {
    await authStore.updateProfile({
      height_cm: parseFloat(height.value),
      weight_kg: parseFloat(weight.value)
    })
    router.push('/activity-type')
  } catch (error) {
    console.error('Submit error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.metrics-step h2 {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 8px;
}

.metrics-step p {
  color: var(--ion-color-medium);
  margin-bottom: 2rem;
}

.input-item {
  margin-bottom: 1rem;
}
</style>
