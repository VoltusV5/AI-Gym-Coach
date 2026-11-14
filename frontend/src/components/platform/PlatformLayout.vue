<template>
  <ion-page>
    <ion-header class="ion-no-border">
      <ion-toolbar>
        <div class="platform-top">
          <router-link to="/home" class="platform-brand">Спортик</router-link>
          <router-link
            v-if="!auth.isRegistered"
            :to="{ path: '/login', query: { redirect: route.fullPath } }"
            >Войти</router-link
          >
          <button v-else type="button" @click="logout">Выйти</button>
        </div>
      </ion-toolbar>
    </ion-header>
    <ion-content>
      <div class="platform-shell">
        <nav class="platform-nav" aria-label="Разделы приложения">
          <router-link v-for="item in links" :key="item.to" :to="item.to">{{
            item.label
          }}</router-link>
        </nav>
        <main class="platform-main">
          <header class="platform-heading">
            <div>
              <h1>{{ title }}</h1>
              <p v-if="subtitle">{{ subtitle }}</p>
            </div>
            <slot name="actions" />
          </header>
          <slot />
        </main>
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup>
import { computed } from 'vue'
import { IonPage, IonHeader, IonToolbar, IonContent } from '@ionic/vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
defineProps({ title: { type: String, required: true }, subtitle: { type: String, default: '' } })
const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const links = computed(() => [
  { to: '/home', label: 'Сегодня' },
  { to: '/catalog', label: 'Каталог' },
  { to: '/recommendations', label: 'Рекомендации' },
  { to: '/plans', label: 'Мои планы' },
  { to: '/progress', label: 'Прогресс' },
  { to: '/favorites', label: 'Избранное' },
  { to: '/trainers', label: 'Мои тренеры' },
  { to: '/notifications', label: 'Уведомления' },
  { to: '/orders', label: 'Покупки' },
  { to: '/profile', label: 'Профиль' },
  { to: '/notes', label: 'Заметки' },
  { to: '/nutrition', label: 'Питание' },
  { to: '/previous-plan', label: 'Прежний план' },
  ...(auth.roles.includes('trainer')
    ? [
        { to: '/trainer', label: 'Кабинет тренера' },
        { to: '/author/programs', label: 'Мои программы' },
        { to: '/trainer/analytics', label: 'Аналитика тренера' },
        { to: '/trainer/finance', label: 'Финансы тренера' },
      ]
    : []),
  ...(auth.roles.includes('admin') ? [{ to: '/admin', label: 'Управление' }] : []),
])
async function logout() {
  await auth.logout()
  await router.replace('/login')
}
</script>

<style>
.platform-shell {
  max-width: 1180px;
  margin: 0 auto;
  padding: 18px 20px 40px;
  display: grid;
  grid-template-columns: 180px minmax(0, 1fr);
  gap: 30px;
}
.platform-top {
  max-width: 1180px;
  margin: auto;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
}
.platform-brand {
  color: var(--sportik-text);
  text-decoration: none;
  font-size: 1.4rem;
  font-weight: 800;
}
.platform-nav {
  display: flex;
  flex-direction: column;
  gap: 5px;
}
.platform-nav a {
  padding: 10px 14px;
  border-radius: 12px;
  text-decoration: none;
  color: var(--sportik-text-soft);
}
.platform-nav a.router-link-exact-active {
  background: var(--sportik-brand);
  color: white;
}
.platform-main {
  min-width: 0;
}
.platform-heading {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 24px;
}
.platform-heading h1 {
  margin: 0 0 8px;
  font-size: clamp(1.65rem, 4vw, 2.25rem);
  letter-spacing: -0.035em;
}
.platform-heading p {
  margin: 0;
  line-height: 1.5;
  color: var(--sportik-text-muted);
  max-width: 65ch;
}
.platform-main button,
.platform-top button,
.platform-button {
  padding: 11px 17px;
  border: 1px solid var(--sportik-border);
  border-radius: 12px;
  background: var(--sportik-surface);
  color: var(--sportik-text);
  cursor: pointer;
  font: inherit;
}
.platform-main button.primary,
a.platform-button.primary {
  background: var(--sportik-brand);
  color: white;
  border-color: transparent;
}
.platform-main button:disabled {
  opacity: 0.55;
  cursor: wait;
}
.platform-main a {
  color: var(--sportik-brand);
}
.platform-main .platform-button {
  display: inline-block;
  text-decoration: none;
}
.platform-main label {
  display: flex;
  flex-direction: column;
  gap: 7px;
  font-size: 0.9rem;
  color: var(--sportik-text-soft);
}
.platform-main input,
.platform-main select,
.platform-main textarea {
  width: 100%;
  border: 1px solid var(--sportik-border);
  background: var(--sportik-surface);
  color: var(--sportik-text);
  border-radius: 10px;
  min-height: 44px;
  padding: 10px 12px;
  font: inherit;
}
.platform-main input[type='checkbox'] {
  width: 18px;
  min-height: 18px;
}
.platform-main .check-label {
  flex-direction: row;
  align-items: center;
  gap: 9px;
}
.platform-main textarea {
  min-height: 100px;
  resize: vertical;
}
.platform-panel {
  padding: 22px;
  background: var(--sportik-surface);
  border: 1px solid var(--sportik-border);
  border-radius: 18px;
  margin-bottom: 18px;
}
.platform-panel h2 {
  margin: 0 0 15px;
  font-size: 1.2rem;
}
.platform-fields {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(170px, 1fr));
  gap: 16px;
  margin-bottom: 18px;
}
.platform-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  margin-top: 18px;
}
.platform-error {
  color: var(--sportik-danger);
}
.platform-success {
  color: var(--sportik-success);
}
.platform-muted {
  color: var(--sportik-text-muted);
  line-height: 1.5;
}
.platform-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(235px, 1fr));
  gap: 18px;
}
.platform-card {
  border: 1px solid var(--sportik-border);
  border-radius: 18px;
  padding: 20px;
  background: var(--sportik-surface);
}
.platform-card h2,
.platform-card h3 {
  margin: 0 0 10px;
}
.platform-badge {
  display: inline-block;
  padding: 4px 9px;
  border-radius: 8px;
  background: var(--sportik-surface-soft);
  font-size: 0.8rem;
}
.platform-main :focus-visible,
.platform-nav :focus-visible {
  outline: 3px solid var(--sportik-brand);
  outline-offset: 3px;
}
@media (max-width: 740px) {
  .platform-shell {
    grid-template-columns: minmax(0, 1fr);
    padding: 10px 15px 28px;
    gap: 20px;
  }
  .platform-nav {
    flex-direction: row;
    overflow-x: auto;
    white-space: nowrap;
    padding-bottom: 8px;
  }
  .platform-nav a {
    flex-shrink: 0;
  }
  .platform-heading {
    align-items: flex-start;
  }
  .platform-panel {
    padding: 17px;
  }
}
@media (prefers-reduced-motion: reduce) {
  .platform-shell *,
  .platform-shell *::before {
    animation: none !important;
    transition: none !important;
  }
}
</style>
