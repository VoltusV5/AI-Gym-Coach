<template>
  <nutrition-chrome title="Добавить приём пищи" subtitle="Современный быстрый ввод" :show-back="true">
    <section class="nutrition-card">
      <div class="head-row">
        <p class="block-title">Новое добавление</p>
        <div class="head-actions">
          <ion-button fill="outline" size="small" @click="openSearch">Поиск блюда</ion-button>
          <ion-button fill="clear" size="small" @click="startCustomDish">Добавить новое блюдо</ion-button>
        </div>
      </div>
      <div class="form-grid">
        <ion-segment v-model="form.mealType">
          <ion-segment-button value="breakfast"><ion-label>Завтрак</ion-label></ion-segment-button>
          <ion-segment-button value="lunch"><ion-label>Обед</ion-label></ion-segment-button>
          <ion-segment-button value="dinner"><ion-label>Ужин</ion-label></ion-segment-button>
          <ion-segment-button value="snack"><ion-label>Перекус</ion-label></ion-segment-button>
        </ion-segment>

        <ion-input
          :value="form.title"
          class="field search-trigger"
          label="Блюдо"
          label-placement="stacked"
          placeholder="Нажмите для полноэкранного поиска"
          readonly
          @click="openSearch"
        />
        <ion-input v-model="form.grams" class="field" inputmode="decimal" label="Съедено, грамм" label-placement="stacked" />

        <div class="triple-row">
          <ion-input v-model="form.protein" class="field" inputmode="decimal" label="Белки на 100г" label-placement="stacked" />
          <ion-input v-model="form.fat" class="field" inputmode="decimal" label="Жиры на 100г" label-placement="stacked" />
          <ion-input v-model="form.carbs" class="field" inputmode="decimal" label="Углеводы на 100г" label-placement="stacked" />
        </div>
        <ion-input v-model="form.calories" class="field" inputmode="decimal" label="Калории на 100г" label-placement="stacked" />
        <ion-button expand="block" :disabled="saving" @click="onAddEntry">Добавить приём пищи</ion-button>
      </div>
    </section>

    <section v-for="section in mealSections" :key="section.key" class="nutrition-card">
      <p class="block-title">{{ section.label }}</p>
      <p v-if="!entriesByMeal[section.key].length" class="empty">Нет записей</p>
      <div v-else class="entries-list">
        <article v-for="e in entriesByMeal[section.key]" :key="e.id" class="entry-item">
          <div class="entry-main">
            <p class="entry-title">{{ e.title }}</p>
            <p class="entry-meta">{{ formatMacros(e) }} | {{ toNum(e.calories).toFixed(0) }} ккал</p>
            <p class="entry-date">{{ toNum(e.grams).toFixed(0) }} г</p>
          </div>
          <div class="entry-actions">
            <ion-button size="small" fill="outline" @click="onCloneToFavorite(e)">В избранное</ion-button>
            <ion-button size="small" color="danger" fill="clear" @click="onDeleteEntry(e.id)">Удалить</ion-button>
          </div>
        </article>
      </div>
    </section>
  </nutrition-chrome>
</template>

<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import {
  IonButton,
  IonInput,
  IonSegment,
  IonSegmentButton,
  IonLabel,
  toastController,
  onIonViewWillEnter
} from '@ionic/vue'
import NutritionChrome from '@/components/nutrition/NutritionChrome.vue'
import { useNutritionStore } from '@/stores/nutrition'

const nutrition = useNutritionStore()
const router = useRouter()
const saving = ref(false)
const form = reactive({ dishId: null, mealType: 'breakfast', title: '', grams: '150', protein: '', fat: '', carbs: '', calories: '' })
const entriesByMeal = computed(() => nutrition.entriesByMeal || { breakfast: [], lunch: [], dinner: [], snack: [] })
const mealSections = [
  { key: 'breakfast', label: 'Завтрак' },
  { key: 'lunch', label: 'Обед' },
  { key: 'dinner', label: 'Ужин' },
  { key: 'snack', label: 'Перекусы' }
]

onMounted(async () => {
  await nutrition.hydrateAll()
  applySelectedDish()
})

onIonViewWillEnter(async () => {
  await nutrition.hydrateAll()
  applySelectedDish()
})

