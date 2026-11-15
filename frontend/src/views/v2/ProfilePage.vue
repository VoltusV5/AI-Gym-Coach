<script setup>
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import { getReference, v2 } from '@/api/v2'
import { useRead, useWrite } from '@/composables/useRequest'
import { useAuthStore } from '@/stores/auth'
import { stashDraft, takeDraft } from '@/utils/reauthDrafts'
import { profilePatch, levels, weekdays } from '@/utils/profile'

const props = defineProps({ onboarding: Boolean })
const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const draft = ref(null)
const original = ref(null)
const saved = ref(false)
const conflict = ref(false)
const training = ['training_days', 'equipment_ids', 'session_duration_min']
const fields = computed(() => props.onboarding ? training : [
  'age', 'height_cm', 'weight_kg', 'gender', 'activity_level', 'injuries_notes', 'goal', 'level', ...training,
])
const profile = useRead(async (signal) => {
  const [data, equipment] = await Promise.all([
    v2('GET', '/me/profile', { signal }),
    getReference('/equipment'),
  ])
  return { data, equipment }
})
const { saving, error, status, run } = useWrite()
const patch = computed(() => draft.value && profilePatch(original.value, draft.value, fields.value))

watch(() => auth.accountId, (account, before) => {
  if (account || before === undefined) {
    draft.value = null
    original.value = null
    profile.load()
  }
}, { immediate: true })
watch(profile.data, (loaded) => {
  if (!loaded) return
  if (!draft.value) {
    const savedDraft = takeDraft(auth.accountId, props.onboarding ? 'equipment' : 'profile')
    draft.value = savedDraft
      ? { ...loaded.data, ...profilePatch(savedDraft.original, savedDraft.draft, fields.value), version: loaded.data.version }
      : structuredClone(loaded.data)
  } else if (conflict.value) {
    // Preserve only the edits made against the old version when rebasing.
    draft.value = { ...loaded.data, ...patch.value, version: loaded.data.version }
  }
  original.value = loaded.data
  conflict.value = false
})

/** Store edited fields once, keeping the form intact on transport and version failures. */
function save() {
  return run(async () => {
    saved.value = false
    if (Object.keys(patch.value).length > 1) {
      const data = await v2('PATCH', '/me/profile', { data: patch.value })
      original.value = data
      draft.value = structuredClone(data)
      auth.fetchProfile().catch(() => {})
    }
    saved.value = true
    if (props.onboarding) {
      auth.setEquipmentStep(false)
      await router.replace('/plans/generate')
    }
  })
}

watch(() => auth.token, (token) => {
  if (!token && draft.value) {
    stashDraft(auth.lastSubject, props.onboarding ? 'equipment' : 'profile', { original: original.value, draft: draft.value })
  }
})

/** Refresh the version while preserving the user's modified fields for review. */
function refreshVersion() {
  conflict.value = true
  return profile.load().then(() => {
    if (!conflict.value) {
      status.value = 0
      error.value = ''
    }
  })
}
</script>

<template>
  <FeaturePage :eyebrow="onboarding ? 'Подготовка к плану' : 'Профиль'"
    :title="onboarding ? 'Оборудование и расписание' : 'Мой профиль'"
    :subtitle="onboarding ? 'Укажите, где и сколько времени вы можете тренироваться.' : 'Эти данные учитываются при создании тренировочного плана.'"
    :back="onboarding ? '/training-days' : '/settings'">
    <div class="feature-stack">
      <RouterLink v-if="!auth.token || (!onboarding && auth.isGuest)" class="feature-button secondary"
        :to="{ path: '/auth/login', query: { return_to: route.fullPath } }">Войти или зарегистрироваться</RouterLink>
      <RequestState :loading="profile.loading.value" :error="profile.error.value" :status="profile.status.value" :retry="profile.load" />
      <form v-if="draft && auth.token" class="feature-card" @submit.prevent="save">
        <fieldset :disabled="saving || profile.loading.value || status === 409" class="feature-stack">
          <legend>{{ onboarding ? 'Настройки тренировок' : 'Анкета' }}</legend>
          <div v-if="!onboarding" class="feature-fields">
            <label>Возраст<input v-model.number="draft.age" type="number" min="1" max="120" /></label>
            <label>Рост, см<input v-model.number="draft.height_cm" type="number" min="1" max="300" /></label>
            <label>Вес, кг<input v-model.number="draft.weight_kg" type="number" min="0.01" max="1000" step="any" /></label>
            <label>Пол<input v-model.trim="draft.gender" maxlength="100" /></label>
            <label>Активность<input v-model.trim="draft.activity_level" maxlength="200" /></label>
            <label>Цель<input v-model.trim="draft.goal" maxlength="200" /></label>
            <label>Уровень<select v-model="draft.level">
              <option :value="null" disabled>Выберите уровень</option>
              <option v-for="(label, key) in levels" :key="key" :value="key">{{ label }}</option>
            </select></label>
            <label>Ограничения по здоровью<select v-model="draft.injuries_notes">
              <option :value="null" disabled>Укажите ответ</option>
              <option :value="false">Нет</option><option :value="true">Есть</option>
            </select></label>
          </div>
          <label>Длительность занятия, мин
            <input v-model.number="draft.session_duration_min" type="number" min="1" max="480" :required="onboarding" />
          </label>
          <fieldset>
            <legend>Дни тренировок</legend>
            <div class="feature-row">
              <label v-for="(day, index) in weekdays" :key="day" class="feature-choice">
                <input v-model="draft.training_days" type="checkbox" :value="index + 1" />{{ day }}
              </label>
            </div>
          </fieldset>
          <fieldset>
            <legend>Доступное оборудование</legend>
            <div class="feature-row">
              <label class="feature-choice">
                <input type="checkbox" :checked="draft.equipment_ids.length === 0" @change="draft.equipment_ids = []" />Без оборудования
              </label>
              <label v-for="item in profile.data.value?.equipment ?? []" :key="item.id" class="feature-choice">
                <input v-model="draft.equipment_ids" type="checkbox" :value="item.id" />{{ item.name }}
              </label>
            </div>
            <p v-if="profile.data.value && !profile.data.value.equipment.length" class="feature-muted">Справочник пока пуст. Можно продолжить без оборудования.</p>
          </fieldset>
        </fieldset>
        <button class="feature-button" :disabled="saving || profile.loading.value || status === 409 || (onboarding && !draft.training_days.length)" type="submit">
          {{ saving ? 'Сохраняем…' : onboarding ? 'Сохранить и продолжить' : 'Сохранить профиль' }}
        </button>
        <button v-if="status === 409" type="button" class="feature-button secondary" :disabled="profile.loading.value" @click="refreshVersion">
          Перечитать профиль, сохранив мой ввод
        </button>
        <p v-if="saved" class="feature-success" role="status">Профиль сохранён</p>
      </form>
      <RequestState :error="error" :status="status" />
    </div>
  </FeaturePage>
</template>
