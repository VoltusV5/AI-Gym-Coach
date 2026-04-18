<template>
  <nutrition-chrome title="Цели и аналитика" subtitle="" :show-back="true">
    <section class="nutrition-card">
      <p class="block-title">Цель по весу</p>
      <p class="stats-line">Текущий вес: {{ currentWeight.toFixed(1) }} кг</p>
      <p class="stats-line">Желаемый вес: {{ targetWeight.toFixed(1) }} кг</p>
      <ion-range v-model="targetDelta" :min="-15" :max="15" :step="0.5" pin />
      <div class="goal-row">
        <ion-button @click="recalcBySlider">Сохранить цель</ion-button>
      </div>
      <p class="stats-line" v-if="nutrition.goal">
        Цель: {{ fmt(nutrition.goal.protein_g) }}/{{ fmt(nutrition.goal.fat_g) }}/{{ fmt(nutrition.goal.carbs_g) }} г, {{ kcal(nutrition.goal.calories) }} ккал
      </p>
      <p class="stats-line" v-if="nutrition.dashboard">
        Съедено сегодня: {{ kcal(nutrition.dashboard.today?.calories) }} ккал; осталось: {{ kcal(nutrition.dashboard.goal?.remaining_calories) }} ккал
      </p>
    </section>

    <section class="nutrition-card">
      <p class="block-title">Мой прогресс</p>
      <div class="period-grid">
        <button type="button" class="period-btn" :class="{ active: period === 1 }" @click="setPeriod(1)">1 день</button>
        <button type="button" class="period-btn" :class="{ active: period === 7 }" @click="setPeriod(7)">7 дней</button>
        <button type="button" class="period-btn" :class="{ active: period === 30 }" @click="setPeriod(30)">30 дней</button>
        <button type="button" class="period-btn" :class="{ active: period === 3650 }" @click="setPeriod(3650)">Всё время</button>
      </div>
    </section>

    <section class="nutrition-card">
      <p class="block-title">Съедено калорий</p>
      <p v-if="!caloriesChart.length" class="empty-msg">Нет данных</p>
      <div v-else class="chart-list">
        <div v-for="item in caloriesChart" :key="item.day" class="chart-row">
          <span class="day">{{ item.day.slice(5) }}</span>
          <div class="bar-wrap"><div class="bar" :style="{ width: `${item.pct}%` }"></div></div>
          <span class="value">{{ item.value }}</span>
        </div>
      </div>
    </section>

    <section class="nutrition-card">
      <p class="block-title">Вес</p>
      <p v-if="!weightChart.length" class="empty-msg">Нет данных</p>
      <div v-else class="chart-list">
        <div v-for="item in weightChart" :key="item.day" class="chart-row">
          <span class="day">{{ item.day.slice(5) }}</span>
          <div class="bar-wrap"><div class="bar bar--weight" :style="{ width: `${item.pct}%` }"></div></div>
          <span class="value">{{ item.value }}</span>
        </div>
      </div>
    </section>

    <section class="nutrition-card">
      <p class="block-title">Выпито воды</p>
      <div class="water-list">
        <p v-for="item in waterRows" :key="item.day" class="water-item">{{ item.day }} — {{ item.value }} мл</p>
        <p v-if="!waterRows.length" class="water-item">Нет данных</p>
      </div>
    </section>
  </nutrition-chrome>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { IonButton, IonRange, toastController, onIonViewWillEnter } from '@ionic/vue'
import NutritionChrome from '@/components/nutrition/NutritionChrome.vue'
import { useNutritionStore } from '@/stores/nutrition'

const nutrition = useNutritionStore()
const targetDelta = ref(0)
const period = ref(7)

onMounted(async () => {
  await nutrition.hydrateAll()
  await nutrition.fetchGoal()
  await nutrition.fetchReports(period.value)
})

onIonViewWillEnter(async () => {
  await nutrition.hydrateAll()
  await nutrition.fetchReports(period.value)
})

