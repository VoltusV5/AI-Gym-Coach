import { ref, onScopeDispose } from 'vue'
import { onIonViewWillLeave, onIonViewWillEnter } from '@ionic/vue'
import api, { errorMessage } from '@/api/api'

export function useRequest() {
  const loading = ref(false)
  const error = ref('')
  const data = ref(null)
  let sequence = 0
  onScopeDispose(() => {
    sequence += 1
  })
  async function run(request) {
    const current = ++sequence
    loading.value = true
    error.value = ''
    try {
      const result = await request()
      if (current !== sequence) return null
      data.value = result
      return result
    } catch (cause) {
      if (current === sequence) error.value = errorMessage(cause)
      return null
    } finally {
      if (current === sequence) loading.value = false
    }
  }
  return { data, loading, error, run }
}

export function usePage(endpoint, params = () => ({})) {
  const request = useRequest()
  const items = ref([])
  const nextCursor = ref(null)
  let appliedParams = {}
  let appliedEndpoint
  async function load(append = false) {
    if (append && (request.loading.value || !nextCursor.value)) return
    const queryParams = append ? appliedParams : JSON.parse(JSON.stringify(params()))
    const requestEndpoint = append
      ? appliedEndpoint
      : typeof endpoint === 'function'
        ? endpoint()
        : endpoint
    const value = await request.run(async () => {
      const { data } = await api.get(requestEndpoint, {
        params: { ...queryParams, limit: 20, cursor: append ? nextCursor.value : undefined },
      })
      if (!Array.isArray(data?.items)) throw new Error('Некорректный ответ списка')
      return data
    })
    if (!value) return
    appliedParams = queryParams
    appliedEndpoint = requestEndpoint
    const existing = new Set(append ? items.value.map((item) => item.id) : [])
    items.value = [
      ...(append ? items.value : []),
      ...value.items.filter((item) => !existing.has(item.id)),
    ]
    nextCursor.value = value.next_cursor ?? null
  }
  return { ...request, items, nextCursor, load }
}

export function useAction() {
  const busy = ref(false)
  const error = ref('')
  const success = ref('')
  const attempts = new Map()
  let alive = true
  let epoch = 0
  onIonViewWillLeave(() => {
    alive = false
    epoch += 1
    busy.value = false
  })
  onIonViewWillEnter(() => {
    alive = true
  })
  onScopeDispose(() => {
    alive = false
  })
  async function run(
    path,
    body,
    { method = 'post', idempotent = false, message = 'Сохранено' } = {},
  ) {
    if (busy.value) return null
    busy.value = true
    error.value = ''
    success.value = ''
    const current = epoch
    const signature = JSON.stringify([method, path, body])
    if (idempotent && !attempts.has(signature)) attempts.set(signature, crypto.randomUUID())
    try {
      const { data, status } = await api.request({
        method,
        url: path,
        data: body,
        headers: idempotent ? { 'Idempotency-Key': attempts.get(signature) } : {},
      })
      if (!alive || current !== epoch) return null
      attempts.delete(signature)
      success.value = message
      return status === 204 ? { ok: true } : (data ?? { ok: true })
    } catch (cause) {
      if (alive && current === epoch) error.value = errorMessage(cause)
      return null
    } finally {
      if (alive && current === epoch) busy.value = false
    }
  }
  return { busy, error, success, run }
}
