import { ref, shallowRef, onScopeDispose } from 'vue'
import { apiStatus, errorText } from '@/api/v2'
import { appendPage } from '@/utils/pages'

/**
 * Track the newest read and discard results from superseded or unmounted requests.
 * @param {function(AbortSignal): Promise<unknown>} read
 */
export function useRead(read, missing) {
  const data = shallowRef(null)
  const loading = ref(true)
  const error = ref('')
  const status = ref(0)
  let controller
  onScopeDispose(() => controller?.abort())

  async function load() {
    controller?.abort()
    const current = new AbortController()
    controller = current
    loading.value = true
    error.value = ''
    status.value = 0
    data.value = null
    try {
      const result = await read(current.signal)
      if (!current.signal.aborted) data.value = result
    } catch (err) {
      if (current.signal.aborted) return
      status.value = apiStatus(err)
      error.value = errorText(err, missing)
    } finally {
      if (!current.signal.aborted) loading.value = false
    }
  }

  return { data, loading, error, status, load }
}

/**
 * Track a cursor list with explicit reset, cancellation, deduplication, and retry.
 * @param {function(string|null, AbortSignal): Promise<{items: unknown[], next_cursor: string|null}>} read
 */
export function usePages(read) {
  const items = shallowRef([])
  const cursor = ref(null)
  const loading = ref(true)
  const error = ref('')
  const status = ref(0)
  let controller
  onScopeDispose(() => controller?.abort())

  async function load(reset = false) {
    if (!reset && (loading.value || !cursor.value)) return
    controller?.abort()
    const current = new AbortController()
    controller = current
    if (reset) {
      items.value = []
      cursor.value = null
    }
    const requestedCursor = cursor.value
    loading.value = true
    error.value = ''
    status.value = 0
    try {
      const result = await read(requestedCursor, current.signal)
      if (!current.signal.aborted) {
        items.value = appendPage(items.value, result.items)
        cursor.value = result.next_cursor !== requestedCursor ? result.next_cursor : null
      }
    } catch (err) {
      if (current.signal.aborted) return
      status.value = apiStatus(err)
      if ([401, 403].includes(status.value)) {
        items.value = []
        cursor.value = null
      }
      error.value = errorText(err)
    } finally {
      if (!current.signal.aborted) loading.value = false
    }
  }

  const reset = () => load(true)
  const more = () => load()
  const retry = () => load(!items.value.length)
  return { items, cursor, loading, error, status, reset, more, retry }
}

/** Prevent repeated writes and expose a safe error and the last failure status. */
export function useWrite() {
  const saving = ref(false)
  const error = ref('')
  const status = ref(0)

  async function run(write, describe = errorText) {
    if (saving.value) return
    saving.value = true
    error.value = ''
    status.value = 0
    try {
      await write()
    } catch (err) {
      status.value = apiStatus(err)
      error.value = describe(err)
    } finally {
      saving.value = false
    }
  }

  return { saving, error, status, run }
}
