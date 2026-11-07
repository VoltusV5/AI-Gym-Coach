<script setup>
import { IonIcon } from '@ionic/vue'
import { RouterLink } from 'vue-router'
import { MAIN_TABS } from '@/config/mainTabs'

defineProps({ activeKey: { type: String, default: 'main' } })
</script>

<template>
  <nav class="app-tabbar" aria-label="Основное меню">
    <RouterLink v-for="item in MAIN_TABS" :key="item.key" :to="item.path" class="tab-link"
      :class="{ active: item.key === activeKey }" :aria-current="item.key === activeKey ? 'page' : undefined">
      <ion-icon :icon="item.icon" aria-hidden="true" />
      <span>{{ item.label }}</span>
    </RouterLink>
  </nav>
</template>

<style scoped>
.app-tabbar {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: .2rem;
  padding: .4rem max(.35rem, env(safe-area-inset-right)) max(.45rem, env(safe-area-inset-bottom)) max(.35rem, env(safe-area-inset-left));
  background: var(--sportik-surface);
  border-top: 1px solid var(--sportik-border);
}
.tab-link {
  display: grid; justify-items: center; align-content: center; gap: .18rem;
  min-height: 3rem; border-radius: .7rem; color: var(--sportik-text-soft);
  text-decoration: none; font-size: clamp(.62rem, 2.5vw, .82rem); font-weight: 640;
}
.tab-link.active { color: var(--sportik-text); background: color-mix(in srgb, var(--sportik-brand) 16%, var(--sportik-surface)); }
.tab-link:focus-visible { outline: 3px solid var(--sportik-brand); outline-offset: -3px; }
ion-icon { font-size: 1.3rem; }
</style>
