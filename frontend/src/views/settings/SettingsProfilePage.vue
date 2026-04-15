<template>
  <ion-page class="settings-profile-page">
    <ion-content class="settings-profile-content" fullscreen>
      <div class="settings-profile-frame ion-padding">
        <ion-button fill="clear" class="back-btn" @click="goBack">{{ t.back }}</ion-button>
        <profile-auth-panel class="profile-panel-wrap" />
      </div>

      <div class="settings-footer-stack">
        <app-tab-bar active-key="settings" />
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup>
defineOptions({ name: 'SettingsProfilePage' })

import { useRouter } from 'vue-router'
import { IonPage, IonContent, IonButton } from '@ionic/vue'
import AppTabBar from '@/components/navigation/AppTabBar.vue'
import ProfileAuthPanel from '@/components/settings/ProfileAuthPanel.vue'

const router = useRouter()

const t = {
  back: '\u041d\u0430\u0437\u0430\u0434'
}

function goBack() {
  router.push('/settings')
}
</script>

<style scoped>
.settings-profile-content { --background: var(--sportik-cream); }

.settings-profile-frame {
  min-height: calc(100svh - 96px - env(safe-area-inset-bottom, 0px));
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.back-btn {
  align-self: flex-start;
  --color: var(--sportik-text-muted);
}

.profile-panel-wrap {
  flex: 1;
}

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
