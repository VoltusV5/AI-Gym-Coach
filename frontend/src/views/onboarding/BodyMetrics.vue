<template>
  <onboarding-layout
    question="Рост и вес"
    :progress="12"
    :disabled="!canSubmit"
    :loading="isSubmitting"
    @next="submit"
  >
    <p class="hint">Сначала рост, затем вес — так удобнее сверяться с макетом.</p>
    <div class="metrics-row">
      <div class="metric-card">
        <label class="metric-label" for="onb-height">Рост</label>
        <div class="metric-field">
          <input
            id="onb-height"
            v-model="heightStr"
            class="metric-input metric-input-native"
            type="number"
            inputmode="numeric"
            placeholder="180"
            min="1"
            step="1"
            autocomplete="off"
          />
          <span class="metric-unit">см</span>
        </div>
      </div>
      <div class="metric-card">
        <label class="metric-label" for="onb-weight">Вес</label>
        <div class="metric-field">
          <input
            id="onb-weight"
            v-model="weightStr"
            class="metric-input metric-input-native"
            type="number"
            inputmode="decimal"
            placeholder="75"
            min="1"
            step="0.1"
            autocomplete="off"
          />
          <span class="metric-unit">кг</span>
        </div>
      </div>
    </div>
  </onboarding-layout>
</template>

<script setup>
import { computed, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import OnboardingLayout from '@/components/layout/OnboardingLayout.vue'
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
const router = useRouter()
const route = useRoute()

function initialStr(n) {
  return n != null && n !== '' && Number(n) > 0 ? String(n) : ''
}

const heightStr = ref(initialStr(authStore.profile?.height_cm))
const weightStr = ref(initialStr(authStore.profile?.weight_kg))
const isSubmitting = ref(false)

// Native <input type="number"> отдаёт строку. ion-input + v-model.number оставлял null/NaN,
// из-за чего :disabled оставался true и кнопка «Далее» казалась нерабочей.
const canSubmit = computed(() => {
  const hn = Number(String(heightStr.value).replace(',', '.'))
  const wn = Number(String(weightStr.value).replace(',', '.'))
  return Number.isFinite(hn) && Number.isFinite(wn) && hn > 0 && wn > 0
})

const submit = async () => {
  if (!canSubmit.value || isSubmitting.value) return

  isSubmitting.value = true
  try {
    const height_cm = Math.round(Number(String(heightStr.value).replace(',', '.')))
    const weight_kg = Math.round(Number(String(weightStr.value).replace(',', '.')))
    await authStore.updateProfile({ height_cm, weight_kg })
    
    if (route.query.regenerate === '1') {
      await router.push('/plan-generating')
    } else {
      await router.push('/gender')
    }
  } catch (error) {
    console.error('Submit error:', error)
    const status = error?.response?.status
    const msg = error?.response?.data?.message || error?.message || 'неизвестная ошибка'
    alert('Не удалось сохранить рост/вес: ' + (status ? `${status} ` : '') + msg)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.hint {
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
  background: var(--sportik-surface);
  border: 1px solid var(--sportik-border);
  border-radius: var(--sportik-radius-lg);
  padding: 8px 12px;
  box-shadow: var(--sportik-shadow-md);
}

.metric-input {
  font-weight: 600;
  font-size: 1.75rem;
  text-align: center;
}

.metric-input-native {
  flex: 1;
  min-width: 0;
  border: none;
  background: transparent;
  color: var(--sportik-text);
  font: inherit;
  outline: none;
  -moz-appearance: textfield;
}

.metric-input-native::-webkit-outer-spin-button,
.metric-input-native::-webkit-inner-spin-button {
  -webkit-appearance: none;
  margin: 0;
}

.metric-unit {
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
