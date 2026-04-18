<template>
  <nutrition-chrome title="Поиск блюда" subtitle="каталог и свои блюда" :show-back="true">
    <section class="nutrition-card">
      <ion-input
        v-model="query"
        class="field"
        label="Введите название блюда"
        label-placement="stacked"
        placeholder="Например: омлет"
        @ionInput="onInput"
      />
    </section>
    <section class="nutrition-card list-card">
      <p v-if="!items.length" class="empty">Введите минимум 2 символа</p>
      <button v-for="item in items" :key="item.id" type="button" class="dish-item" @click="pick(item)">
        <p class="dish-title">{{ item.title }}</p>
        <p class="dish-meta">{{ fmt(item) }} на 100г</p>
      </button>
    </section>
  </nutrition-chrome>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { IonInput } from '@ionic/vue'
import NutritionChrome from '@/components/nutrition/NutritionChrome.vue'
import { useNutritionStore } from '@/stores/nutrition'

const store = useNutritionStore()
const router = useRouter()
const query = ref('')
const items = ref([])

function toNum(v) {
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}
function fmt(x) {
  return `Б ${toNum(x.protein_g).toFixed(1)} / Ж ${toNum(x.fat_g).toFixed(1)} / У ${toNum(x.carbs_g).toFixed(1)} / ${toNum(x.calories).toFixed(0)} ккал`
}
async function onInput() {
  if (query.value.trim().length < 2) {
    items.value = []
    return
  }
  items.value = await store.searchDishes(query.value, 40)
}
function pick(item) {
  store.setSelectedDish({
    source: 'catalog',
    listKey: `cat-${item.id}`,
    dish_id: Number(item.id),
    title: item.title,
    protein_g: item.protein_g,
    fat_g: item.fat_g,
    carbs_g: item.carbs_g,
    calories: item.calories
  })
  router.back()
}
</script>

<style scoped>
.nutrition-card {
  background: var(--sportik-surface-soft);
  border: 1px solid var(--sportik-border);
  border-radius: 14px;
  box-shadow: var(--sportik-shadow-md);
  padding: 12px;
  margin-bottom: 10px;
}
.field {
  --background: var(--sportik-surface);
  --color: var(--sportik-text);
  border-radius: 10px;
  border: 1px solid var(--sportik-border);
}
.list-card {
  display: grid;
  gap: 8px;
}
.dish-item {
  text-align: left;
  background: var(--sportik-surface);
  border: 1px solid var(--sportik-border);
  border-radius: 12px;
  padding: 10px;
}
.dish-title {
  margin: 0;
  font-weight: 700;
  color: var(--sportik-text);
}
.dish-meta {
  margin: 2px 0 0;
  color: var(--sportik-text-muted);
  font-size: 0.84rem;
}
.empty {
  margin: 0;
  color: var(--sportik-text-muted);
}
</style>
