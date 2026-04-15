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
  --background: var(--sportik-bg);
}

.nutrition-frame {
  min-height: calc(100svh - env(safe-area-inset-bottom, 0px));
  display: flex;
  flex-direction: column;
}

.nutrition-strip {
  flex: 0 0 clamp(108px, 26vw, 152px);
  background: linear-gradient(145deg, var(--sportik-brand) 0%, var(--sportik-brand-2) 100%);
}

.nutrition-sheet {
  flex: 1;
  margin-top: -14px;
  border-radius: 28px 28px 0 0;
  background: var(--sportik-surface);
  padding-bottom: calc(120px + env(safe-area-inset-bottom, 0px));
  color: var(--sportik-text);
  box-shadow: var(--sportik-shadow-lg);
}

.nutrition-title {
  margin: 0;
  text-align: center;
  font-size: clamp(1.45rem, 5vw, 2rem);
  font-weight: 700;
}

.nutrition-subtitle {
  margin: 0.4rem 0 1rem;
  text-align: center;
  font-size: 0.9rem;
  color: var(--sportik-text-muted);
}

.nutrition-card {
  background: var(--sportik-surface-soft);
  border: 1px solid var(--sportik-border);
  border-radius: 14px;
  box-shadow: var(--sportik-shadow-md);
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
  background: var(--sportik-surface-glass);
  box-shadow: 0 -8px 22px rgba(0, 0, 0, 0.1);
  backdrop-filter: blur(12px);
  padding-bottom: env(safe-area-inset-bottom, 0px);
}
</style>