function fmt(v) {
  const n = Number(v)
  return Number.isFinite(n) ? n.toFixed(1) : '0.0'
}
function kcal(v) {
  const n = Number(v)
  return Number.isFinite(n) ? n.toFixed(0) : '0'
}
const currentWeight = computed(() => Number(nutrition.dashboard?.weight?.last_weight_kg || 70))
const targetWeight = computed(() => currentWeight.value + Number(targetDelta.value || 0))
const caloriesChart = computed(() => {
  const rows = nutrition.reports?.food ?? []
  const max = Math.max(...rows.map((x) => Number(x.calories || 0)), 1)
  return rows.slice(0, Math.min(rows.length, period.value)).reverse().map((x) => ({
    day: x.day,
    value: Math.round(Number(x.calories || 0)),
    pct: Math.max(5, (Number(x.calories || 0) / max) * 100)
  }))
})
const weightChart = computed(() => {
  const rows = nutrition.reports?.weight ?? []
  const max = Math.max(...rows.map((x) => Number(x.weight_kg || 0)), 1)
  return rows.slice(0, Math.min(rows.length, period.value)).reverse().map((x) => ({
    day: x.day,
    value: Number(x.weight_kg || 0).toFixed(1),
    pct: Math.max(5, (Number(x.weight_kg || 0) / max) * 100)
  }))
})
const waterRows = computed(() => {
  const rows = nutrition.reports?.water ?? []
  return rows.slice(0, Math.min(rows.length, period.value)).reverse().map((x) => ({
    day: x.day,
    value: Number(x.amount_ml || 0).toFixed(0)
  }))
})

async function setPeriod(next) {
  period.value = next
  await nutrition.fetchReports(next)
}

async function recalcBySlider() {
  try {
    const delta = Number(targetDelta.value || 0)
    let target = 'maintain'
    if (delta < 0) target = 'lose'
    if (delta > 0) target = 'gain'
    await nutrition.recalculateGoal({ target, target_delta_kg: delta })
    await nutrition.fetchDashboard()
    await nutrition.fetchReports(period.value)
    const t = await toastController.create({ message: 'Цель пересчитана', duration: 1200, color: 'success' })
    await t.present()
  } catch (e) {
    const raw = e?.response?.data
    let msg = 'Не удалось пересчитать цель. Проверьте подключение и попробуйте снова.'
    if (typeof raw === 'string' && raw.trim()) {
      msg = raw.trim()
    } else if (raw && typeof raw === 'object' && raw.message) {
      msg = String(raw.message)
    }
    const t = await toastController.create({ message: msg, duration: 2600, color: 'danger' })
    await t.present()
  }
}
</script>

<style scoped>
.nutrition-card { background: var(--sportik-surface-soft); border: 1px solid var(--sportik-border); border-radius: 14px; box-shadow: var(--sportik-shadow-md); padding: 12px; margin-bottom: 10px; }
.block-title { margin: 0 0 8px; font-weight: 700; color: var(--sportik-text); }
.goal-row { display: grid; grid-template-columns: 1fr; gap: 8px; margin-bottom: 8px; }
.stats-line { margin: 0 0 6px; color: var(--sportik-text-muted); }
.period-grid { display: flex; flex-wrap: nowrap; gap: 6px; align-items: stretch; }
.period-btn { flex: 1 1 0; min-width: 0; border-radius: 12px; border: 1px solid var(--sportik-border); background: var(--sportik-surface); color: var(--sportik-text); padding: 8px 4px; font-weight: 600; font-size: 0.72rem; line-height: 1.2; white-space: nowrap; }
.period-btn.active { border-color: var(--sportik-brand); box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--sportik-brand) 55%, transparent); background: color-mix(in srgb, var(--sportik-brand) 14%, var(--sportik-surface)); }
.chart-list { display: grid; gap: 6px; }
.chart-row { display: grid; grid-template-columns: 48px 1fr 54px; gap: 8px; align-items: center; }
.day, .value { font-size: 0.78rem; color: var(--sportik-text-muted); }
.bar-wrap { height: 10px; background: color-mix(in srgb, var(--sportik-border) 42%, transparent); border-radius: 999px; overflow: hidden; }
.bar { height: 100%; background: linear-gradient(90deg, var(--sportik-brand), var(--sportik-brand-2)); border-radius: 999px; }
.bar--weight { background: linear-gradient(90deg, #4da7ff, #7a5cff); }
.water-list { display: grid; gap: 6px; }
.water-item { margin: 0; color: var(--sportik-text-muted); }
.empty-msg { margin: 0; color: var(--sportik-text-muted); font-size: 0.9rem; }
</style>
