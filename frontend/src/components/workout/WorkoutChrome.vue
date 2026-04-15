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

      <div class="chrome-footer-stack">
        <div v-if="$slots.footer" class="chrome-footer-extra ion-padding">
          <slot name="footer" />
        </div>
        <app-tab-bar :active-key="activeTabKey" />
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup>
import { IonPage, IonContent } from '@ionic/vue'
import { getWorkoutBackgroundImageUrl } from '@/utils/localImages'
import AppTabBar from '@/components/navigation/AppTabBar.vue'

defineProps({
  activeTabKey: {
    type: String,
    default: 'main'
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
  --background: var(--sportik-mint-soft);
}

.chrome-scroll {
  padding-bottom: 0;
  background: var(--sportik-cream);
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
  background: linear-gradient(
    165deg,
    #b8fcff 0%,
    var(--sportik-cyan) 45%,
    #52e8e8 100%
  );
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
  margin-top: -8px;
  background: var(--sportik-cream);
  border-radius: 28px 28px 0 0;
  box-shadow: 0 -10px 40px rgba(0, 0, 0, 0.1);
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
  padding-top: 0.5rem;
  padding-bottom: 0.25rem;
}

.chrome-footer-stack {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 10;
  display: flex;
  flex-direction: column;
  padding-bottom: env(safe-area-inset-bottom, 0px);
  background: var(--sportik-cream);
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.06);
}

.chrome-footer-extra {
  order: 1;
  padding-top: 0.35rem;
}

</style>
