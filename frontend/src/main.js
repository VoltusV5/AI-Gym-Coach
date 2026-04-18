// main.js
import { createApp } from 'vue'
import { IonicVue } from '@ionic/vue'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'

// Импорт стилей для Ionic
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
import { useAuthStore } from '@/stores/auth'

async function bootstrap() {
  // подставляется Vite define при сборке
  // eslint-disable-next-line no-undef -- Vite define
  console.info('[Спортик] сборка', __SPORTIK_BUILD_ID__)
  if (import.meta.env.VITE_USE_MOCK === 'true') {
    await import('@/api/mock')
  }

  const app = createApp(App)
  app.use(IonicVue)
  const pinia = createPinia()
  app.use(pinia)
  app.use(router)

  await router.isReady()
  // До mount: дочерние onMounted ещё не выполнялись. Иначе питание/другие экраны
  // дергают API без токена (родительский App.vue onMounted идёт ПОСЛЕ дочерних).
  await useAuthStore(pinia).initialize()

  app.mount('#app')
}

bootstrap().catch((e) => {
  console.error('Bootstrap failed:', e)
})
