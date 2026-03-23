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

async function bootstrap() {
  if (import.meta.env.VITE_USE_MOCK === 'true') {
    await import('@/api/mock')
  }

  const app = createApp(App)
  app.use(IonicVue)
  app.use(createPinia())
  app.use(router)

  await router.isReady()
  app.mount('#app')
}

bootstrap().catch((e) => {
  console.error('Bootstrap failed:', e)
})
