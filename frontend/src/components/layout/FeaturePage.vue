<script setup>
import { IonContent, IonPage } from '@ionic/vue'
import { RouterLink } from 'vue-router'
import AppTabBar from '@/components/navigation/AppTabBar.vue'

defineProps({
  eyebrow: { type: String, required: true },
  title: { type: String, required: true },
  subtitle: { type: String, default: '' },
  back: { type: [String, Object], default: '' },
  tab: { type: String, default: '' },
})
</script>

<template>
  <ion-page>
    <ion-content class="feature-content">
      <main class="feature">
        <RouterLink v-if="back" :to="back" class="feature-back">← Назад</RouterLink>
        <header class="feature-hero">
          <div>
            <p class="feature-eyebrow">{{ eyebrow }}</p>
            <h1>{{ title }}</h1>
            <p v-if="subtitle" class="feature-subtitle">{{ subtitle }}</p>
          </div>
          <slot name="actions" />
        </header>
        <slot />
      </main>
    </ion-content>
    <AppTabBar v-if="tab" :active-key="tab" />
  </ion-page>
</template>
