<template>
  <nutrition-chrome title="Избранные блюда" subtitle="Заготовки для быстрого добавления" :show-back="true">
    <section class="nutrition-card">
      <p class="block-title">Новое избранное</p>
      <div class="form-grid">
        <ion-input v-model="form.title" class="field" label="Название" label-placement="stacked" />
        <div class="triple-row">
          <ion-input v-model="form.protein" class="field" inputmode="decimal" label="Белки, г" label-placement="stacked" />
          <ion-input v-model="form.fat" class="field" inputmode="decimal" label="Жиры, г" label-placement="stacked" />
          <ion-input v-model="form.carbs" class="field" inputmode="decimal" label="Углеводы, г" label-placement="stacked" />
        </div>
        <ion-button @click="onAddFavorite">Добавить в избранное</ion-button>
      </div>
    </section>

    <section class="nutrition-card">
      <p class="block-title">Список</p>
      <p v-if="!nutrition.favorites.length" class="empty">Пока пусто</p>
      <div v-else class="favorites-list">
        <article v-for="f in nutrition.favorites" :key="f.id" class="fav-item">
          <div class="entry-main">
            <p class="entry-title">{{ f.title }}</p>
            <p class="entry-meta">{{ macros(f) }} | {{ calories(f) }} ккал</p>
          </div>
          <div class="entry-actions">
            <ion-button size="small" fill="outline" @click="toDiary(f)">+ В дневник</ion-button>
            <ion-button size="small" color="danger" fill="clear" @click="remove(f.id)">Удалить</ion-button>
          </div>
        </article>
      </div>
    </section>
  </nutrition-chrome>
</template>

<script setup>
import { onMounted, reactive } from 'vue'
import { IonButton, IonInput, toastController } from '@ionic/vue'
import NutritionChrome from '@/components/nutrition/NutritionChrome.vue'
import { useNutritionStore } from '@/stores/nutrition'

const nutrition = useNutritionStore()
const form = reactive({ title: '', protein: '', fat: '', carbs: '' })

onMounted(async () => {
  await nutrition.hydrateAll()
})

function toNum(v) {
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}
function macros(x) {
  return `${toNum(x.protein_g).toFixed(1)}/${toNum(x.fat_g).toFixed(1)}/${toNum(x.carbs_g).toFixed(1)} г`
}
function calories(x) {
  const kcal = toNum(x.protein_g) * 4 + toNum(x.fat_g) * 9 + toNum(x.carbs_g) * 4
  return kcal.toFixed(0)
}
async function toast(message, color = 'success') {
  const t = await toastController.create({ message, duration: 1200, color })
  await t.present()
}
async function onAddFavorite() {
  try {
    await nutrition.addFavorite({ title: form.title, protein_g: form.protein, fat_g: form.fat, carbs_g: form.carbs, unit_type: 'gram' })
    form.title = ''
    form.protein = ''
    form.fat = ''
    form.carbs = ''
    await toast('Добавлено')
  } catch {
    await toast('Ошибка добавления', 'danger')
  }
}
async function toDiary(fav) {
  await nutrition.addEntry({ title: fav.title, protein_g: fav.protein_g, fat_g: fav.fat_g, carbs_g: fav.carbs_g })
  await toast('Добавлено в дневник')
}
async function remove(id) {
  await nutrition.deleteFavorite(id)
  await toast('Удалено')
}
</script>

<style scoped>
.nutrition-card { background: var(--sportik-surface-soft); border: 1px solid var(--sportik-border); border-radius: 14px; box-shadow: var(--sportik-shadow-md); padding: 12px; margin-bottom: 10px; }
.block-title { margin: 0 0 8px; font-weight: 700; color: var(--sportik-text); }
.form-grid { display: grid; gap: 10px; }
.triple-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }
.field { --background: var(--sportik-surface); --color: var(--sportik-text); border-radius: 10px; border: 1px solid var(--sportik-border); }
.favorites-list { display: grid; gap: 8px; }
.fav-item { display: flex; justify-content: space-between; gap: 8px; background: var(--sportik-surface); border: 1px solid var(--sportik-border); border-radius: 12px; padding: 10px; }
.entry-title { margin: 0; font-weight: 700; color: var(--sportik-text); }
.entry-meta { margin: 2px 0 0; font-size: 0.82rem; color: var(--sportik-text-muted); }
.entry-actions { display: grid; align-content: start; gap: 4px; }
.empty { margin: 0; font-size: 0.86rem; color: var(--sportik-text-muted); }
@media (max-width: 760px) { .triple-row { grid-template-columns: 1fr; } }
</style>
