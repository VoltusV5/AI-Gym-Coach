<template>
  <ion-page class="settings-page">
    <ion-content class="settings-content" fullscreen>
      <div class="settings-frame ion-padding">
        <div class="settings-top-row">
          <ion-button class="sportik-footer-btn subscription-btn" @click="noActionYet(t.subscription)">
            {{ t.subscription }}
          </ion-button>

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
      </div>

      <div class="settings-footer-stack">
        <app-tab-bar active-key="settings" />
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup>
defineOptions({ name: 'SettingsPage' })

import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { IonPage, IonContent, IonButton, IonToggle, IonIcon, toastController } from '@ionic/vue'
import { moonOutline, sunnyOutline, notificationsOutline, notificationsOffOutline } from 'ionicons/icons'
import AppTabBar from '@/components/navigation/AppTabBar.vue'

const router = useRouter()

const t = {
  subscription: '\u041f\u043e\u0434\u043f\u0438\u0441\u043a\u0430',
  avatar: '\u041f',
  achievements: '\u0414\u043e\u0441\u0442\u0438\u0436\u0435\u043d\u0438\u044f',
  achievementsHint: '\u0417\u0434\u0435\u0441\u044c \u0431\u0443\u0434\u0435\u0442 \u043b\u0435\u043d\u0442\u0430 \u0434\u043e\u0441\u0442\u0438\u0436\u0435\u043d\u0438\u0439 \u0438 \u043d\u0430\u0433\u0440\u0430\u0434.',
  toastSuffix: '\u0444\u0443\u043d\u043a\u0446\u0438\u043e\u043d\u0430\u043b \u043f\u043e\u044f\u0432\u0438\u0442\u0441\u044f \u043f\u043e\u0437\u0436\u0435'
}

const themeOn = ref(false)
const notificationsOn = ref(false)

function openProfilePage() {
  router.push('/settings/profile')
}

async function noActionYet(name) {
  const toast = await toastController.create({
    message: `${name}: ${t.toastSuffix}`,
    duration: 1500,
    color: 'medium'
  })
  await toast.present()
}
</script>

<style scoped>
.settings-content { --background: var(--sportik-cream); }

.settings-frame {
  min-height: calc(100svh - 96px - env(safe-area-inset-bottom, 0px));
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.settings-top-row { display: grid; grid-template-columns: 1fr auto; gap: 10px; align-items: center; }
.subscription-btn { margin: 0; }

.avatar-btn {
  border: none;
  width: 46px;
  height: 46px;
  border-radius: 50%;
  background: var(--sportik-card-gray);
  color: var(--sportik-text);
  font-size: 1.1rem;
  font-weight: 700;
}

.achievements-block {
  flex: 1;
  min-height: 36vh;
  background: var(--sportik-card-gray);
  border-radius: 16px;
  padding: 12px;
}

.block-title { margin: 0 0 10px; font-weight: 700; color: var(--sportik-text); }

.achievements-field {
  min-height: calc(100% - 30px);
  border-radius: 12px;
  background: #efefef;
  padding: 12px;
  color: var(--sportik-text-soft);
  line-height: 1.4;
}

.quick-options-row {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 10px;
  background: var(--sportik-card-gray);
  border-radius: 14px;
  padding: 10px 12px;
}

.quick-item { display: flex; align-items: center; gap: 8px; color: var(--sportik-text); }
.quick-item--left { justify-self: start; }
.quick-item--right { justify-self: end; }

.faq-link { color: var(--sportik-text); text-decoration: underline; font-weight: 600; justify-self: center; }

.settings-footer-stack {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 10;
  background: var(--sportik-cream);
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.06);
  padding-bottom: env(safe-area-inset-bottom, 0px);
}
</style>
