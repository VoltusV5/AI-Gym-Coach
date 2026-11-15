import { afterEach, beforeEach, expect, it, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import MockAdapter from 'axios-mock-adapter'
import api from '@/api/api'
import ProgressPage from '@/views/platform/ProgressPage.vue'
import ProgressMetrics from '@/components/ui/ProgressMetrics.vue'

vi.mock('vue-router', () => ({ useRoute: () => ({ path: '/progress' }) }))
const progress = {
  period: 'month',
  start: '2026-11-01T00:00:00+03:00',
  end: '2026-12-01T00:00:00+03:00',
  time_zone: 'Europe/Moscow',
  workout_count: 6,
  volume_kg: 8775,
  adherence: null,
  streak_weeks: 0,
  weight_change_kg: -1.2000000000000028,
  highlights: [],
}
let mock
beforeEach(() => {
  mock = new MockAdapter(api)
  mock.onGet('/api/v2/progress').reply(200, progress)
  mock.onGet('/api/v2/workouts').reply(200, { items: [], next_cursor: null })
  mock.onGet('/api/v2/progress/summaries').reply(200, { items: [], next_cursor: null })
})
afterEach(() => mock.restore())

it('distinguishes a reduction in body weight from a negative measurement', async () => {
  const wrapper = mount(ProgressPage, {
    global: {
      stubs: {
        PlatformLayout: { template: '<main><slot /></main>' },
        AsyncState: { template: '<div><slot /></div>' },
        RouterLink: true,
      },
    },
  })
  await flushPromises()
  expect(wrapper.text()).toContain('Вес снизился')
  expect(wrapper.text()).toContain('1,2 кг')
  expect(wrapper.text()).not.toContain('0000000000028')
  expect(wrapper.text()).toContain('Объём за период')
  expect(wrapper.text()).toContain('600 кг')
  wrapper.unmount()
})

it('explains period totals consistently in the other progress interface', () => {
  const wrapper = mount(ProgressMetrics, { props: { metrics: progress } })
  expect(wrapper.text()).toContain('Объём за период')
  expect(wrapper.text()).toContain('Вес снизился')
  expect(wrapper.text()).not.toContain('-1,2')
  wrapper.unmount()
})

it('groups repeated record transitions by exercise and keeps the best result', async () => {
  const { default: ProgressHighlights } = await import('@/components/ui/ProgressHighlights.vue')
  mock
    .onGet('/api/v2/exercises')
    .reply(200, { items: [{ id: 1, name: 'Приседания' }], next_cursor: null })
  const wrapper = mount(ProgressHighlights, {
    props: {
      items: [
        { exercise_id: 1, metric: 'working_weight_kg', current: 10 },
        { exercise_id: 1, metric: 'working_weight_kg', current: 22.5 },
        { exercise_id: 1, metric: 'volume_kg', current: 675 },
      ],
    },
    global: { stubs: { RouterLink: { template: '<a><slot /></a>' } } },
  })
  await flushPromises()
  expect(wrapper.findAll('li')).toHaveLength(1)
  expect(wrapper.text()).toContain('Приседания')
  expect(wrapper.text()).toContain('22,5 кг')
  expect(wrapper.text()).not.toContain('10 кг')
  wrapper.unmount()
})
