<template>
  <ion-page>
    <ion-content fullscreen class="sportik-welcome-content welcome-ion-content">
      <div class="welcome-bg" aria-hidden="true">
        <img
          v-if="apolloLeftUrl"
          class="welcome-bg-apollo"
          :src="apolloLeftUrl"
          alt=""
        />
      </div>

      <div class="welcome-wrap ion-padding">
        <div class="welcome-zone welcome-zone--top">
          <div class="welcome-badge">
            <span class="welcome-badge-text">Твой ИИ-тренер в зале</span>
          </div>

          <div v-if="heroPhotoUrl" class="welcome-hero-photo">
            <img class="welcome-hero-photo-img" :src="heroPhotoUrl" alt="" />
          </div>
        </div>

        <div class="welcome-zone welcome-zone--mid">
          <p class="welcome-lead">
            Персональный AI-план: веса и нагрузка автоматически подстраиваются под тебя
          </p>
        </div>

        <div class="welcome-zone welcome-zone--bot">
          <ion-button class="sportik-footer-btn welcome-cta" expand="block" @click="startOnboarding">
            Далее
          </ion-button>

          <p class="welcome-sub">Займёт меньше минуты</p>

        </div>
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup>
defineOptions({ name: 'WelcomePage' })

import { IonPage, IonContent, IonButton } from '@ionic/vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { getWelcomeApolloLeftUrl, getWelcomeHeroPhotoUrl } from '@/utils/localImages'

const router = useRouter()
const authStore = useAuthStore()

const apolloLeftUrl = getWelcomeApolloLeftUrl()
const heroPhotoUrl = getWelcomeHeroPhotoUrl()

const startOnboarding = () => {
  const step = authStore.nextOnboardingStep
  if (step === 'Home') {
    router.push('/home')
    return
  }
  const name = step === 'Welcome' ? 'BodyMetrics' : step
  router.push({ name })
}

const newTestSession = async () => {
  try {
    await authStore.restartSessionForTesting()
    await router.replace('/')
  } catch (e) {
    console.error(e)
  }
}
</script>

<style scoped>
.welcome-ion-content {
  --overflow: hidden;
}

.welcome-bg {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
  overflow: hidden;
}

/* Аполлон справа, компактный; якорь — правый нижний угол */
.welcome-bg-apollo {
  position: absolute;
  right: 0;
  left: auto;
  bottom: 0;
  width: min(95vw, 400px);
  max-height: 90%;
  object-fit: contain;
  object-position: right bottom;
  opacity: 0.48;
  filter: saturate(1.05);
  /* было 0.6; +~22% ≈ 0.73 */
  transform: scale(0.73);
  transform-origin: right bottom;
}

.welcome-wrap {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  min-height: 100%;
  max-width: 520px;
  margin: 0 auto;
  padding-top: max(0.25rem, env(safe-area-inset-top, 0px));
  padding-bottom: max(0.5rem, env(safe-area-inset-bottom, 0px));
}

/* Верх: бейдж + фото ≈ 4/10 экрана */
.welcome-zone--top {
  flex: 4 1 0;
  min-height: min(42dvh, 40svh);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  padding-top: 0.25rem;
}

.welcome-zone--mid {
  flex: 3 1 0;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 0.5rem 0;
}

.welcome-zone--bot {
  flex: 3 1 0;
  min-height: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  gap: 0.35rem;
  padding-bottom: 0.25rem;
}

.welcome-badge {
  background: var(--sportik-surface);
  border: 1px solid var(--sportik-border);
  border-radius: var(--sportik-radius-pill);
  padding: 0.8rem 1.85rem;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.06);
}

.welcome-badge-text {
  font-family: 'Roboto', sans-serif;
  font-weight: 700;
  font-size: clamp(1.05rem, 3.2vw, 1.2rem);
  color: var(--sportik-text);
}

.welcome-hero-photo {
  width: 100%;
  max-width: min(94vw, 420px);
  margin: 0.5rem 0 0;
  flex: 1 1 auto;
  min-height: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.welcome-hero-photo-img {
  display: block;
  width: 100%;
  height: auto;
  max-height: min(30vh, 280px);
  object-fit: contain;
  object-position: center;
  border-radius: 20px;
  border: 1px solid var(--sportik-border);
  box-shadow: var(--sportik-shadow-md);
}

.welcome-lead {
  font-family: 'Roboto', sans-serif;
  font-weight: 700;
  font-size: clamp(1.15rem, 3.5vw, 1.4rem);
  line-height: 1.35;
  color: var(--sportik-text-muted);
  margin: 0;
  max-width: 22rem;
  text-align: center;
}

.welcome-sub {
  font-family: 'Roboto', sans-serif;
  font-weight: 500;
  font-size: 0.95rem;
  color: var(--sportik-text-muted);
  margin: 0;
  text-align: center;
}

.welcome-cta {
  width: 100%;
  max-width: 340px;
  margin: 0;
}

.welcome-reset-test {
  --color: var(--sportik-text-muted);
  font-size: 0.85rem;
  margin: 0;
  text-transform: none;
}
</style>
