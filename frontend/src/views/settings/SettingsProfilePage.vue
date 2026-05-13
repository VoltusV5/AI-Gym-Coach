<template>
  <workout-chrome active-tab-key="settings">
    <div class="settings-profile-frame">
      <ion-button fill="clear" class="back-btn" @click="goBack">{{ t.back }}</ion-button>
      <profile-auth-panel class="profile-panel-wrap" />
    </div>
  </workout-chrome>
</template>

<script setup>
defineOptions({ name: 'SettingsProfilePage' })

import { useRouter } from 'vue-router'
import { IonButton } from '@ionic/vue'
import WorkoutChrome from '@/components/workout/WorkoutChrome.vue'
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
.settings-profile-frame {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding-bottom: 4px;
}

.back-btn {
  align-self: flex-start;
  --color: var(--sportik-text-muted);
}

.profile-panel-wrap {
  flex: 1;
}
</style>