<template>
  <ion-page class="platform-page">
    <ion-header class="ion-no-border">
      <ion-toolbar>
        <div class="platform-top">
          <router-link to="/home" class="platform-brand"
            ><span class="platform-brand-mark" aria-hidden="true"
              ><ion-icon :icon="barbellOutline" /></span
            >Спортик</router-link
          >
          <router-link
            v-if="!auth.isRegistered"
            :to="{ path: '/login', query: { redirect: route.fullPath } }"
            >Войти</router-link
          >
          <button v-else type="button" class="platform-logout" @click="logout">
            <ion-icon :icon="logOutOutline" aria-hidden="true" />Выйти
          </button>
        </div>
      </ion-toolbar>
    </ion-header>
    <ion-content>
      <div class="platform-shell">
        <nav class="platform-nav" aria-label="Разделы приложения">
          <router-link
            v-for="item in primaryLinks"
            :key="item.to"
            :to="item.to"
            :class="{ 'platform-nav-active': activeLink === item.to }"
            ><ion-icon :icon="item.icon" aria-hidden="true" /><span>{{
              item.label
            }}</span></router-link
          >
          <div class="platform-extra-links platform-extra-links--desktop">
            <router-link
              v-for="item in secondaryLinks"
              :key="item.to"
              :to="item.to"
              :class="{ 'platform-nav-active': activeLink === item.to }"
              ><ion-icon :icon="item.icon" aria-hidden="true" /><span>{{
                item.label
              }}</span></router-link
            >
          </div>
          <details ref="menu" class="platform-extra-menu">
            <summary>
              <ion-icon :icon="menuOutline" aria-hidden="true" /><span>Все разделы</span
              ><ion-icon :icon="chevronDownOutline" aria-hidden="true" />
            </summary>
            <div class="platform-extra-links">
              <router-link
                v-for="item in secondaryLinks"
                :key="item.to"
                :to="item.to"
                :class="{ 'platform-nav-active': activeLink === item.to }"
                ><ion-icon :icon="item.icon" aria-hidden="true" /><span>{{
                  item.label
                }}</span></router-link
              >
            </div>
          </details>
        </nav>
        <main class="platform-main">
          <header class="platform-heading">
            <div>
              <h1>{{ title }}</h1>
              <p v-if="subtitle">{{ subtitle }}</p>
            </div>
            <div v-if="$slots.actions" class="platform-heading-actions">
              <slot name="actions" />
            </div>
          </header>
          <slot />
        </main>
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup>
import { computed, ref, watch } from 'vue'
import { IonPage, IonHeader, IonToolbar, IonContent, IonIcon } from '@ionic/vue'
import {
  homeOutline,
  barbellOutline,
  gridOutline,
  statsChartOutline,
  personOutline,
  heartOutline,
  peopleOutline,
  notificationsOutline,
  receiptOutline,
  documentTextOutline,
  restaurantOutline,
  fitnessOutline,
  schoolOutline,
  walletOutline,
  pieChartOutline,
  settingsOutline,
  menuOutline,
  chevronDownOutline,
  logOutOutline,
} from 'ionicons/icons'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
defineProps({ title: { type: String, required: true }, subtitle: { type: String, default: '' } })
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const menu = ref(null)
const primaryLinks = [
  { to: '/home', label: 'Сегодня', icon: homeOutline },
  { to: '/catalog', label: 'Каталог', icon: gridOutline },
  { to: '/plans', label: 'Мои планы', icon: barbellOutline },
  { to: '/progress', label: 'Прогресс', icon: statsChartOutline },
  { to: '/profile', label: 'Профиль', icon: personOutline },
]
const secondaryLinks = computed(() => [
  { to: '/recommendations', label: 'Рекомендации', icon: fitnessOutline },
  { to: '/favorites', label: 'Избранное', icon: heartOutline },
  { to: '/trainers', label: 'Мои тренеры', icon: peopleOutline },
  { to: '/notifications', label: 'Уведомления', icon: notificationsOutline },
  { to: '/orders', label: 'Покупки', icon: receiptOutline },
  { to: '/notes', label: 'Заметки', icon: documentTextOutline },
  { to: '/nutrition', label: 'Питание', icon: restaurantOutline },
  { to: '/previous-plan', label: 'Прежний план', icon: barbellOutline },
  ...(auth.roles.includes('trainer')
    ? [
        { to: '/trainer', label: 'Кабинет тренера', icon: schoolOutline },
        { to: '/author/programs', label: 'Мои программы', icon: documentTextOutline },
        { to: '/trainer/analytics', label: 'Аналитика тренера', icon: pieChartOutline },
        { to: '/trainer/finance', label: 'Финансы тренера', icon: walletOutline },
      ]
    : []),
  ...(auth.roles.includes('admin')
    ? [{ to: '/admin', label: 'Управление', icon: settingsOutline }]
    : []),
])
const activeLink = computed(() => {
  const path = route.path
  if (/^\/programs(?:\/|$)/.test(path)) return '/catalog'
  if (/^\/training(?:\/|$)/.test(path)) return '/plans'
  if (/^\/workouts(?:\/|$)/.test(path)) return '/progress'
  if (/^\/checkout(?:\/|$)/.test(path)) return '/orders'
  return [...primaryLinks, ...secondaryLinks.value]
    .filter((item) => path === item.to || path.startsWith(item.to + '/'))
    .sort((a, b) => b.to.length - a.to.length)[0]?.to
})
function closeMenu() {
  if (menu.value) menu.value.open = false
}
watch(() => route.fullPath, closeMenu)
async function logout() {
  await auth.logout()
  await router.replace('/login')
}
</script>
