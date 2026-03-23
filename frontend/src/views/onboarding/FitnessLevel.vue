<template>
  <onboarding-layout 
    title="Опыт" 
    :progress="50" 
    :disabled="!fitnessLevel"
    :loading="isSubmitting"
    @next="submit"
  >
    <div class="fitness-step">
      <h2>Твой уровень подготовки?</h2>
      <p>Это поможет нам подобрать упражнения нужной сложности.</p>
      
      <ion-list lines="none">
        <ion-radio-group v-model="fitnessLevel">
          <ion-item 
            v-for="item in fitnessLevels" 
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

const fitnessLevel = ref(authStore.profile?.fitness_level || null)
const isSubmitting = ref(false)

const fitnessLevels = [
  { value: 'beginner', label: 'Новичок (раньше не занимался)' },
  { value: 'intermediate', label: 'Средний (занимаюсь полгода-год)' },
  { value: 'advanced', label: 'Продвинутый (занимаюсь регулярно)' },
  { value: 'professional', label: 'Профи (спортсмен)' }
]

const submit = async () => {
  if (!fitnessLevel.value) return
  
  isSubmitting.value = true
  try {
    await authStore.updateProfile({ fitness_level: fitnessLevel.value })
    router.push('/health-restrictions')
  } catch (error) {
    console.error('Submit error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.fitness-step h2 {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 8px;
}

.fitness-step p {
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
