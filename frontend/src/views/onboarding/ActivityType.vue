<template>
  <onboarding-layout
    question="Тип активности"
    :progress="38"
    :disabled="!activityLevel"
    :loading="isSubmitting"
    @next="submit"
  >
    <div class="options-list" role="radiogroup" aria-label="Тип активности">
      <button
        v-for="item in activityLevels"
        :key="item.value"
        type="button"
        class="option-card"
        :class="{ 'option-card--checked': activityLevel === item.value }"
        role="radio"
        :aria-checked="activityLevel === item.value"
        @click="activityLevel = item.value"
      >
        <span class="radio-dot" :class="{ 'radio-dot--on': activityLevel === item.value }" aria-hidden="true"></span>
        <span class="option-label">{{ item.label }}</span>
      </button>
    </div>
  </onboarding-layout>
</template>

<script setup>
import { ref } from 'vue'
import OnboardingLayout from '@/components/layout/OnboardingLayout.vue'
import { useAuthStore } from '@/stores/auth'
import { useRouter } from 'vue-router'

const authStore = useAuthStore()
const router = useRouter()

const activityLevel = ref(authStore.profile?.activity_level || null)
const isSubmitting = ref(false)

const activityLevels = [
  {
    value: 'Сидячий и малоподвижный',
    label: 'Сидячий и малоподвижный'
  },
  {
    value: 'Лёгкая активность (физические нагрузки 1-3 раза в неделю)',
    label: 'Лёгкая активность (физические нагрузки 1–3 раза в неделю)'
  },
  {
    value: 'Средняя активность (физические нагрузки 3-5 раза в неделю)',
    label: 'Средняя активность (физические нагрузки 3–5 раза в неделю)'
  },
  {
    value: 'Высокая активность (физические нагрузки 6-7 раз в неделю)',
    label: 'Высокая активность (физические нагрузки 6–7 раз в неделю)'
  },
  {
    value: 'Очень высокая активность (постоянно высокая физическая нагрузка)',
    label: 'Очень высокая активность (постоянно высокая физическая нагрузка)'
  }
]

const submit = async () => {
  if (!activityLevel.value) return

  isSubmitting.value = true
  try {
    await authStore.updateProfile({ activity_level: activityLevel.value })
    router.push('/health-restrictions')
  } catch (error) {
    console.error('Submit error:', error)
  } finally {
    isSubmitting.value = false
  }
}
</script>

<style scoped>
.options-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
}

.option-card {
  display: flex;
  flex-direction: row;
  align-items: flex-start;
  gap: 12px;
  width: 100%;
  margin: 0;
  padding: 14px 16px;
  text-align: left;
  border: 1px solid var(--sportik-border);
  border-radius: var(--sportik-radius-lg);
  background: var(--sportik-surface);
  box-shadow: var(--sportik-shadow-md);
  cursor: pointer;
  font: inherit;
  color: inherit;
  transition:
    border-color 0.2s,
    background 0.2s,
    transform 0.1s;
}

.option-card:active {
  transform: scale(0.99);
}

.option-card--checked {
  border-color: color-mix(in srgb, var(--sportik-brand) 70%, transparent);
  background: color-mix(in srgb, var(--sportik-brand) 8%, var(--sportik-surface));
}

.radio-dot {
  flex-shrink: 0;
  width: 22px;
  height: 22px;
  margin-top: 2px;
  border-radius: 50%;
  border: 2px solid var(--sportik-border);
  background: var(--sportik-surface-soft);
  box-sizing: border-box;
}

.radio-dot--on {
  border-color: var(--sportik-brand);
  box-shadow: inset 0 0 0 5px color-mix(in srgb, var(--sportik-brand) 75%, transparent);
}

.option-label {
  flex: 1;
  min-width: 0;
  font-weight: 600;
  font-size: 0.92rem;
  line-height: 1.35;
  white-space: normal;
  text-align: left;
}
</style>