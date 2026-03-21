<template>
  <onboarding-layout
    title="Здоровье"
    :progress="60"
    :loading="isSubmitting"
    @next="submit"
  >
    <div class="health-step">
      <h2>Ограничения?</h2>
      <p>Расскажи о травмах или заболеваниях, если они есть. Если нет — просто нажми «Далее».</p>

      <ion-item fill="outline" mode="md" class="textarea-item">
        <ion-label position="stacked">Травмы или болезни</ion-label>
        <ion-textarea
          v-model="notes"
          placeholder="Напр: болят колени, грыжа в пояснице..."
          rows="4"
          auto-grow
        ></ion-textarea>
      </ion-item>
    </div>
  </onboarding-layout>
</template>

<script setup>
import { ref } from 'vue'
import { IonItem, IonLabel, IonTextarea } from '@ionic/vue'
import OnboardingLayout from '@/components/layout/OnboardingLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const notes = ref(authStore.profile?.injuries_notes || '')
const isSubmitting = ref(false)

const submit = async () => {
  isSubmitting.value = true
  try {
    await authStore.updateProfile({ injuries_notes: notes.value || 'none' })
    router.push('/goal-selection')
  } catch (error) {
    console.error('Submit error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.health-step h2 {
  font-size: 24px;
  font-weight: 700;
  margin-bottom: 8px;
}

.health-step p {
  color: var(--ion-color-medium);
  margin-bottom: 2rem;
}

.textarea-item {
  margin-bottom: 1rem;
}
</style>