function toNum(v) {
  const n = Number(v)
  return Number.isFinite(n) ? n : 0
}
function formatMacros(x) {
  return `${toNum(x.protein_g).toFixed(1)}/${toNum(x.fat_g).toFixed(1)}/${toNum(x.carbs_g).toFixed(1)} г`
}
function formatDate(v) {
  const d = new Date(v)
  if (Number.isNaN(d.getTime())) return '—'
  return d.toLocaleString('ru-RU')
}
async function showToast(message, color = 'success') {
  const toast = await toastController.create({ message, duration: 1500, color })
  await toast.present()
}
function applySelectedDish() {
  const dish = nutrition.selectedDish
  if (!dish) return
  const catalogId = dish.dish_id ?? dish.id
  form.dishId =
    catalogId != null && catalogId !== '' && Number.isFinite(Number(catalogId)) ? Number(catalogId) : null
  form.title = dish.title
  form.protein = String(dish.protein_g ?? '')
  form.fat = String(dish.fat_g ?? '')
  form.carbs = String(dish.carbs_g ?? '')
  form.calories = String(dish.calories ?? '')
  nutrition.clearSelectedDish()
}
watch(() => nutrition.selectedDish, applySelectedDish)

function openSearch() {
  router.push('/nutrition/add-meal/search')
}
function startCustomDish() {
  form.dishId = null
  form.title = ''
  form.protein = ''
  form.fat = ''
  form.carbs = ''
  form.calories = ''
}
async function onAddEntry() {
  saving.value = true
  try {
    if (form.title && !form.dishId) {
      const savedDish = await nutrition.createDish({
        title: form.title,
        protein_g: form.protein,
        fat_g: form.fat,
        carbs_g: form.carbs,
        calories: form.calories,
        base_grams: 100
      })
      form.dishId = savedDish.id
    }
    await nutrition.addEntry({
      dish_id: form.dishId,
      meal_type: form.mealType,
      grams: form.grams,
      title: form.title,
      protein_g: form.protein,
      fat_g: form.fat,
      carbs_g: form.carbs,
      calories: form.calories
    })
    form.dishId = null
    form.mealType = 'breakfast'
    form.title = ''
    form.grams = '150'
    form.protein = ''
    form.fat = ''
    form.carbs = ''
    form.calories = ''
    await showToast('Приём пищи добавлен')
  } catch {
    await showToast('Ошибка добавления', 'danger')
  } finally {
    saving.value = false
  }
}
async function onDeleteEntry(id) {
  await nutrition.deleteEntry(id)
}
async function onCloneToFavorite(entry) {
  await nutrition.addFavorite({ title: entry.title, protein_g: entry.protein_g, fat_g: entry.fat_g, carbs_g: entry.carbs_g, unit_type: 'gram' })
  await showToast('Добавлено в избранное')
}
</script>

<style scoped>
.nutrition-card { background: var(--sportik-surface-soft); border: 1px solid var(--sportik-border); border-radius: 14px; box-shadow: var(--sportik-shadow-md); padding: 12px; margin-bottom: 10px; }
.head-row { display: flex; justify-content: space-between; align-items: center; gap: 10px; }
.head-actions { display: flex; align-items: center; gap: 4px; }
.block-title { margin: 0 0 8px; font-weight: 700; color: var(--sportik-text); }
.form-grid { display: grid; gap: 10px; }
.triple-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }
.search-trigger { cursor: pointer; }
.field { --background: var(--sportik-surface); --color: var(--sportik-text); border-radius: 10px; border: 1px solid var(--sportik-border); }
.empty { margin: 0; font-size: 0.86rem; color: var(--sportik-text-muted); }
.entries-list { display: grid; gap: 8px; }
.entry-item { display: flex; justify-content: space-between; gap: 8px; background: var(--sportik-surface); border: 1px solid var(--sportik-border); border-radius: 12px; padding: 10px; }
.entry-main { min-width: 0; }
.entry-title { margin: 0; font-weight: 700; color: var(--sportik-text); }
.entry-meta, .entry-date { margin: 2px 0 0; font-size: 0.82rem; color: var(--sportik-text-muted); }
.entry-actions { display: grid; align-content: start; gap: 4px; }
@media (max-width: 760px) { .triple-row { grid-template-columns: 1fr; } }
</style>