<template>
  <platform-layout
    title="Ваш тренировочный профиль"
    subtitle="Эти данные помогают подобрать нагрузку и программу под ваши условия."
  >
    <async-state :loading="load.loading.value" :error="load.error.value" @retry="fetchProfile">
      <form @submit.prevent="save">
        <fieldset :disabled="action.busy.value" class="editor-fields">
          <section class="platform-panel">
            <h2>О вас</h2>
            <div class="platform-fields">
              <label
                >Возраст<input v-model.number="form.age" type="number" min="1" max="120" /></label
              ><label
                >Рост, см<input
                  v-model.number="form.height_cm"
                  type="number"
                  min="1"
                  max="250" /></label
              ><label
                >Вес, кг<input
                  v-model.number="form.weight_kg"
                  type="number"
                  min="1"
                  max="500"
                  step="0.1"
              /></label>
              <label
                >Пол<select v-model="form.gender">
                  <option
                    v-if="form.gender && !['м', 'ж'].includes(form.gender)"
                    :value="form.gender"
                  >
                    {{ { male: 'Мужской', female: 'Женский' }[form.gender] || form.gender }}
                  </option>
                  <option value="">Не указан</option>
                  <option value="м">Мужской</option>
                  <option value="ж">Женский</option>
                </select></label
              >
              <label
                >Активность<select v-model="form.activity_level">
                  <option
                    v-if="
                      form.activity_level &&
                      !['Низкая активность', 'Средняя активность', 'Высокая активность'].includes(
                        form.activity_level,
                      )
                    "
                    :value="form.activity_level"
                  >
                    {{
                      {
                        low: 'Низкая активность',
                        moderate: 'Средняя активность',
                        high: 'Высокая активность',
                      }[form.activity_level] || form.activity_level
                    }}
                  </option>
                  <option value="">Не указана</option>
                  <option>Низкая активность</option>
                  <option>Средняя активность</option>
                  <option>Высокая активность</option>
                </select></label
              >
            </div>
            <label class="check-label"
              ><input v-model="form.injuries_notes" type="checkbox" />Есть ограничения по
              здоровью</label
            >
          </section>
          <section class="platform-panel">
            <h2>Цели и режим</h2>
            <div class="platform-fields">
              <label
                >Цель<select v-model="form.goal">
                  <option v-if="form.goal && !goalNames[form.goal]" :value="form.goal">
                    {{ form.goal }}
                  </option>
                  <option value="">Выберите цель</option>
                  <option v-for="(label, key) in goalNames" :key="key" :value="key">
                    {{ label }}
                  </option>
                </select></label
              >
              <label
                >Уровень<select v-model="form.level">
                  <option value="">Выберите уровень</option>
                  <option v-for="(label, key) in levelNames" :key="key" :value="key">
                    {{ label }}
                  </option>
                </select></label
              >
              <label
                >Продолжительность занятия, мин.<input
                  v-model.number="form.session_duration_min"
                  type="number"
                  min="1"
                  max="300"
              /></label>
            </div>
            <fieldset>
              <legend>Тренировочные дни</legend>
              <label v-for="(day, index) in weekDays" :key="day" class="check-label"
                ><input v-model="form.training_days" type="checkbox" :value="index + 1" />{{
                  day
                }}</label
              >
            </fieldset>
          </section>
          <section class="platform-panel">
            <h2>Оборудование</h2>
            <p v-if="!equipment.length" class="platform-muted">
              Справочник пуст. Сохранение оборудования доступно после его загрузки.
            </p>
            <label v-for="item in equipment" :key="item.id" class="check-label"
              ><input v-model="form.equipment_ids" type="checkbox" :value="item.id" />{{
                item.name
              }}</label
            >
            <p class="platform-muted">
              Не отмечайте оборудование, которого у вас нет. Пустой выбор означает занятия с
              собственным весом.
            </p>
          </section>
          <p v-if="action.error.value" role="alert" class="platform-error">
            {{ action.error.value }}
          </p>
          <p v-if="action.success.value" role="status" class="platform-success">
            {{ action.success.value }}
          </p>
          <div class="platform-actions">
            <button class="primary" :disabled="action.busy.value">Сохранить профиль</button
            ><button type="button" :disabled="action.busy.value || dirty" @click="generate">
              Создать AI-план</button
            ><button type="button" @click="refreshVersion">Проверить изменения на сервере</button>
          </div>
          <p v-if="dirty" class="platform-muted">Сохраните изменения перед генерацией.</p>
        </fieldset>
      </form>
      <section v-if="serverVersion" class="platform-panel">
        <h2>Профиль был обновлён</h2>
        <p>
          Ваши поля сохранены в форме. Перед повторной отправкой сравните их с текущим профилем.
        </p>
        <p>
          Цель: {{ goalNames[serverVersion.goal] || serverVersion.goal }}. Уровень:
          {{ levelNames[serverVersion.level] || serverVersion.level }}. Тренировочных дней:
          {{ serverVersion.training_days.length }}.
        </p>
        <button @click="rebase">Использовать мои поля после сравнения</button>
      </section>
    </async-state>
  </platform-layout>
</template>
<script setup>
import { referenceItems } from '@/api/referenceLists'
import { useViewLoad } from '@/composables/useViewLoad'
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import api, { errorMessage } from '@/api/api'
import { useAuthStore } from '@/stores/auth'
import { useRequest, useAction } from '@/composables/useAsync'
import { goalNames, levelNames, weekDays, clone } from '@/utils/platform'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const auth = useAuthStore(),
  router = useRouter(),
  load = useRequest(),
  action = useAction()
const form = ref({}),
  equipment = ref([]),
  baseline = ref(''),
  serverVersion = ref(null)
const dirty = computed(() => JSON.stringify(form.value) !== baseline.value)
async function fetchProfile() {
  const result = await load.run(async () => {
    const [profile, gear] = await Promise.all([
      auth.fetchProfileV2(),
      referenceItems('/api/v2/equipment'),
    ])
    return { profile, equipment: gear }
  })
  if (!result) return
  form.value = clone(result.profile)
  equipment.value = result.equipment
  baseline.value = JSON.stringify(form.value)
}
async function save() {
  const payload = Object.fromEntries(
    Object.entries(form.value).filter(([, value]) => value !== null && value !== ''),
  )
  const result = await action.run('/api/v2/me/profile', payload, {
    method: 'patch',
    message: 'Профиль сохранён',
  })
  if (result) {
    form.value = result
    auth.profileV2 = result
    baseline.value = JSON.stringify(result)
    await auth.fetchProfile()
  }
}
async function refreshVersion() {
  try {
    serverVersion.value = (await api.get('/api/v2/me/profile')).data
  } catch (cause) {
    action.error.value = errorMessage(cause)
  }
}
function rebase() {
  if (confirm('Записать поля вашей формы поверх текущего профиля при следующем сохранении?')) {
    form.value.version = serverVersion.value.version
    serverVersion.value = null
  }
}
async function generate() {
  const result = await action.run(
    '/api/v2/plans/generate',
    { profile_version: form.value.version },
    { idempotent: true, message: 'План создан' },
  )
  if (result) await router.push(`/plans/${result.plan_id}`)
}
useViewLoad(fetchProfile, () => !dirty.value)
</script>
<style scoped>
fieldset {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  border: 0;
  padding: 0;
  margin: 0;
}
legend {
  margin-bottom: 12px;
}
</style>

<style scoped>
.editor-fields {
  display: block;
  border: 0;
  padding: 0;
  margin: 0;
  min-width: 0;
}
</style>
