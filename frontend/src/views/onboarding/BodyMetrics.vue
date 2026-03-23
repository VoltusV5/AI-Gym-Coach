<template>
  <onboarding-layout
    question="Рост и вес"
    :progress="12"
    :disabled="!height || !weight"
    :loading="isSubmitting"
    @next="submit"
  >
    <p class="hint">Сначала рост, затем вес — так удобнее сверяться с макетом.</p>
    <div class="metrics-row">
      <div class="metric-card">
        <label class="metric-label">Рост</label>
        <div class="metric-field">
          <ion-input
            v-model.number="height"
            type="number"
            inputmode="numeric"
            placeholder="180"
            class="metric-input"
          ></ion-input>
          <span class="metric-unit">см</span>
        </div>
      </div>
      <div class="metric-card">
        <label class="metric-label">Вес</label>
        <div class="metric-field">
          <ion-input
            v-model.number="weight"
            type="number"
            inputmode="decimal"
            placeholder="75"
            class="metric-input"
          ></ion-input>
          <span class="metric-unit">кг</span>
        </div>
      </div>
    </div>
  </onboarding-layout>
</template>

<script setup>
import { ref } from 'vue'
import { IonInput } from '@ionic/vue'
import OnboardingLayout from '@/components/layout/OnboardingLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const height = ref(authStore.profile?.height_cm ?? null)
const weight = ref(authStore.profile?.weight_kg ?? null)
const isSubmitting = ref(false)

const submit = async () => {
  if (!height.value || !weight.value) return

  isSubmitting.value = true
  try {
    await authStore.updateProfile({
      height_cm: Math.round(Number(height.value)),
      weight_kg: Math.round(Number(weight.value))
    })
    router.push('/gender')
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
  font-size: 1rem;
  color: var(--sportik-text-muted);
  text-align: center;
  margin: 0 0 1.5rem;
  line-height: 1.4;
}

.metrics-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1rem;
  width: 100%;
}

.metric-card {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.metric-label {
  font-family: 'Roboto', sans-serif;
  font-weight: 600;
  font-size: 1.25rem;
  color: var(--sportik-text);
  text-align: center;
}

.metric-field {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  background: var(--sportik-cream);
  border-radius: var(--sportik-radius-lg);
  padding: 8px 12px;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.06);
}

.metric-input {
  font-family: 'Roboto', sans-serif;
  font-weight: 600;
  font-size: 1.75rem;
  text-align: center;
}

.metric-unit {
  font-family: 'Roboto', sans-serif;
  font-weight: 500;
  color: var(--sportik-text-muted);
  font-size: 1rem;
  align-self: center;
  margin-right: 4px;
}

@media (max-width: 380px) {
  .metrics-row {
    grid-template-columns: 1fr;
  }
}
</style>
