<template>
  <nutrition-chrome title="Выбор блюда" subtitle="" :show-back="true">
    <section class="nutrition-card top-row">
      <ion-input
        :value="query"
        class="field"
        label="Поиск блюда"
        label-placement="stacked"
        placeholder="Нажмите, чтобы открыть поиск"
        readonly
        @click="openSearch"
      />
      <ion-button size="small" @click="openMyDishes">+</ion-button>
    </section>

    <section class="nutrition-card">
      <ion-input v-model="grams" class="field" inputmode="decimal" label="Грамм съедено" label-placement="stacked" />
      <ion-segment v-model="tab">
        <ion-segment-button value="frequent"><ion-label>Частые</ion-label></ion-segment-button>
        <ion-segment-button value="recent"><ion-label>Недавние</ion-label></ion-segment-button>
        <ion-segment-button value="favorites"><ion-label>Избранные</ion-label></ion-segment-button>
      </ion-segment>

      <div class="list">
        <button v-for="item in currentItems" :key="item.listKey" type="button" class="dish-item" :class="{ active: selectedDish?.listKey === item.listKey }" @click="selectedDish = item">
          <p class="title">{{ item.title }}</p>
          <p class="meta">{{ macros(item) }}</p>
        </button>
      </div>
    </section>

    <section class="nutrition-card">
      <ion-button expand="block" :disabled="!selectedDish" @click="done">Готово</ion-button>
    </section>
  </nutrition-chrome>
</template>

<script setup>
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { IonButton, IonInput, IonSegment, IonSegmentButton, IonLabel } from '@ionic/vue'
import NutritionChrome from '@/components/nutrition/NutritionChrome.vue'
import { useNutritionStore } from '@/stores/nutrition'

const store = useNutritionStore()
const route = useRoute()
const router = useRouter()
const mealType = String(route.query.meal || 'breakfast')

const query = ref('')
const tab = ref('frequent')
const grams = ref('100')
const searchItems = ref([])
const selectedDish = ref(null)

function normalizeSelectedDish(d) {
  if (!d) return null
  if (d.listKey) return d
  const did = d.dish_id != null && Number(d.dish_id) > 0 ? Number(d.dish_id) : null
  if (did) {
    return {
      listKey: `cat-${did}`,
      dish_id: did,
      title: d.title,
      protein_g: d.protein_g,
      fat_g: d.fat_g,
      carbs_g: d.carbs_g,
      calories: d.calories
    }
  }
  // Старый pick из поиска: только id блюда каталога (не путать с id избранного).
  if (d.source === 'catalog' && d.id != null && Number(d.id) > 0) {
    const id = Number(d.id)
    return {
      listKey: `cat-${id}`,
      dish_id: id,
      title: d.title,
      protein_g: d.protein_g,
      fat_g: d.fat_g,
      carbs_g: d.carbs_g,
      calories: d.calories
    }
  }
  return { ...d, listKey: d.listKey || `tmp-${String(d.title)}` }
}

onMounted(async () => {
  await store.hydrateAll()
  if (store.selectedDish) selectedDish.value = normalizeSelectedDish(store.selectedDish)
})

watch(
  () => store.selectedDish,
  (dish) => {
    if (dish) {
      selectedDish.value = normalizeSelectedDish(dish)
      query.value = String(dish.title || '')
    }
  }
)

function toNum(v) {
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}
/** БЖУ записи дневника → значения на 100 г (для повторного добавления с новыми граммами). */
function per100FromEntryPortion(e) {
  const g = Math.max(1, toNum(e.grams))
  return {
    protein_g: (toNum(e.protein_g) / g) * 100,
    fat_g: (toNum(e.fat_g) / g) * 100,
    carbs_g: (toNum(e.carbs_g) / g) * 100,
    calories: (toNum(e.calories) / g) * 100
  }
}
function macros(x) {
  return `${toNum(x.protein_g).toFixed(1)}/${toNum(x.fat_g).toFixed(1)}/${toNum(x.carbs_g).toFixed(1)} • ${toNum(x.calories).toFixed(0)} ккал/100г`
}

