<template>
  <nutrition-chrome title="Питание" subtitle="">
    <div class="nutrition-home">
      <div class="streak-corner">🔥 {{ streakDays }}</div>
      <button type="button" class="progress-corner" @click="go('/nutrition/goals')">🏆 Прогресс</button>

    <section class="nutrition-card">
      <p class="section-title">Сводка</p>
      <div class="summary-grid">
        <div class="ring-wrap">
          <div class="ring" :style="{ '--p': ringPercent }">
            <div class="ring-inner">
              <p class="ring-line">
                <span class="ring-num">{{ consumedCalories.toFixed(0) }}</span>
                <span class="ring-label">съедено</span>
              </p>
              <p class="ring-line">
                <span class="ring-num">{{ remainingCalories.toFixed(0) }}</span>
                <span class="ring-label">осталось</span>
              </p>
            </div>
          </div>
        </div>
        <div class="macro-row">
          <div class="macro">
            <p class="macro-title">Углеводы</p>
            <p class="macro-val">{{ consumedCarbs.toFixed(0) }} / {{ goalCarbs.toFixed(0) }} г</p>
            <div class="bar"><div class="bar-fill" :style="{ width: macroPct(consumedCarbs, goalCarbs) + '%' }"></div></div>
          </div>
          <div class="macro">
            <p class="macro-title">Белки</p>
            <p class="macro-val">{{ consumedProtein.toFixed(0) }} / {{ goalProtein.toFixed(0) }} г</p>
            <div class="bar"><div class="bar-fill" :style="{ width: macroPct(consumedProtein, goalProtein) + '%' }"></div></div>
          </div>
          <div class="macro">
            <p class="macro-title">Жиры</p>
            <p class="macro-val">{{ consumedFat.toFixed(0) }} / {{ goalFat.toFixed(0) }} г</p>
            <div class="bar"><div class="bar-fill" :style="{ width: macroPct(consumedFat, goalFat) + '%' }"></div></div>
          </div>
        </div>
      </div>
    </section>

    <section class="nutrition-card">
      <p class="section-title">Питание</p>
      <div class="meals-grid">
        <button
          v-for="m in meals"
          :key="m.key"
          type="button"
          class="meal-tile"
          @click="openMealPicker(m.key)"
        >
          <span class="meal-tile-plus" aria-hidden="true">+</span>
          <span class="meal-tile-label">{{ m.label }}</span>
          <span class="meal-tile-kcal">{{ mealKcalLine(m) }}</span>
        </button>
      </div>
    </section>

    <section class="nutrition-card water-track" :class="{ 'water-ok': waterGoalReached }">
      <p class="section-title">Отслеживание выпитой воды</p>
      <p class="water-caption">Цель: {{ waterGoalLitersLabel }} л ({{ waterGoalMlRounded }} мл)</p>
      <p class="water-caption">Выпито: {{ waterMl }} мл</p>
      <div class="quick-row bottom">
        <ion-button fill="outline" @click="saveWater(250)">+250</ion-button>
      </div>
      <div class="quick-row top">
        <ion-button fill="outline" @click="saveWater(100)">+100</ion-button>
        <ion-button fill="outline" @click="saveWater(500)">+500</ion-button>
      </div>
    </section>

    <section class="nutrition-card">
      <p class="section-title">Измерение тела</p>
      <div class="weight-row">
        <ion-button fill="outline" @click="changeWeightDraft(-0.1)">-100</ion-button>
        <div class="weight-display">Текущий вес {{ weightKg.toFixed(1) }}</div>
        <ion-button fill="outline" @click="changeWeightDraft(0.1)">+100</ion-button>
      </div>
      <ion-button class="save-weight-btn" expand="block" @click="saveWeightNow">Сохранить</ion-button>
      <p v-if="needWeightReminder" class="reminder">Вы не взвешивались 3 дня. Введите текущий вес.</p>
    </section>

    <section class="nutrition-card progress-card" @click="go('/nutrition/goals')">
      <p class="section-title">Мой прогресс</p>
      <p class="progress-text">Изменение веса: {{ weightDeltaText }}</p>
      <div v-if="weightMiniChart.length" class="mini-chart">
        <div v-for="p in weightMiniChart" :key="p.day" class="mini-col">
          <div class="mini-bar" :style="{ height: `${p.pct}%` }"></div>
          <span class="mini-day">{{ p.day.slice(5) }}</span>
        </div>
      </div>
    </section>
    </div>
  </nutrition-chrome>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { IonButton, toastController, onIonViewWillEnter } from '@ionic/vue'
import { useRouter } from 'vue-router'
import NutritionChrome from '@/components/nutrition/NutritionChrome.vue'
import { useNutritionStore } from '@/stores/nutrition'
import { useAuthStore } from '@/stores/auth'

