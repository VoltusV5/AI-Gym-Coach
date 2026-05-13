<template>
  <ion-page class="nutrition-chrome-page">
    <ion-content class="nutrition-chrome-content" fullscreen>
      <div class="nutrition-scroll">
        <div class="nutrition-frame">
          <div class="nutrition-apollo-strip" aria-hidden="true">
            <img v-if="apolloSrc" class="nutrition-apollo-img" :src="apolloSrc" alt="" />
          </div>
          <div class="nutrition-sheet">
            <div class="nutrition-sheet-inner ion-padding">
              <div class="nutrition-head">
                <ion-button v-if="showBack" fill="clear" size="small" class="back-btn" @click="router.back()">
                  Назад
                </ion-button>
                <h1 class="nutrition-title">{{ title }}</h1>
                <p v-if="subtitle" class="nutrition-subtitle">{{ subtitle }}</p>
              </div>
              <slot />
            </div>
          </div>
        </div>
      </div>
    </ion-content>

    <ion-footer class="ion-no-border nutrition-chrome-footer">
      <div class="nutrition-footer-stack">
        <app-tab-bar active-key="nutrition" />
      </div>
    </ion-footer>
  </ion-page>
</template>

<script setup>
import { IonPage, IonContent, IonFooter, IonButton } from '@ionic/vue'
import { useRouter } from 'vue-router'
import AppTabBar from '@/components/navigation/AppTabBar.vue'
import { getApolloHeaderImageUrl } from '@/utils/localImages'

const apolloSrc = getApolloHeaderImageUrl()

defineProps({
  title: { type: String, default: 'Питание' },
  subtitle: { type: String, default: '' },
  showBack: { type: Boolean, default: false }
})

const router = useRouter()
</script>

<style scoped>
.nutrition-chrome-content { --background: var(--sportik-bg); }
.nutrition-scroll {
  padding-bottom: 0;
  background: transparent;
}
.nutrition-frame {
  --nutrition-apollo-h: clamp(124px, 31vw, 176px);
  --nutrition-footer-pad: calc(120px + env(safe-area-inset-bottom, 0px));
  min-height: calc(100svh - env(safe-area-inset-bottom, 0px));
  width: 100%;
  display: flex;
  flex-direction: column;
}
.nutrition-apollo-strip {
  flex: 0 0 var(--nutrition-apollo-h);
  width: 100%;
  overflow: hidden;
  position: relative;
  background: linear-gradient(145deg, var(--sportik-brand), var(--sportik-brand-2));
}
.nutrition-apollo-img {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  object-fit: cover;
  object-position: center 18%;
  display: block;
}
.nutrition-sheet {
  flex: 1 1 auto;
  width: 100%;
  min-width: 0;
  margin-top: -16px;
  position: relative;
  z-index: 1;
  border-radius: var(--sportik-radius-xl) var(--sportik-radius-xl) 0 0;
  background: var(--sportik-surface);
  min-height: calc(100svh - var(--nutrition-apollo-h) - env(safe-area-inset-bottom, 0px) + 8px);
  padding-bottom: calc(var(--nutrition-footer-pad) + 4px);
  color: var(--sportik-text);
  box-shadow: var(--sportik-shadow-lg);
  transform: translateZ(0);
}
.nutrition-sheet-inner {
  padding-top: 1rem;
  padding-bottom: 0.25rem;
}
.nutrition-head { margin-bottom: 10px; }
.back-btn { margin-left: -10px; margin-bottom: 2px; }
.nutrition-title { margin: 0; text-align: center; font-size: clamp(1.45rem, 5vw, 2rem); font-weight: 700; }
.nutrition-subtitle { margin: 0.4rem 0 0.8rem; text-align: center; font-size: 0.9rem; color: var(--sportik-text-muted); }
.nutrition-chrome-footer {
  box-shadow: 0 -8px 22px rgba(0, 0, 0, 0.08);
}
.nutrition-footer-stack {
  background: var(--sportik-surface-glass);
  backdrop-filter: blur(12px);
  padding-bottom: env(safe-area-inset-bottom, 0px);
}
</style>