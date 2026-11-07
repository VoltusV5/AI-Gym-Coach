<template>
  <onboarding-layout
    question="Уровень подготовки"
    :progress="75"
    :disabled="!fitnessLevel"
    :loading="isSubmitting"
    :error="submitError"
    :status="submitStatus"
    @next="submit"
  >
    <ion-list lines="none" class="options-list">
      <ion-radio-group v-model="fitnessLevel">
        <ion-item
          v-for="item in fitnessLevels"
          :key="item.value"
          lines="none"
          class="option-card"
          :class="{ 'option-card--checked': fitnessLevel === item.value }"
        >
          <ion-radio :value="item.value" justify="space-between" label-placement="end">
            {{ item.label }}
          </ion-radio>
        </ion-item>
      </ion-radio-group>
    </ion-list>
  </onboarding-layout>
</template>

<script setup>
import { ref } from 'vue'
import { IonList, IonRadioGroup, IonItem, IonRadio } from '@ionic/vue'
import OnboardingLayout from '@/components/layout/OnboardingLayout.vue'
import { apiStatus, errorText } from '@/api/v2'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const fitnessLevel = ref(authStore.profile?.fitness_level || null)
const isSubmitting = ref(false)
const submitError = ref('')
const submitStatus = ref(0)

const fitnessLevels = [
  { value: 'Новичок', label: 'Новичок' },
  { value: 'Любитель', label: 'Любитель' },
  { value: 'Продвинутый', label: 'Продвинутый' }
]

/** Save this questionnaire step once and leave the input intact on errors. */
const submit = async () => {
  if (isSubmitting.value || !fitnessLevel.value) return

  isSubmitting.value = true
  submitError.value = ''
  submitStatus.value = 0
  try {
    await authStore.updateProfile({ fitness_level: fitnessLevel.value })
    router.push('/training-days')
  } catch (error) {
    submitStatus.value = apiStatus(error)
    submitError.value = errorText(error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.options-list {
  background: transparent;
  padding: 0;
}

.option-card {
  --background: var(--sportik-surface);
  --border-radius: var(--sportik-radius-lg);
  margin-bottom: 12px;
  --padding-start: 16px;
  --padding-end: 12px;
  box-shadow: var(--sportik-shadow-md);
  border: 1px solid var(--sportik-border);
}

.option-card--checked {
  border-color: color-mix(in srgb, var(--sportik-brand) 70%, transparent);
  --background: color-mix(in srgb, var(--sportik-brand) 8%, var(--sportik-surface));
}

ion-radio {
  width: 100%;
  font-weight: 600;
  font-size: 1.15rem;
}
</style>