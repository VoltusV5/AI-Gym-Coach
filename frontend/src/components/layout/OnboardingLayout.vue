<template>
  <ion-page>
    <ion-header class="ion-no-border">
      <ion-toolbar>
        <ion-buttons slot="start">
          <ion-back-button default-href="/"></ion-back-button>
        </ion-buttons>
        <ion-title v-if="title">{{ title }}</ion-title>
        <slot name="header-right"></slot>
      </ion-toolbar>
    </ion-header>

    <ion-content class="ion-padding">
      <div class="onboarding-container">
        <div class="progress-info" v-if="progress > 0">
          <ion-progress-bar :value="progress / 100"></ion-progress-bar>
        </div>
        <slot></slot>
      </div>
    </ion-content>

    <ion-footer class="ion-no-border ion-padding" v-if="hasFooter">
      <ion-button
        expand="block"
        :disabled="disabled"
        :loading="loading"
        @click="$emit('next')"
      >
        {{ nextLabel }}
      </ion-button>
    </ion-footer>
  </ion-page>
</template>

<script setup>
import {
  IonPage,
  IonHeader,
  IonToolbar,
  IonButtons,
  IonBackButton,
  IonTitle,
  IonContent,
  IonFooter,
  IonButton,
  IonProgressBar
} from '@ionic/vue'

defineProps({
  title: String,
  progress: {
    type: Number,
    default: 0
  },
  nextLabel: {
    type: String,
    default: 'Далее'
  },
  disabled: {
    type: Boolean,
    default: false
  },
  loading: {
    type: Boolean,
    default: false
  },
  hasFooter: {
    type: Boolean,
    default: true
  }
})

defineEmits(['next'])
</script>

<style scoped>
.onboarding-container {
  display: flex;
  flex-direction: column;
  width: 100%;
}

.progress-info {
  margin-bottom: 2rem;
  width: 100%;
}

ion-progress-bar {
  border-radius: 4px;
}
</style>
