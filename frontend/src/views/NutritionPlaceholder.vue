<template>
  <ion-page class="nutrition-page">
    <ion-content class="nutrition-content" fullscreen>
      <div class="nutrition-frame">
        <div class="nutrition-strip" aria-hidden="true"></div>

        <div class="nutrition-sheet">
          <div class="nutrition-inner ion-padding">
            <h1 class="nutrition-title">Питание</h1>
            <p class="nutrition-subtitle">Раздел пока в разработке</p>

            <section class="nutrition-card">
              <p class="nutrition-text">
                Здесь позже появятся план питания, трекинг приёмов пищи и рекомендации.
              </p>
            </section>
          </div>
        </div>
      </div>

      <div class="nutrition-footer-stack">
        <app-tab-bar active-key="nutrition" />
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup>
defineOptions({ name: 'NutritionPlaceholder' })

import { IonPage, IonContent } from '@ionic/vue'
import AppTabBar from '@/components/navigation/AppTabBar.vue'
</script>

<style scoped>
.nutrition-content {
  --background: var(--sportik-mint-soft);
}

.nutrition-frame {
  min-height: calc(100svh - env(safe-area-inset-bottom, 0px));
  display: flex;
  flex-direction: column;
}

.nutrition-strip {
  flex: 0 0 clamp(108px, 26vw, 152px);
  background: linear-gradient(165deg, #b8fcff 0%, var(--sportik-cyan) 50%, #52e8e8 100%);
}

.nutrition-sheet {
  flex: 1;
  margin-top: -8px;
  border-radius: 28px 28px 0 0;
  background: var(--sportik-cream);
  padding-bottom: calc(120px + env(safe-area-inset-bottom, 0px));
  color: var(--sportik-text);
}

.nutrition-title {
  margin: 0;
  text-align: center;
  font-family: 'Roboto Mono', 'Roboto', monospace;
  font-size: clamp(1.65rem, 5vw, 2.2rem);
}

.nutrition-subtitle {
  margin: 0.4rem 0 1rem;
  text-align: center;
  font-size: 0.9rem;
  color: var(--sportik-text-muted);
}

.nutrition-card {
  background: var(--sportik-card-gray);
  border-radius: 14px;
  padding: 14px;
}

.nutrition-text {
  margin: 0;
  line-height: 1.45;
  color: var(--sportik-text-soft);
}

.nutrition-footer-stack {
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
