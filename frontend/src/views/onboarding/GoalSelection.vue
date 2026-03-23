<template>
  <onboarding-layout 
    title="Цель" 
    :progress="80" 
    :disabled="!goal"
    :loading="isSubmitting"
    @next="submit"
  >
    <div class="goal-step">
      <h2>Какая твоя цель?</h2>
      <p>Мы составим план, который поможет её достичь быстрее.</p>
      
      <ion-list lines="none">
        <ion-radio-group v-model="goal">
          <ion-item 
            v-for="item in goals" 
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

const goal = ref(authStore.profile?.goal || null)
const isSubmitting = ref(false)

const goals = [
  { value: 'weight_loss', label: 'Похудеть (сбросить лишний жир)' },
  { value: 'muscle_gain', label: 'Набрать мышечную массу' },
  { value: 'keep_fit', label: 'Поддерживать форму и здоровье' },
  { value: 'increase_stamina', label: 'Увеличить выносливость' }
]

const submit = async () => {
  if (!goal.value) return
  
  isSubmitting.value = true
  try {
    await authStore.updateProfile({ goal: goal.value })
    router.push('/training-days')
  } catch (error) {
    console.error('Submit error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.goal-step h2 {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 8px;
}

.goal-step p {
  color: var(--ion-color-medium);
  margin-bottom: 2rem;
}

.radio-item {
  --background: var(--ion-color-light);
  --border-radius: 12px;
}

ion-radio {
  width: 100%;
}
</style>
