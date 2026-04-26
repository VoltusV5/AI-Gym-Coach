<template>
  <ion-page class="workout-chrome-page">
    <ion-content class="workout-chrome-content" fullscreen>
      <div class="chrome-scroll">
        <div class="chrome-frame">
          <div v-if="showApollo && apolloSrc" class="chrome-apollo-strip" aria-hidden="true">
            <img class="chrome-apollo-strip-img" :src="apolloSrc" alt="" />
          </div>

          <div class="chrome-sheet" :class="{ 'chrome-sheet--full': !showApollo }">
            <div class="chrome-sheet-inner ion-padding">
              <slot />
            </div>
          </div>
        </div>
      </div>
    </ion-content>

    <ion-footer class="ion-no-border workout-chrome-footer">
      <div class="chrome-footer-stack">
        <div v-if="$slots.footer" class="chrome-footer-extra ion-padding">
          <slot name="footer" />
        </div>
        <app-tab-bar :active-key="activeTabKey" />
      </div>
    </ion-footer>
  </ion-page>
</template>

<script setup>
import { IonPage, IonContent, IonFooter } from '@ionic/vue'
import { getWorkoutBackgroundImageUrl } from '@/utils/localImages'
import AppTabBar from '@/components/navigation/AppTabBar.vue'

defineProps({
  activeTabKey: {
    type: String,
    default: 'workout'
  },
  showApollo: {
    type: Boolean,
    default: true
  }
})

const apolloSrc = getWorkoutBackgroundImageUrl()
</script>

<style scoped>
.workout-chrome-content {
  --background: var(--sportik-bg);
}

.chrome-scroll {
  padding-bottom: 0;
  background: transparent;
}

.chrome-frame {
  --chrome-apollo-h: clamp(124px, 31vw, 176px);
  --chrome-footer-pad: calc(88px + env(safe-area-inset-bottom, 0px));
  display: flex;
  flex-direction: column;
  min-height: calc(100svh - env(safe-area-inset-bottom, 0px));
  width: 100%;
}

.chrome-apollo-strip {
  flex: 0 0 var(--chrome-apollo-h);
  width: 100%;
  overflow: hidden;
  position: relative;
  background: linear-gradient(145deg, var(--sportik-brand), var(--sportik-brand-2));
}

.chrome-apollo-strip-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center 18%;
  display: block;
}

.chrome-sheet {
  flex: 1 1 auto;
  width: 100%;
  margin-top: -16px;
  background: var(--sportik-surface);
  border-radius: 28px 28px 0 0;
  box-shadow: var(--sportik-shadow-lg);
  min-height: calc(100svh - var(--chrome-apollo-h) - env(safe-area-inset-bottom, 0px) + 8px);
  padding-bottom: calc(var(--chrome-footer-pad) + 4px);
  position: relative;
  z-index: 1;
}

.chrome-sheet--full {
  margin-top: 0;
  border-radius: 0;
  min-height: calc(100svh - env(safe-area-inset-bottom, 0px));
}

.chrome-sheet-inner {
  padding-top: 1rem;
  padding-bottom: 0.25rem;
}

.workout-chrome-footer {
  box-shadow: 0 -8px 22px rgba(0, 0, 0, 0.08);
}

.chrome-footer-stack {
  display: flex;
  flex-direction: column;
  padding-bottom: env(safe-area-inset-bottom, 0px);
  background: var(--sportik-surface-glass);
  backdrop-filter: blur(12px);
}

.chrome-footer-extra {
  order: 1;
  padding-top: 0.35rem;
}

</style>
