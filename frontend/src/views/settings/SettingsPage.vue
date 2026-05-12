<template>
  <workout-chrome active-tab-key="settings">
    <div class="settings-frame">
      <div class="settings-top-row">
        <p class="settings-brand-line" :class="{ 'settings-brand-line--auth': isAuthHint }">{{ brandLine }}</p>
        <button type="button" class="avatar-btn" @click="openProfilePage">
          {{ t.avatar }}
        </button>
      </div>

      <section class="achievements-block">
        <p class="block-title">{{ t.achievements }}</p>
        <div class="achievements-field">
          <span>{{ t.achievementsHint }}</span>
        </div>
      </section>

      <section class="trackers-block">
        <ion-button expand="block" fill="outline" class="trackers-btn" @click="onTrackers">
          {{ t.trackers }}
        </ion-button>
      </section>

      <section class="subscription-block">
        <ion-button expand="block" class="sportik-footer-btn subscription-btn" @click="noActionYet()">
          {{ t.subscription }}
        </ion-button>
      </section>

      <section class="quick-options-row">
        <div class="quick-item quick-item--left">
          <ion-icon :icon="themeOn ? sunnyOutline : moonOutline" />
          <ion-toggle v-model="themeOn" />
        </div>

        <a href="#" class="faq-link" @click.prevent="noActionYet('FAQ')">FAQ</a>

        <div class="quick-item quick-item--right">
          <ion-icon :icon="notificationsOn ? notificationsOutline : notificationsOffOutline" />
          <ion-toggle v-model="notificationsOn" />
        </div>
      </section>

      <p
        class="build-stamp"
        title="Меняется при каждой сборке. Хвост после # — уникален; если не меняется после деплоя, подгружается старый образ (пересоберите frontend в Docker: npm run docker:rebuild)."
      >
        Сборка: {{ sportikBuildId }}
      </p>
    </div>
  </workout-chrome>
</template>

<script setup>
defineOptions({ name: 'SettingsPage' })

import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { IonButton, IonToggle, IonIcon, toastController } from '@ionic/vue'
import { moonOutline, sunnyOutline, notificationsOutline, notificationsOffOutline } from 'ionicons/icons'
import WorkoutChrome from '@/components/workout/WorkoutChrome.vue'
import { useAuthStore } from '@/stores/auth'

// eslint-disable-next-line no-undef -- подставляется Vite define при сборке
const sportikBuildId = __SPORTIK_BUILD_ID__

const router = useRouter()
const authStore = useAuthStore()

const t = {
  subscription: '\u041f\u043e\u0434\u043f\u0438\u0441\u043a\u0430',
  avatar: '\u041f',
  achievements: '\u0414\u043e\u0441\u0442\u0438\u0436\u0435\u043d\u0438\u044f',
  achievementsHint: '\u0417\u0434\u0435\u0441\u044c \u0431\u0443\u0434\u0435\u0442 \u043b\u0435\u043d\u0442\u0430 \u0434\u043e\u0441\u0442\u0438\u0436\u0435\u043d\u0438\u0439 \u0438 \u043d\u0430\u0433\u0440\u0430\u0434.',
  trackers: '\u041f\u043e\u0434\u043a\u043b\u044e\u0447\u0438\u0442\u044c \u0442\u0440\u0435\u043a\u043a\u0435\u0440\u044b',
  trackersToast: '\u041f\u043e\u0434\u043a\u043b\u044e\u0447\u0438\u0442\u044c \u0442\u0440\u0435\u043a\u043a\u0435\u0440\u044b: UI \u0433\u043e\u0442\u043e\u0432, \u0438\u043d\u0442\u0435\u0433\u0440\u0430\u0446\u0438\u044f \u0441 API \u043f\u043e\u0441\u043b\u0435 backend',
  authHint: '\u0410\u0432\u0442\u043e\u0440\u0438\u0437\u0443\u0439\u0442\u0435\u0441\u044c, \u0447\u0442\u043e\u0431\u044b \u0441\u043e\u0445\u0440\u0430\u043d\u0438\u0442\u044c \u043f\u0440\u043e\u0433\u0440\u0435\u0441\u0441',
  brand: '\u0421\u043f\u043e\u0440\u0442\u0438\u043a',
  toastSuffix: '\u0444\u0443\u043d\u043a\u0446\u0438\u043e\u043d\u0430\u043b \u043f\u043e\u044f\u0432\u0438\u0442\u0441\u044f \u043f\u043e\u0437\u0436\u0435'
}

