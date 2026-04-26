<template>
  <ion-page class="sportik-page">
    <ion-header class="sportik-header ion-no-border">
      <ion-toolbar class="sportik-toolbar">
        <ion-buttons slot="start">
          <ion-back-button default-href="/" text="" color="dark"></ion-back-button>
        </ion-buttons>
      </ion-toolbar>
    </ion-header>

    <!-- без fullscreen: иначе ion-content рисуется ПОД footer и перехватывает тапы по «Далее» -->
    <ion-content class="sportik-onboarding-content">
      <div class="onboarding-stack">
        <div v-if="progress > 0" class="sportik-progress-strip">
          <ion-progress-bar :value="progress / 100"></ion-progress-bar>
        </div>

        <div class="onboarding-inner ion-padding">
          <p v-if="stepHint" class="sportik-step-hint">{{ stepHint }}</p>
          <h1 v-if="question" class="sportik-question">{{ question }}</h1>
          <slot></slot>

          <!-- Декор сразу под контентом шага, крупнее и плотнее -->
          <div
            v-if="showBottomIllustration && (bottomQuestionUrl || bottomExclamUrl)"
            class="onboarding-deco-row"
          >
            <img
              v-if="bottomQuestionUrl"
              class="onboarding-deco-img onboarding-deco-img--q"
              :src="bottomQuestionUrl"
              alt=""
            />
            <img
              v-if="bottomExclamUrl"
              class="onboarding-deco-img onboarding-deco-img--e"
              :src="bottomExclamUrl"
              alt=""
            />
          </div>
        </div>
      </div>
    </ion-content>

    <ion-footer class="ion-no-border ion-padding sportik-footer" v-if="hasFooter">
      <ion-button
        type="button"
        class="sportik-footer-btn"
        expand="block"
        :disabled="disabled || loading"
        @click="onNextClick"
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
  IonContent,
  IonFooter,
  IonButton,
  IonProgressBar
} from '@ionic/vue'
import {
  getOnboardingBottomIllustrationUrl,
  getOnboardingExclamationIllustrationUrl
} from '@/utils/localImages'

const bottomQuestionUrl = getOnboardingBottomIllustrationUrl()
const bottomExclamUrl = getOnboardingExclamationIllustrationUrl()

const emit = defineEmits(['next'])

const props = defineProps({
  question: {
    type: String,
    default: ''
  },
  stepHint: {
    type: String,
    default: 'Ответьте на вопрос'
  },
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
  },
  showBottomIllustration: {
    type: Boolean,
    default: true
  }
})

function onNextClick() {
  if (props.disabled || props.loading) return
  emit('next')
}
</script>

<style scoped>
.onboarding-stack {
  min-height: 100%;
  display: flex;
  flex-direction: column;
}

.onboarding-inner {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 600px;
  margin: 0 auto;
  padding-bottom: 1.2rem;
  position: relative;
  z-index: 1;
  flex: 0 0 auto;
}

.onboarding-deco-row {
  display: flex;
  flex-direction: row;
  align-items: flex-end;
  justify-content: center;
  gap: 8px;
  margin-top: 1rem;
  max-width: 100%;
}

.onboarding-deco-img {
  display: block;
  height: min(42vw, 200px);
  width: auto;
  max-width: 48%;
  object-fit: contain;
  object-position: center bottom;
}

.onboarding-deco-img--e {
  height: min(36vw, 168px);
  max-width: 40%;
}

.sportik-footer {
  --background: transparent;
  background: linear-gradient(
    to top,
    color-mix(in srgb, var(--sportik-bg) 95%, transparent),
    transparent
  );
  backdrop-filter: blur(8px);
}
</style>
