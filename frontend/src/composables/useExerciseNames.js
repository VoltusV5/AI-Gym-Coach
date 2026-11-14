import { ref } from 'vue'
import api from '@/api/api'
export function useExerciseNames() {
  const names = ref({}),
    error = ref('')
  async function load(ids = []) {
    error.value = ''
    let cursor
    const needed = new Set(ids)
    try {
      do {
        const { data } = await api.get('/api/v2/exercises', { params: { limit: 100, cursor } })
        for (const item of data.items) {
          names.value[item.id] = item.name
          needed.delete(item.id)
        }
        cursor = data.next_cursor
      } while (cursor && needed.size)
    } catch {
      error.value = 'Не удалось загрузить названия упражнений'
    }
  }
  function name(id) {
    return names.value[id] || 'Упражнение из плана'
  }
  return { names, error, load, name }
}
