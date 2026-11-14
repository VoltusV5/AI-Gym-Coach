import { it, expect, vi } from 'vitest'
import { defineComponent, h } from 'vue'
import { mount } from '@vue/test-utils'
import { useAction } from '@/composables/useAsync'
import api from '@/api/api'
vi.mock('@/api/api', () => ({ default: { request: vi.fn() }, errorMessage: () => 'Ошибка' }))
it('confirms a successful 204 deletion with an empty Axios body', async () => {
  api.request.mockResolvedValue({ status: 204, data: '' })
  let action
  const wrapper = mount(
    defineComponent({
      setup() {
        action = useAction()
        return () => h('div')
      },
    }),
  )
  const result = await action.run('/api/v2/shares/fixture', undefined, { method: 'delete' })
  expect(result).toEqual({ ok: true })
  expect(action.success.value).toBe('Сохранено')
  wrapper.unmount()
})
