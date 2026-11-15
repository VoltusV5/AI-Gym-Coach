<template>
  <platform-layout
    :title="verificationMode ? 'Проверка тренеров' : 'Модерация программ'"
    subtitle="Проверьте содержание перед решением и укажите причину."
  >
    <template #actions><router-link to="/admin">Управление</router-link></template>
    <async-state
      :loading="page.loading.value"
      :error="page.error.value"
      :empty="!page.items.value.length"
      @retry="page.load()"
      ><div class="platform-grid">
        <article v-for="item in page.items.value" :key="item.id" class="platform-card">
          <h2>{{ verificationMode ? `Заявка тренера ${item.trainer_id}` : item.draft.title }}</h2>
          <p>{{ statuses[item.status] || item.status }}</p>
          <template v-if="verificationMode"
            ><p>{{ item.qualification }}</p>
            <ul>
              <li v-for="document in item.document_refs" :key="document">
                <a
                  v-if="safeLink(document)"
                  :href="document"
                  target="_blank"
                  rel="noopener noreferrer"
                  >Подтверждающий материал</a
                ><span v-else>{{ document }}</span>
              </li>
            </ul>
            <p v-if="item.reason">{{ item.reason }}</p></template
          ><template v-else
            ><p>{{ item.draft.description }}</p>
            <p>{{ money(item.draft.price_minor, item.draft.currency) }}</p>
            <details>
              <summary>Тренировочная структура</summary>
              <article v-for="day in item.draft.days" :key="day.id">
                <h3>{{ day.title }}</h3>
                <ul>
                  <li v-for="ex in day.exercises" :key="ex.position">
                    Упражнение {{ ex.exercise_id }} · {{ ex.sets }} × {{ ex.reps_min }}–{{
                      ex.reps_max
                    }}
                    · отдых {{ ex.rest_sec }} сек.
                  </li>
                </ul>
              </article>
            </details></template
          ><button
            v-if="item.status === 'pending' || item.status === 'in_review'"
            @click="selectItem(item)"
          >
            Принять решение
          </button>
        </article>
      </div></async-state
    >
    <button v-if="page.nextCursor.value" :disabled="page.loading.value" @click="page.load(true)">
      Ещё записи
    </button>
    <form v-if="selected" class="platform-panel" @submit.prevent="decide">
      <h2>Решение по {{ verificationMode ? 'заявке' : selected.draft.title }}</h2>
      <label
        >Результат<select v-model="decision">
          <option value="approved">Одобрить</option>
          <option value="rejected">Отклонить</option>
        </select></label
      ><label>Причина<textarea v-model.trim="reason" required maxlength="4000" /></label>
      <div class="platform-actions">
        <button class="primary" :disabled="action.busy.value">Подтвердить решение</button
        ><button type="button" @click="selected = null">Отмена</button>
      </div>
    </form>
    <p v-if="action.error.value" class="platform-error" role="alert">{{ action.error.value }}</p>
    <p v-if="action.success.value" class="platform-success" role="status">
      {{ action.success.value }}
    </p>
  </platform-layout>
</template>
<script setup>
import { useViewLoad } from '@/composables/useViewLoad'
import { computed, ref } from 'vue'
import { useRoute } from 'vue-router'
import { usePage, useAction } from '@/composables/useAsync'
import { money } from '@/utils/platform'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const route = useRoute(),
  verificationMode = computed(() => route.path.endsWith('/verifications')),
  selected = ref(null),
  reason = ref(''),
  decision = ref('approved'),
  action = useAction()
const page = usePage(() =>
  verificationMode.value ? '/api/v2/admin/verifications' : '/api/v2/admin/programs',
)
const statuses = {
  pending: 'Ожидает проверки',
  approved: 'Одобрено',
  rejected: 'Отклонено',
  in_review: 'На проверке',
  draft: 'Черновик',
  published: 'Опубликовано',
  hidden: 'Скрыто',
}
function safeLink(value) {
  try {
    return ['https:', 'http:'].includes(new URL(value).protocol)
  } catch {
    return false
  }
}
function selectItem(item) {
  selected.value = item
  reason.value = ''
}
async function decide() {
  if (!confirm('Сохранить решение администрации?')) return
  const value = await action.run(
    verificationMode.value
      ? `/api/v2/admin/verifications/${selected.value.id}`
      : `/api/v2/admin/programs/${selected.value.id}/moderation`,
    {
      status: decision.value,
      reason: reason.value,
      ...(!verificationMode.value ? { program_version_id: selected.value.version_id } : {}),
    },
    { method: 'patch' },
  )
  if (value) {
    selected.value = null
    await page.load()
  }
}
useViewLoad(() => page.load())
</script>