const nutrition = useNutritionStore()
const authStore = useAuthStore()
const router = useRouter()
const weightKg = ref(70)

onMounted(async () => {
  await nutrition.hydrateAll()
  const lastWeight = nutrition.dashboard?.weight?.last_weight_kg
  const profileWeight = authStore.profile?.weight_kg
  weightKg.value = Number(lastWeight ?? profileWeight ?? 70)
  await nutrition.fetchReports(30)
})

onIonViewWillEnter(async () => {
  await nutrition.hydrateAll()
  const lastWeight = nutrition.dashboard?.weight?.last_weight_kg
  const profileWeight = authStore.profile?.weight_kg
  weightKg.value = Number(lastWeight ?? profileWeight ?? weightKg.value)
})

function toNum(v) {
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}

/** Если дашборд отдал нули, суммируем записи за день (тот же список, что в блоке «Питание»). */
function totalsFromEntries() {
  const list = nutrition.entries || []
  return list.reduce(
    (acc, x) => ({
      calories: acc.calories + toNum(x.calories),
      protein_g: acc.protein_g + toNum(x.protein_g),
      fat_g: acc.fat_g + toNum(x.fat_g),
      carbs_g: acc.carbs_g + toNum(x.carbs_g)
    }),
    { calories: 0, protein_g: 0, fat_g: 0, carbs_g: 0 }
  )
}

const consumedCalories = computed(() => {
  // Список записей за день — единственный надёжный источник после добавления блюда (дашборд мог отставать).
  if (nutrition.entries?.length) return totalsFromEntries().calories
  return toNum(nutrition.dashboard?.today?.calories)
})
const goalCalories = computed(() =>
  toNum(nutrition.dashboard?.goal?.calories) || toNum(nutrition.goal?.calories)
)
const remainingCalories = computed(() => Math.max(0, goalCalories.value - consumedCalories.value))
const consumedProtein = computed(() => {
  if (nutrition.entries?.length) return totalsFromEntries().protein_g
  return toNum(nutrition.dashboard?.today?.protein_g)
})
const consumedFat = computed(() => {
  if (nutrition.entries?.length) return totalsFromEntries().fat_g
  return toNum(nutrition.dashboard?.today?.fat_g)
})
const consumedCarbs = computed(() => {
  if (nutrition.entries?.length) return totalsFromEntries().carbs_g
  return toNum(nutrition.dashboard?.today?.carbs_g)
})
const goalProtein = computed(() =>
  toNum(nutrition.goal?.protein_g) || toNum(nutrition.dashboard?.goal?.protein_g)
)
const goalFat = computed(() => toNum(nutrition.goal?.fat_g) || toNum(nutrition.dashboard?.goal?.fat_g))
const goalCarbs = computed(() => toNum(nutrition.goal?.carbs_g) || toNum(nutrition.dashboard?.goal?.carbs_g))
const ringPercent = computed(() => {
  const g = goalCalories.value || 1
  return `${Math.min(100, Math.max(0, (consumedCalories.value / g) * 100))}%`
})
const streakDays = computed(() => Number(nutrition.dashboard?.streak_days || 0))
const waterMl = computed(() => Number(nutrition.dashboard?.water?.amount_ml || 0))
const waterGoalMl = computed(() => {
  const g = Number(nutrition.dashboard?.water?.goal_ml)
  return Number.isFinite(g) && g > 0 ? g : 2000
})
const waterGoalLitersLabel = computed(() => {
  const L = waterGoalMl.value / 1000
  return L.toFixed(2).replace('.', ',')
})
const waterGoalMlRounded = computed(() => Math.round(waterGoalMl.value))
const waterGoalReached = computed(() => waterMl.value >= waterGoalMl.value)
const dailyCalorieBudget = computed(() => {
  const g = toNum(nutrition.dashboard?.goal?.calories) || toNum(nutrition.goal?.calories)
  return g > 0 ? g : 2000
})
const meals = computed(() => {
  const by = nutrition.entriesByMeal || {}
  const sum = (key) => (by[key] || []).reduce((acc, x) => acc + toNum(x.calories), 0)
  const k = dailyCalorieBudget.value
  const share = { breakfast: 0.26, lunch: 0.34, dinner: 0.32, snack: 0.08 }
  return [
    { key: 'breakfast', label: 'Завтрак', target: Math.round(k * share.breakfast), current: sum('breakfast') },
    { key: 'lunch', label: 'Обед', target: Math.round(k * share.lunch), current: sum('lunch') },
    { key: 'dinner', label: 'Ужин', target: Math.round(k * share.dinner), current: sum('dinner') },
    { key: 'snack', label: 'Перекус', target: Math.max(50, Math.round(k * share.snack)), current: sum('snack') }
  ]
})
const needWeightReminder = computed(() => Boolean(nutrition.dashboard?.weight?.need_weight_reminder))
const weightDeltaText = computed(() => {
  const rows = [...(nutrition.reports?.weight ?? [])].sort((a, b) => String(a.day).localeCompare(String(b.day)))
  if (rows.length === 0) return 'нет данных'
  if (rows.length === 1) {
    const w = Number(rows[0]?.weight_kg ?? 0)
    return `${w.toFixed(1)} кг (одна запись)`
  }
  const first = Number(rows[0]?.weight_kg ?? 0)
  const last = Number(rows[rows.length - 1]?.weight_kg ?? 0)
  const d = last - first
  return `${d > 0 ? '+' : ''}${d.toFixed(1)} кг`
})
const weightMiniChart = computed(() => {
  const sorted = [...(nutrition.reports?.weight ?? [])].sort((a, b) => String(a.day).localeCompare(String(b.day)))
  const rows = sorted.slice(-7)
  if (!rows.length) return []
  const values = rows.map((x) => Number(x.weight_kg || 0))
  const min = Math.min(...values)
  const max = Math.max(...values)
  const span = Math.max(max - min, 0.0001)
  return rows.map((x) => {
    const val = Number(x.weight_kg || 0)
    return {
      day: x.day,
      pct: 20 + ((val - min) / span) * 80
    }
  })
})

