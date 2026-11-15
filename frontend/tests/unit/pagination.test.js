import { it, expect, vi } from 'vitest'
import { defineComponent, h, reactive } from 'vue'
import { mount } from '@vue/test-utils'
import api from '@/api/api'
import { usePage } from '@/composables/useAsync'
import { referenceItems } from '@/api/referenceLists'
vi.mock('@/api/api', () => ({ default: { get: vi.fn() }, errorMessage: () => 'Ошибка' }))

it('keeps the applied filter snapshot until the user submits new filters', async () => {
  const filter = reactive({ level: 'beginner', equipment: ['barbell'] })
  let page
  const wrapper = mount(
    defineComponent({
      setup() {
        page = usePage('/programs', () => ({ ...filter }))
        return () => h('div')
      },
    }),
  )
  api.get.mockResolvedValueOnce({ data: { items: [{ id: 'a' }], next_cursor: 'first' } })
  await page.load()
  filter.level = 'advanced'
  filter.equipment.push('bench')
  api.get.mockResolvedValueOnce({ data: { items: [{ id: 'b' }], next_cursor: null } })
  await page.load(true)
  expect(api.get.mock.lastCall[1].params).toEqual({
    level: 'beginner',
    equipment: ['barbell'],
    limit: 20,
    cursor: 'first',
  })
  expect(page.items.value.map(({ id }) => id)).toEqual(['a', 'b'])
  api.get.mockResolvedValueOnce({ data: { items: [{ id: 'c' }], next_cursor: null } })
  await page.load()
  expect(api.get.mock.lastCall[1].params.level).toBe('advanced')
  expect(page.items.value.map(({ id }) => id)).toEqual(['c'])
  wrapper.unmount()
})

it('loads categories beyond the first hundred and stops on repeated cursors', async () => {
  api.get.mockResolvedValueOnce({
    data: { items: Array.from({ length: 100 }, (_, id) => ({ id })), next_cursor: 'next' },
  })
  api.get.mockResolvedValueOnce({ data: { items: [{ id: 100 }], next_cursor: null } })
  expect(await referenceItems('/categories')).toHaveLength(101)
  expect(api.get.mock.lastCall[1].params.cursor).toBe('next')
  api.get.mockResolvedValue({ data: { items: [], next_cursor: 'cycle' } })
  await expect(referenceItems('/categories')).rejects.toThrow('целиком')
})
