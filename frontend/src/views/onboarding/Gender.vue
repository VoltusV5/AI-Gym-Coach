<template>
  <onboarding-layout 
    title="Твой пол" 
    :progress="10" 
    :disabled="!gender"
    :loading="isSubmitting"
    @next="submit"
  >
    <div class="gender-selection">
      <h2>Кто ты?</h2>
      <p>Это поможет нам точнее рассчитать твои показатели.</p>
      
      <ion-list lines="none">
        <ion-radio-group v-model="gender">
          <ion-item class="radio-item ion-margin-bottom">
            <ion-radio value="male" justify="space-between">Мужчина</ion-radio>
          </ion-item>
          <ion-item class="radio-item">
            <ion-radio value="female" justify="space-between">Женщина</ion-radio>
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

const gender = ref(authStore.profile?.gender || null)
const isSubmitting = ref(false)

const submit = async () => {
  if (!gender.value) return
  
  isSubmitting.value = true
  try {
    await authStore.updateProfile({ gender: gender.value })
    router.push('/birth-year')
  } catch (error) {
    console.error('Submit error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.gender-selection h2 {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 8px;
}

.gender-selection p {
  color: var(--ion-color-medium);
  margin-bottom: 2rem;
}

.radio-item {
  --background: var(--ion-color-light);
  --border-radius: 12px;
  --padding-start: 16px;
  --padding-end: 16px;
  border: 1px solid transparent;
}

ion-item::part(native) {
  border-radius: 12px;
}

ion-radio {
  --color-checked: var(--ion-color-primary);
  width: 100%;
}
</style>