function macroPct(current, goal) {
  if (!goal || goal <= 0) return 0
  return Math.min(100, Math.max(0, (current / goal) * 100))
}

function mealKcalLine(m) {
  return `${Number(m.current || 0).toFixed(0)} / ${m.target} ккал`
}

async function toast(message, color = 'success') {
  const t = await toastController.create({ message, duration: 1400, color })
  await t.present()
}

async function saveWater(addMl) {
  await nutrition.saveWater(waterMl.value + Number(addMl || 0))
}

function changeWeightDraft(delta) {
  weightKg.value = Math.max(20, Number((Number(weightKg.value) + Number(delta || 0)).toFixed(1)))
}

async function saveWeightNow() {
  await nutrition.saveWeight(weightKg.value)
  await nutrition.fetchReports(30)
  await toast('Вес обновлен')
}

function openMealPicker(meal) {
  router.push({ path: '/nutrition/add-meal/picker', query: { meal } })
}

function go(path) {
  router.push(path)
}
</script>

<style scoped>
/* Контейнер с position: relative, чтобы бейджи не перекрывали «Сводку», и отступ сверху под них */
.nutrition-home {
  position: relative;
  padding-top: 46px;
}
.nutrition-card { background: var(--sportik-surface-soft); border: 1px solid var(--sportik-border); border-radius: 14px; box-shadow: var(--sportik-shadow-md); padding: 12px; margin-bottom: 10px; }
.streak-corner {
  position: absolute;
  top: 0;
  left: 0;
  z-index: 2;
  margin: 0;
  padding: 6px 10px;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, #ff8a00 55%, var(--sportik-border));
  background: color-mix(in srgb, #ffb347 20%, var(--sportik-surface-soft));
  color: var(--sportik-text);
  font-weight: 700;
  font-size: 0.88rem;
}
.progress-corner {
  position: absolute;
  top: 0;
  right: 0;
  z-index: 2;
  margin: 0;
  padding: 6px 10px;
  border-radius: 999px;
  border: 1px solid color-mix(in srgb, #5f9bff 50%, var(--sportik-border));
  background: color-mix(in srgb, #7bb6ff 18%, var(--sportik-surface-soft));
  color: var(--sportik-text);
  font-weight: 700;
  font-size: 0.84rem;
}
.section-title { margin: 0 0 8px; font-weight: 700; color: var(--sportik-text); }
.summary-grid { display: flex; align-items: center; gap: 10px; }
.ring-wrap { display: grid; place-items: center; }
.ring-wrap { width: 60%; min-width: 0; }
.ring {
  --size: 194px;
  width: var(--size);
  height: var(--size);
  border-radius: 50%;
  background: conic-gradient(var(--sportik-brand) var(--p), color-mix(in srgb, var(--sportik-border) 60%, transparent) 0);
  display: grid;
  place-items: center;
}
.ring-inner {
  width: calc(var(--size) - 22px);
  height: calc(var(--size) - 22px);
  border-radius: 50%;
  background: var(--sportik-surface);
  display: grid;
  place-items: center;
  text-align: center;
  padding: 12px;
}
.ring-line { margin: 0; display: grid; gap: 0; justify-items: center; }
.ring-num { font-size: 1.05rem; line-height: 1.05; font-weight: 800; color: var(--sportik-text); }
.ring-label { font-size: 0.72rem; line-height: 1; color: var(--sportik-text-muted); }
.macro-row { width: 40%; min-width: 0; display: grid; gap: 6px; }
.macro { background: var(--sportik-surface); border: 1px solid var(--sportik-border); border-radius: 10px; padding: 6px; }
.macro-title { margin: 0; font-size: 0.75rem; color: var(--sportik-text-muted); }
.macro-val { margin: 2px 0 4px; color: var(--sportik-text); font-weight: 700; font-size: 0.82rem; }
.bar { height: 6px; background: color-mix(in srgb, var(--sportik-border) 45%, transparent); border-radius: 999px; overflow: hidden; }
.bar-fill { height: 100%; background: linear-gradient(90deg, var(--sportik-brand), var(--sportik-brand-2)); border-radius: 999px; }

.meals-grid {
  position: relative;
  z-index: 3;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  align-content: stretch;
}

.meal-tile {
  position: relative;
  border: none;
  border-radius: 20px;
  aspect-ratio: 1;
  width: 100%;
  min-height: 0;
  padding: 14px 12px 12px;
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  justify-content: flex-start;
  text-align: left;
  cursor: pointer;
  color: var(--sportik-text);
  background: linear-gradient(
    165deg,
    color-mix(in srgb, var(--sportik-brand) 22%, var(--sportik-surface)) 0%,
    var(--sportik-surface) 100%
  );
  border: 1px solid color-mix(in srgb, var(--sportik-brand) 16%, var(--sportik-border));
  box-shadow: var(--sportik-shadow-md);
  transition:
    transform 0.18s ease,
    box-shadow 0.18s ease;
}

.meal-tile:active {
  transform: translateY(1px) scale(0.99);
}

.meal-tile-label {
  margin-top: auto;
  font-size: 1.02rem;
  font-weight: 700;
  line-height: 1.2;
  margin-bottom: 6px;
  padding-right: 1.35rem;
}

.meal-tile-kcal {
  font-size: 0.8rem;
  font-weight: 600;
  color: var(--sportik-text-soft);
  line-height: 1.3;
}

.meal-tile-plus {
  position: absolute;
  top: 10px;
  right: 10px;
  font-size: 1.35rem;
  font-weight: 300;
  color: var(--sportik-brand);
  line-height: 1;
  pointer-events: none;
}
.water-caption { margin: 0 0 6px; color: var(--sportik-text-muted); }
.quick-row { display: grid; gap: 8px; }
.quick-row.top { grid-template-columns: 1fr 1fr; margin-top: 8px; }
.quick-row.bottom { grid-template-columns: 1fr; }
.weight-row { display: grid; grid-template-columns: auto 1fr auto; gap: 8px; align-items: center; }
.weight-display { text-align: center; font-weight: 700; color: var(--sportik-text); border: 1px solid var(--sportik-border); border-radius: 10px; background: var(--sportik-surface); padding: 10px 12px; }
.save-weight-btn { margin-top: 8px; }
.reminder { margin: 8px 0 0; font-size: 0.84rem; color: #c77412; }
.progress-card { cursor: pointer; }
.progress-text { margin: 0; color: var(--sportik-text); font-weight: 700; }
.progress-sub { margin: 4px 0 0; color: var(--sportik-text-muted); font-size: 0.84rem; }
.mini-chart { margin-top: 8px; display: grid; grid-template-columns: repeat(7, 1fr); gap: 6px; align-items: end; min-height: 82px; }
.mini-col { display: grid; gap: 3px; justify-items: center; align-items: end; }
.mini-bar { width: 100%; border-radius: 8px 8px 4px 4px; min-height: 10px; background: linear-gradient(180deg, #6fb4ff, #6a6dff); }
.mini-day { font-size: 0.62rem; color: var(--sportik-text-muted); }
.water-ok { border-color: color-mix(in srgb, #22c55e 70%, var(--sportik-border)); box-shadow: 0 0 0 1px color-mix(in srgb, #22c55e 42%, transparent), var(--sportik-shadow-md); background: color-mix(in srgb, #22c55e 14%, var(--sportik-surface-soft)); }
@media (max-width: 760px) {
  .quick-row.top { grid-template-columns: 1fr 1fr; }
  .weight-row { grid-template-columns: auto 1fr auto; }
  .summary-grid {
    display: flex;
    flex-direction: row;
    align-items: center;
    gap: 8px;
  }
  .ring-wrap { width: 60%; }
  .macro-row { width: 40%; }
  .ring { --size: 132px; }
  .ring-num { font-size: 0.92rem; }
  .ring-label { font-size: 0.64rem; }
  .macro { padding: 5px; }
  .macro-title { font-size: 0.7rem; }
  .macro-val { font-size: 0.76rem; margin-bottom: 3px; }
  .bar { height: 5px; }
}
</style>
