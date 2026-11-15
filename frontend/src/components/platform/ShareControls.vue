<template>
  <section class="platform-panel">
    <h2>Поделиться планом</h2>
    <p class="platform-muted">
      Ссылка сохраняет текущий снимок. Последующие правки плана её не меняют.
    </p>
    <form @submit.prevent="create">
      <div class="platform-fields">
        <label
          >Содержимое<select v-model="day">
            <option value="">Весь план</option>
            <option v-for="item in days" :key="item.id" :value="item.id">{{ item.title }}</option>
          </select></label
        ><label
          >Срок действия, необязательно<input v-model="expires" type="datetime-local"
        /></label>
      </div>
      <button :disabled="action.busy.value || dirty">Создать публичную ссылку</button>
    </form>
    <p v-if="dirty" class="platform-muted">Сначала сохраните изменения плана.</p>
    <template v-if="share"
      ><p>
        {{
          share.visibility === 'preview'
            ? 'Ссылка покажет только разрешённое превью.'
            : 'Ссылка создана.'
        }}
      </p>
      <label>Ссылка<input :value="share.url" readonly @focus="$event.target.select()" /></label>
      <div class="platform-actions">
        <a :href="share.url" target="_blank" rel="noopener noreferrer">Открыть ссылку</a
        ><button :disabled="action.busy.value" @click="revoke">Отозвать ссылку</button>
      </div></template
    >
    <p v-if="action.error.value" role="alert" class="platform-error">{{ action.error.value }}</p>
  </section>
</template>
<script setup>
import { ref } from 'vue'
import { useAction } from '@/composables/useAsync'
const props = defineProps({
  planId: { type: String, required: true },
  days: { type: Array, required: true },
  dirty: Boolean,
})
const day = ref(''),
  expires = ref(''),
  share = ref(null),
  action = useAction()
async function create() {
  const value = await action.run('/api/v2/shares', {
    target_type: day.value ? 'plan_day' : 'plan',
    target_id: day.value || props.planId,
    expires_at: expires.value ? new Date(expires.value).toISOString() : null,
  })
  if (value) share.value = value
}
async function revoke() {
  if (
    confirm('Отозвать эту публичную ссылку?') &&
    (await action.run(`/api/v2/shares/${share.value.id}`, undefined, { method: 'delete' }))
  )
    share.value = null
}
</script>
