<template>
  <button v-if="!open" type="button" @click="open = true">Пожаловаться</button>
  <form v-else class="platform-panel" @submit.prevent="submit">
    <label>Причина жалобы<textarea v-model.trim="reason" required maxlength="4000" /></label>
    <p v-if="action.error.value" class="platform-error" role="alert">{{ action.error.value }}</p>
    <p v-if="action.success.value" class="platform-success" role="status">
      {{ action.success.value }}
    </p>
    <div class="platform-actions">
      <button class="primary" :disabled="action.busy.value">Отправить жалобу</button
      ><button type="button" :disabled="action.busy.value" @click="open = false">Закрыть</button>
    </div>
  </form>
</template>
<script setup>
import { ref } from 'vue'
import { useAction } from '@/composables/useAsync'
const props = defineProps({
  targetType: { type: String, required: true },
  targetId: { type: [String, Number], required: true },
})
const open = ref(false),
  reason = ref(''),
  action = useAction()
async function submit() {
  const result = await action.run(
    '/api/v2/reports',
    { target_type: props.targetType, target_id: String(props.targetId), reason: reason.value },
    { message: 'Жалоба отправлена на рассмотрение' },
  )
  if (result) reason.value = ''
}
</script>
