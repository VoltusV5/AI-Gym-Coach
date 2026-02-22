// router/index.js
import { createRouter, createWebHistory } from '@ionic/vue-router'
import Welcome from '@/views/Welcome.vue'

const routes = [
  {
    path: '/',
    name: 'Welcome',
    component: Welcome
  },
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

export default router