
import { createApp } from 'vue'
import { IonicVue } from '@ionic/vue'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import '@ionic/vue/css/core.css'
import '@ionic/vue/css/normalize.css'
import '@ionic/vue/css/structure.css'
import '@ionic/vue/css/typography.css'
import '@ionic/vue/css/padding.css'
import '@ionic/vue/css/float-elements.css'
import '@ionic/vue/css/text-alignment.css'
import '@ionic/vue/css/text-transformation.css'
import '@ionic/vue/css/flex-utils.css'
import '@ionic/vue/css/display.css'

import '@/theme/sportik.css'
import '@/theme/features.css'
import api from '@/api/api'
import { useAuthStore } from '@/stores/auth'

async function bootstrap() {
  console.info('[Спортик] сборка', __SPORTIK_BUILD_ID__)
  if (import.meta.env.VITE_USE_MOCK === 'true') {
    await import('@/api/mock')
  }

  const app = createApp(App)
  app.use(IonicVue)
  const pinia = createPinia()
  app.use(pinia)
  const auth = useAuthStore(pinia)
  api.interceptors.response.use(undefined, async (error) => {
    const request = error.config
    const authAction = request?.url?.startsWith('/api/v1/auth/')
    if (error.response?.status === 401 && (!authAction || request.url === '/api/v1/auth/register')) {
      await auth.expire(request?.headers?.get?.('Authorization') ?? request?.headers?.Authorization).catch(() => {})
    }
    throw error
  })
  const initialized = auth.initialize()
  router.beforeEach(async (to) => {
    await initialized
    if (to.meta.session && !auth.token) {
      return { name: 'Login', query: { return_to: to.fullPath } }
    }
  })
  app.use(router)
  await router.isReady()
  app.mount('#app')
}

bootstrap().catch(() => {
  console.error('Не удалось запустить Спортик')
})