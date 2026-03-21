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

// Инициализация Mock API
import '@/api/mock'

// Создание приложения
const app = createApp(App)

// Подключение Ionic и роутинга
app.use(IonicVue)
app.use(createPinia())
app.use(router)

// Ionic + ion-router-outlet
router.isReady().then(() => {
  app.mount('#app')
})
