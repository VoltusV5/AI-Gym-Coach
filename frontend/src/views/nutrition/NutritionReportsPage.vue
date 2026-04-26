<template>
  <nutrition-chrome title="Отчеты" subtitle="Динамика питания, воды и веса" :show-back="true">
    <section class="nutrition-card controls">
      <ion-segment v-model="days" @ionChange="load">
        <ion-segment-button value="7"><ion-label>7 дней</ion-label></ion-segment-button>
        <ion-segment-button value="30"><ion-label>30 дней</ion-label></ion-segment-button>
        <ion-segment-button value="90"><ion-label>90 дней</ion-label></ion-segment-button>
      </ion-segment>
    </section>

    <section class="nutrition-card">
      <p class="block-title">Калории по дням</p>
      <p v-if="!food.length" class="empty">Нет данных</p>
      <div v-else class="list">
        <p v-for="item in food" :key="`f-${item.day}`" class="line">{{ item.day }}: {{ num(item.calories, 0) }} ккал</p>
      </div>
    </section>

    <section class="nutrition-card">
      <p class="block-title">Вес</p>
      <p v-if="!weights.length" class="empty">Нет данных</p>
      <div v-else class="list">
        <p v-for="item in weights" :key="`w-${item.day}`" class="line">{{ item.day }}: {{ num(item.weight_kg, 1) }} кг</p>
      </div>
    </section>

    <section class="nutrition-card">
      <p class="block-title">Вода</p>
      <p v-if="!water.length" class="empty">Нет данных</p>
      <div v-else class="list">
        <p v-for="item in water" :key="`wa-${item.day}`" class="line">{{ item.day }}: {{ num(item.amount_ml, 0) }} мл</p>
      </div>
    </section>
  </nutrition-chrome>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { IonLabel, IonSegment, IonSegmentButton } from '@ionic/vue'
import NutritionChrome from '@/components/nutrition/NutritionChrome.vue'
import { useNutritionStore } from '@/stores/nutrition'

const nutrition = useNutritionStore()
const days = ref('30')

const food = computed(() => nutrition.reports?.food ?? [])
const weights = computed(() => nutrition.reports?.weight ?? [])
const water = computed(() => nutrition.reports?.water ?? [])

function num(v, digits = 1) {
  const n = Number(v)
  return Number.isFinite(n) ? n.toFixed(digits) : (0).toFixed(digits)
}

async function load() {
  await nutrition.fetchReports(Number(days.value))
}

onMounted(load)
</script>

<style scoped>
.nutrition-card { background: var(--sportik-surface-soft); border: 1px solid var(--sportik-border); border-radius: 14px; box-shadow: var(--sportik-shadow-md); padding: 12px; margin-bottom: 10px; }
.block-title { margin: 0 0 8px; font-weight: 700; color: var(--sportik-text); }
.empty { margin: 0; font-size: 0.86rem; color: var(--sportik-text-muted); }
.list { display: grid; gap: 4px; }
.line { margin: 0; color: var(--sportik-text-soft); }
</style>
