import { it, expect, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia } from 'pinia'
import { nextTick } from 'vue'
import App from '@/App.vue'
import NotesEditor from '@/views/notes/NotesEditor.vue'
import { useAuthStore } from '@/stores/auth'
import { useNotesStore } from '@/stores/notes'

vi.mock('vue-router', () => ({
  useRouter: () => ({ push: vi.fn(), replace: vi.fn() }),
  useRoute: () => ({ path: '/notes/note-a', params: { id: 'note-a' } }),
  RouterLink: { template: '<a><slot /></a>' },
  RouterView: { template: '<div />' }
}))
it('removes the previous account text from an already open editor', async () => {
  localStorage.clear()
  const pinia = createPinia()
  const auth = useAuthStore(pinia)
  const notes = useNotesStore(pinia)
  notes.upsertNote('note-a', 'Private title A', 'Private body A')
  const wrapper = mount(App, { global: { plugins: [pinia], stubs: {
    IonApp: { template: '<div><slot /></div>' },
    IonRouterOutlet: { components: { NotesEditor }, template: '<NotesEditor />' },
    IonPage: { template: '<div><slot /></div>' },
    IonContent: { template: '<div><slot /></div>' },
    IonFooter: { template: '<div><slot /></div>' },
    IonButton: { template: '<button><slot /></button>' },
    IonInput: { props: ['modelValue'], template: '<input :value="modelValue" />' },
    IonTextarea: { props: ['modelValue'], template: '<textarea :value="modelValue" />' },
    RouterLink: true,
    AppTabBar: true
  } } })
  await nextTick()
  expect(wrapper.get('input').element.value).toBe('Private title A')
  auth.clearAccountData()
  await nextTick()
  expect(wrapper.get('input').element.value).toBe('')
  expect(wrapper.get('textarea').element.value).toBe('')
  wrapper.unmount()
})
