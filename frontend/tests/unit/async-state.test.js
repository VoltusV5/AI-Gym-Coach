import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import AsyncState from '@/components/platform/AsyncState.vue'
describe('request states', () => {
  it('separates loading, error, empty, and success and allows retry', async () => {
    const wrapper = mount(AsyncState, { props: { loading: true }, slots: { default: 'Result', empty: 'No programs' } })
    expect(wrapper.find('[role="status"]').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('Result')
    await wrapper.setProps({ loading: false, error: 'Unavailable' })
    await wrapper.get('button').trigger('click')
    expect(wrapper.emitted('retry')).toHaveLength(1)
    await wrapper.setProps({ error: '', empty: true })
    expect(wrapper.text()).toBe('No programs')
    await wrapper.setProps({ empty: false })
    expect(wrapper.text()).toBe('Result')
  })
})
