<template>
  <nutrition-chrome title="Вода" subtitle="Отдельный трекер воды" :show-back="true">
    <section class="nutrition-card water-track" :class="{ 'water-ok': goalReached }">
      <p class="block-title">Сегодня выпито</p>
      <p class="goal-line">Цель: {{ goalLiters }} л ({{ goalMl }} мл)</p>
      <p class="big">{{ amount }} мл</p>
      <div class="quick-row">
        <ion-button fill="outline" @click="add(250)">+250</ion-button>
        <ion-button fill="outline" @click="add(500)">+500</ion-button>
        <ion-button fill="outline" @click="add(750)">+750</ion-button>
      </div>
      <div class="save-row">
        <ion-input v-model="manual" class="field" label="Вручную, мл" label-placement="stacked" inputmode="numeric" />
        <ion-button @click="saveManual">Сохранить</ion-button>
      </div>
    </section>
  </nutrition-chrome>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { IonButton, IonInput } from '@ionic/vue'
import NutritionChrome from '@/components/nutrition/NutritionChrome.vue'
import { useNutritionStore } from '@/stores/nutrition'

const store = useNutritionStore()
const manual = ref('')
const amount = computed(() => Number(store.dashboard?.water?.amount_ml || 0))
const goalMl = computed(() => {
  const g = Number(store.dashboard?.water?.goal_ml)
  return Number.isFinite(g) && g > 0 ? Math.round(g) : 2000
})
const goalLiters = computed(() => (goalMl.value / 1000).toFixed(2).replace('.', ','))
const goalReached = computed(() => amount.value >= goalMl.value)

onMounted(async () => {
  await store.fetchDashboard()
  manual.value = String(amount.value || '')
})

async function add(ml) {
  const next = amount.value + ml
  manual.value = String(next)
  await store.saveWater(next)
}
async function saveManual() {
  await store.saveWater(Number(manual.value || 0))
}
</script>

<style scoped>
.nutrition-card { background: var(--sportik-surface-soft); border: 1px solid var(--sportik-border); border-radius: 14px; box-shadow: var(--sportik-shadow-md); padding: 12px; margin-bottom: 10px; }
.block-title { margin: 0 0 8px; font-weight: 700; color: var(--sportik-text); }
.goal-line { margin: 0 0 6px; font-size: 0.9rem; color: var(--sportik-text-muted); }
.big { margin: 0 0 10px; font-size: 2rem; font-weight: 800; color: var(--sportik-text); }
.quick-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; margin-bottom: 10px; }
.save-row { display: grid; grid-template-columns: 1fr auto; gap: 8px; align-items: end; }
.field { --background: var(--sportik-surface); --color: var(--sportik-text); border-radius: 10px; border: 1px solid var(--sportik-border); }
.water-ok { border-color: color-mix(in srgb, #22c55e 70%, var(--sportik-border)); box-shadow: 0 0 0 1px color-mix(in srgb, #22c55e 42%, transparent), var(--sportik-shadow-md); background: color-mix(in srgb, #22c55e 14%, var(--sportik-surface-soft)); }
@media (max-width: 760px) { .quick-row, .save-row { grid-template-columns: 1fr; } }
</style>