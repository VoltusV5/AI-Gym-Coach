<script setup>
import { ref, shallowRef, watch } from 'vue'
import { useRouter } from 'vue-router'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import RequestState from '@/components/ui/RequestState.vue'
import { apiStatus, v2 } from '@/api/v2'
import { useWrite } from '@/composables/useRequest'
import { useAuthStore } from '@/stores/auth'
import { stashDraft, takeDraft } from '@/utils/reauthDrafts'
import { newPlan } from '@/utils/plans'

const auth = useAuthStore()
const savedDraft = takeDraft(auth.accountId, 'new-plan')
const title = ref(savedDraft?.title ?? '')
const attempt = shallowRef(savedDraft?.attempt ?? null)
const router = useRouter()
const { saving, error, status, run } = useWrite()
watch(status, (status) => {
  if (status === 401) stashDraft(auth.lastSubject, 'new-plan', { title: title.value, attempt: attempt.value })
})

/** Find a committed plan after a create request with an ambiguous result. */
async function recover(dayId) {
  const visited = new Set()
  let cursor
  do {
    const page = await v2('GET', '/me/plans', { params: { cursor } })
    const match = page.items?.find((item) => item.days.some((day) => day.id === dayId))
    if (match) return match
    cursor = page.next_cursor
    if (visited.has(cursor)) break
    visited.add(cursor)
  } while (cursor)
  return null
}

/** Reuse the first day UUID and reconcile an ambiguous response before retrying. */
function create() {
  return run(async () => {
    const previous = attempt.value
    const body = previous ?? newPlan(title.value)
    attempt.value = body
    const dayId = body.days[0].id
    let plan = previous ? await recover(dayId) : null
    if (!plan) {
      try {
        plan = await v2('POST', '/me/plans', { data: body })
      } catch (err) {
        if (apiStatus(err) >= 400 && apiStatus(err) < 500) {
          plan = previous ? await recover(dayId) : null
          if (!plan) attempt.value = null
        }
        if (!plan) throw err
      }
    }
    attempt.value = null
    await router.replace(`/plans/${plan.id}/edit`)
  })
}
</script>

<template>
  <FeaturePage eyebrow="Мои планы" title="Новый план" subtitle="Дайте плану имя. Дни и упражнения можно настроить на следующем экране." back="/plans">
    <form class="feature-card" @submit.prevent="create">
      <label>Название плана<input v-model.trim="title" maxlength="200" required :disabled="saving || !!attempt" /></label>
      <button class="feature-button" :disabled="saving || !title">{{ saving ? 'Создаём…' : attempt ? 'Проверить и повторить' : 'Создать план' }}</button>
      <RequestState :error="error" :status="status" />
    </form>
  </FeaturePage>
</template>
