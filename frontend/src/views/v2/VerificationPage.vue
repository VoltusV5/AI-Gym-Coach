<script setup>
import { ref, watch } from 'vue'
import { onIonViewWillEnter } from '@ionic/vue'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import { apiStatus, v2 } from '@/api/v2'
import { useRead, useWrite } from '@/composables/useRequest'
import { useAuthStore } from '@/stores/auth'
import { stashDraft, takeDraft } from '@/utils/reauthDrafts'
import { lines, verificationLabels } from '@/utils/trainers'
import { dateLabel } from '@/utils/pages'

const auth = useAuthStore()
const qualification = ref('')
const documents = ref('')
const { data, loading, error, status, load } = useRead(async (signal) => {
  try {
    return await v2('GET', '/trainer/verification', { signal })
  } catch (err) {
    if (apiStatus(err) !== 404) throw err
    return null
  }
})
const { saving, error: submitError, status: submitStatus, run } = useWrite()
onIonViewWillEnter(async () => {
  await load()
  const previous = takeDraft(auth.accountId, '/trainer/verification')
  if (previous) {
    qualification.value = previous.qualification
    documents.value = previous.documents
  }
})
watch(() => auth.token, (token) => {
  if (!token && (qualification.value || documents.value)) {
    stashDraft(auth.lastSubject, '/trainer/verification', { qualification: qualification.value, documents: documents.value })
  }
})

/** Submit private document references for administrator review. */
function submit() {
  return run(async () => {
    data.value = await v2('POST', '/trainer/verification', {
      data: { qualification: qualification.value, document_refs: lines(documents.value) },
    })
    qualification.value = ''
    documents.value = ''
  })
}
</script>

<template>
  <FeaturePage eyebrow="Кабинет тренера" title="Проверка квалификации" subtitle="Заявку и документы видят администраторы. Публичный профиль показывает только отметку подтверждения." back="/trainer/profile">
    <div class="feature-stack">
      <RequestState v-if="!auth.token" error="Сессия истекла. Войдите, чтобы продолжить." :status="401" />
      <RequestState v-else :loading="loading" :error="error" :status="status" :retry="load" />
      <section v-if="auth.token && data" class="feature-card">
        <span class="feature-badge" role="status">{{ verificationLabels[data.status] }}</span>
        <h2>Заявка от {{ dateLabel(data.created_at) }}</h2>
        <p class="feature-text">{{ data.qualification }}</p>
        <p v-if="data.reason" class="feature-text">Решение: {{ data.reason }}</p>
        <button v-if="data.status === 'pending'" type="button" class="feature-button secondary" @click="load">Обновить статус</button>
      </section>
      <form v-if="auth.token && !loading && !error && (!data || data.status === 'rejected')" class="feature-card" @submit.prevent="submit">
        <h2>{{ data ? 'Новая заявка' : 'Подать заявку' }}</h2>
        <fieldset class="feature-stack" :disabled="saving">
          <legend>Документы и опыт</legend>
          <label>Квалификация<textarea v-model.trim="qualification" required maxlength="10000" /></label>
          <label>Ссылки или идентификаторы документов, по одному на строку<textarea v-model.trim="documents" required /></label>
        </fieldset>
        <button class="feature-button" :disabled="saving || !qualification || !lines(documents).length">{{ saving ? 'Отправляем…' : 'Отправить на проверку' }}</button>
        <RequestState :error="submitError" :status="submitStatus" :retry="submitStatus === 409 ? load : null" />
      </form>
    </div>
  </FeaturePage>
</template>