const frequentItems = computed(() => {
  const m = new Map()
  for (const e of store.entries) {
    if (m.has(e.title)) continue
    const hasDish = e.dish_id != null && Number(e.dish_id) > 0
    const per = per100FromEntryPortion(e)
    m.set(e.title, {
      listKey: hasDish ? `freq-dish-${e.dish_id}` : `freq-${e.id}`,
      dish_id: hasDish ? Number(e.dish_id) : undefined,
      title: e.title,
      protein_g: per.protein_g,
      fat_g: per.fat_g,
      carbs_g: per.carbs_g,
      calories: per.calories
    })
  }
  return Array.from(m.values()).slice(0, 20)
})
const recentItems = computed(() =>
  store.entries.slice(0, 20).map((e) => {
    const hasDish = e.dish_id != null && Number(e.dish_id) > 0
    const per = per100FromEntryPortion(e)
    return {
      listKey: `recent-${e.id}`,
      dish_id: hasDish ? Number(e.dish_id) : undefined,
      title: e.title,
      protein_g: per.protein_g,
      fat_g: per.fat_g,
      carbs_g: per.carbs_g,
      calories: per.calories
    }
  })
)
const favoriteItems = computed(() =>
  store.favorites.map((f) => ({
    listKey: `fav-${f.id}`,
    dish_id: undefined,
    title: f.title,
    protein_g: f.protein_g,
    fat_g: f.fat_g,
    carbs_g: f.carbs_g,
    calories: toNum(f.protein_g) * 4 + toNum(f.fat_g) * 9 + toNum(f.carbs_g) * 4
  }))
)

const currentItems = computed(() => {
  if (query.value.trim().length >= 2) return searchItems.value
  if (tab.value === 'recent') return recentItems.value
  if (tab.value === 'favorites') return favoriteItems.value
  return frequentItems.value
})

async function onSearch() {
  if (query.value.trim().length < 2) {
    searchItems.value = []
    return
  }
  searchItems.value = await store.searchDishes(query.value, 40)
}

function openSearch() {
  router.push('/nutrition/add-meal/search')
}

function openMyDishes() {
  router.push({ path: '/nutrition/my-dishes', query: { meal: mealType } })
}

async function done() {
  if (!selectedDish.value) return
  const sel = selectedDish.value
  const payload = {
    meal_type: mealType,
    grams: Number(grams.value || 100),
    title: sel.title,
    protein_g: sel.protein_g,
    fat_g: sel.fat_g,
    carbs_g: sel.carbs_g,
    calories: sel.calories
  }
  if (sel.dish_id != null && Number(sel.dish_id) > 0) {
    payload.dish_id = Number(sel.dish_id)
  }
  await store.addEntry(payload)
  store.clearSelectedDish()
  router.push('/nutrition')
}
</script>

<style scoped>
.nutrition-card { background: var(--sportik-surface-soft); border: 1px solid var(--sportik-border); border-radius: 14px; box-shadow: var(--sportik-shadow-md); padding: 12px; margin-bottom: 10px; }
.top-row { display: grid; grid-template-columns: 1fr auto; gap: 8px; align-items: center; }
.top-row ion-button { margin: 0; align-self: center; }
.field { --background: var(--sportik-surface); --color: var(--sportik-text); border-radius: 10px; border: 1px solid var(--sportik-border); }
.list { margin-top: 10px; display: grid; gap: 8px; max-height: 42vh; overflow: auto; }
.dish-item { text-align: left; border-radius: 12px; border: 1px solid var(--sportik-border); background: var(--sportik-surface); padding: 10px; }
.dish-item.active { border-color: var(--sportik-brand); box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--sportik-brand) 55%, transparent); }
.title { margin: 0; font-weight: 700; color: var(--sportik-text); }
.meta { margin: 2px 0 0; color: var(--sportik-text-muted); font-size: 0.82rem; }
</style>