function parseJwtPayload(token) {
  if (!token || typeof token !== 'string') return null
  const parts = token.split('.')
  if (parts.length < 2) return null
  try {
    const b64 = parts[1].replace(/-/g, '+').replace(/_/g, '/')
    const pad = b64.length % 4 === 0 ? '' : '='.repeat(4 - (b64.length % 4))
    return JSON.parse(atob(b64 + pad))
  } catch {
    return null
  }
}

const isAuthHint = computed(() => {
  if (!authStore.profile) return true
  const claims = parseJwtPayload(authStore.token)
  return Boolean(claims && claims.is_anonymous === true)
})

const brandLine = computed(() => (isAuthHint.value ? t.authHint : t.brand))

const themeOn = ref(document.documentElement.classList.contains('dark-theme'))

watch(themeOn, (val) => {
  if (val) {
    document.documentElement.classList.add('dark-theme')
  } else {
    document.documentElement.classList.remove('dark-theme')
  }
})

const notificationsOn = ref(false)

function openProfilePage() {
  router.push('/settings/profile')
}

async function noActionYet() {
  const toast = await toastController.create({
    message: 'Скоро...',
    duration: 1500,
    color: 'medium'
  })
  await toast.present()
}

async function onTrackers() {
  await noActionYet()
}
</script>

<style scoped>
.settings-frame {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: calc(100svh - 220px - env(safe-area-inset-bottom, 0px));
  padding-bottom: 4px;
}

.settings-top-row {
  display: grid;
  grid-template-columns: 1fr auto;
  gap: 10px;
  align-items: center;
}

.settings-brand-line {
  margin: 0;
  font-size: 0.92rem;
  line-height: 1.35;
  font-weight: 600;
  color: var(--sportik-text);
}

.settings-brand-line--auth {
  text-align: center;
}

.avatar-btn {
  border: 1px solid var(--sportik-border);
  width: 46px;
  height: 46px;
  border-radius: 50%;
  background: var(--sportik-surface-soft);
  color: var(--sportik-text);
  font-size: 1.1rem;
  font-weight: 700;
}

.achievements-block {
  flex: 1;
  min-height: 36vh;
  background: var(--sportik-surface-soft);
  border: 1px solid var(--sportik-border);
  border-radius: 16px;
  box-shadow: var(--sportik-shadow-md);
  padding: 12px;
}

.block-title {
  margin: 0 0 10px;
  font-weight: 700;
  color: var(--sportik-text);
}

.trackers-block {
  background: var(--sportik-surface-soft);
  border: 1px solid var(--sportik-border);
  border-radius: 14px;
  box-shadow: var(--sportik-shadow-md);
  padding: 10px 12px;
}

.trackers-btn {
  margin: 0;
}

.subscription-block {
  background: var(--sportik-surface-soft);
  border: 1px solid var(--sportik-border);
  border-radius: 14px;
  box-shadow: var(--sportik-shadow-md);
  padding: 10px 12px;
}

.subscription-btn {
  margin: 0;
}

.achievements-field {
  min-height: calc(100% - 30px);
  border-radius: 12px;
  background: var(--sportik-surface);
  padding: 12px;
  color: var(--sportik-text-soft);
  line-height: 1.4;
}

.quick-options-row {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 10px;
  background: var(--sportik-surface-soft);
  border: 1px solid var(--sportik-border);
  border-radius: 14px;
  box-shadow: var(--sportik-shadow-md);
  padding: 10px 12px;
}

.quick-item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--sportik-text);
}

.quick-item--left {
  justify-self: start;
}

.quick-item--right {
  justify-self: end;
}

.faq-link {
  color: var(--sportik-text);
  text-decoration: underline;
  font-weight: 600;
  justify-self: center;
}

.build-stamp {
  margin-top: auto;
  padding-top: 16px;
  font-size: 0.7rem;
  line-height: 1.3;
  color: var(--sportik-text-muted);
  text-align: center;
  word-break: break-all;
}
</style>
