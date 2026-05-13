<template>
  <nutrition-chrome title="Каталог продуктов" subtitle="Поиск и быстрые действия" :show-back="true">
    <section class="nutrition-card">
      <div class="search-row">
        <ion-input v-model="query" class="field" label="Поиск продукта" label-placement="stacked" placeholder="Например: курица" />
        <ion-button fill="outline" @click="loadCatalog">Найти</ion-button>
      </div>
    </section>

    <section class="nutrition-card">
      <p class="block-title">Продукты</p>
      <div class="catalog-list">
        <article v-for="(c, idx) in nutrition.catalog" :key="`${c.title}-${idx}`" class="catalog-item">
          <div>
            <p class="entry-title">{{ c.title }}</p>
            <p class="entry-meta">{{ macros(c) }}</p>
          </div>
          <div class="entry-actions">
            <ion-button size="small" fill="outline" @click="addEntry(c)">+ В дневник</ion-button>
            <ion-button size="small" fill="clear" @click="addFavorite(c)">☆ В избранное</ion-button>
          </div>
        </article>
      </div>
    </section>
  </nutrition-chrome>
</template>

<script setup>
import { onMounted, ref } from 'vue'
import { IonButton, IonInput, toastController } from '@ionic/vue'
import NutritionChrome from '@/components/nutrition/NutritionChrome.vue'
import { useNutritionStore } from '@/stores/nutrition'

const nutrition = useNutritionStore()
const query = ref('')

onMounted(loadCatalog)

function toNum(v) {
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}
function macros(x) {
  return `${toNum(x.protein_g).toFixed(1)}/${toNum(x.fat_g).toFixed(1)}/${toNum(x.carbs_g).toFixed(1)} г`
}
async function toast(message, color = 'success') {
  const t = await toastController.create({ message, duration: 1200, color })
  await t.present()
}
async function loadCatalog() {
  await nutrition.fetchCatalog(query.value, 200)
}
async function addEntry(item) {
  await nutrition.addEntry({ title: item.title, protein_g: item.protein_g, fat_g: item.fat_g, carbs_g: item.carbs_g })
  await toast('Добавлено в дневник')
}
async function addFavorite(item) {
  await nutrition.addFavorite({ title: item.title, protein_g: item.protein_g, fat_g: item.fat_g, carbs_g: item.carbs_g, unit_type: 'gram' })
  await toast('Добавлено в избранное')
}
</script>

<style scoped>
.nutrition-card { background: var(--sportik-surface-soft); border: 1px solid var(--sportik-border); border-radius: 14px; box-shadow: var(--sportik-shadow-md); padding: 12px; margin-bottom: 10px; }
.search-row { display: grid; grid-template-columns: 1fr auto; gap: 8px; align-items: end; }
.field { --background: var(--sportik-surface); --color: var(--sportik-text); border-radius: 10px; border: 1px solid var(--sportik-border); }
.block-title { margin: 0 0 8px; font-weight: 700; color: var(--sportik-text); }
.catalog-list { display: grid; gap: 8px; }
.catalog-item { display: flex; justify-content: space-between; gap: 8px; background: var(--sportik-surface); border: 1px solid var(--sportik-border); border-radius: 12px; padding: 10px; }
.entry-title { margin: 0; font-weight: 700; color: var(--sportik-text); }
.entry-meta { margin: 2px 0 0; font-size: 0.82rem; color: var(--sportik-text-muted); }
.entry-actions { display: grid; gap: 4px; align-content: start; }
@media (max-width: 760px) { .search-row { grid-template-columns: 1fr; } }
</style>