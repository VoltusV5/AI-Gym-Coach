<template>
  <onboarding-layout 
    title="Активность" 
    :progress="40" 
    :disabled="!activityLevel"
    :loading="isSubmitting"
    @next="submit"
  >
    <div class="activity-step">
      <h2>Твой образ жизни?</h2>
      <p>Оцени свою активность в течение дня (не считая тренировок).</p>
      
      <ion-list lines="none">
        <ion-radio-group v-model="activityLevel">
          <ion-item 
            v-for="item in activityLevels" 
            :key="item.value"
            class="radio-item ion-margin-bottom"
          >
            <ion-radio :value="item.value" justify="space-between">
              {{ item.label }}
            </ion-radio>
          </ion-item>
        </ion-radio-group>
      </ion-list>
    </div>
  </onboarding-layout>
</template>

<script setup>
import { ref } from 'vue'
import { IonList, IonRadioGroup, IonItem, IonRadio } from '@ionic/vue'
import OnboardingLayout from '@/components/layout/OnboardingLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const activityLevel = ref(authStore.profile?.activity_level || null)
const isSubmitting = ref(false)

const activityLevels = [
  { value: 'sedentary', label: 'Сидячий образ жизни' },
  { value: 'light', label: 'Небольшая активность (пешие прогулки)' },
  { value: 'moderate', label: 'Средняя (активный образ жизни)' },
  { value: 'active', label: 'Высокая (много движения)' },
  { value: 'very_active', label: 'Очень высокая (физический труд)' }
]

const submit = async () => {
  if (!activityLevel.value) return
  
  isSubmitting.value = true
  try {
    await authStore.updateProfile({ activity_level: activityLevel.value })
    router.push('/fitness-level')
  } catch (error) {
    console.error('Submit error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.activity-step h2 {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 8px;
}

.activity-step p {
  color: var(--ion-color-medium);
  margin-bottom: 2rem;
}

.radio-item {
  --background: var(--ion-color-light);
  --border-radius: 12px;
  border: 1px solid transparent;
}

ion-radio {
  width: 100%;
}
</style>
