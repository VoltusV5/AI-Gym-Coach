<template>
  <ion-page class="nutrition-page">
    <ion-content class="nutrition-content" fullscreen>
      <div class="nutrition-frame">
        <div class="nutrition-strip" aria-hidden="true"></div>

        <div class="nutrition-sheet">
          <div class="nutrition-inner ion-padding">
            <h1 class="nutrition-title">Питание</h1>
            <p class="nutrition-subtitle">Журнал приёмов пищи, цели БЖУ и избранные блюда</p>

            <section class="nutrition-card">
              <p class="block-title">Фильтры</p>
              <div class="filters-row">
                <ion-input v-model="filters.search" class="field" label="Поиск" label-placement="stacked" placeholder="Название блюда" />
                <ion-input v-model="filters.dateFrom" class="field" type="date" label="Дата с" label-placement="stacked" />
                <ion-input v-model="filters.dateTo" class="field" type="date" label="Дата по" label-placement="stacked" />
              </div>
              <div class="pager-row">
                <ion-button fill="outline" @click="applyFilters">Применить</ion-button>
                <ion-button fill="clear" @click="resetFilters">Сбросить</ion-button>
              </div>
            </section>

            <section class="nutrition-card">
              <p class="block-title">Добавить приём пищи</p>
              <div class="form-grid">
                <ion-input v-model="form.title" class="field" label="Название" label-placement="stacked" placeholder="Курица с рисом" />
                <div class="triple-row">
                  <ion-input v-model="form.protein" class="field" inputmode="decimal" label="Белки, г" label-placement="stacked" />
                  <ion-input v-model="form.fat" class="field" inputmode="decimal" label="Жиры, г" label-placement="stacked" />
                  <ion-input v-model="form.carbs" class="field" inputmode="decimal" label="Углеводы, г" label-placement="stacked" />
                </div>
                <ion-button class="sportik-footer-btn" expand="block" :disabled="saving" @click="onAddEntry">
                  Добавить
                </ion-button>
              </div>
            </section>

            <section class="nutrition-card">
              <p class="block-title">Цели и статистика</p>
              <div class="goal-row">
                <ion-button fill="outline" @click="onRecalcGoal('lose')">Цель: похудение</ion-button>
                <ion-button fill="outline" @click="onRecalcGoal('maintain')">Поддержание</ion-button>
                <ion-button fill="outline" @click="onRecalcGoal('gain')">Набор</ion-button>
              </div>
              <p class="stats-line" v-if="goal">
                Цель: {{ goal.protein_g.toFixed(1) }}/{{ goal.fat_g.toFixed(1) }}/{{ goal.carbs_g.toFixed(1) }} г,
                {{ goal.calories.toFixed(0) }} ккал
              </p>
              <p class="stats-line" v-else>Цель не рассчитана</p>
              <p class="stats-line" v-if="stats">
                Сегодня: {{ stats.day.protein_g.toFixed(1) }}/{{ stats.day.fat_g.toFixed(1) }}/{{ stats.day.carbs_g.toFixed(1) }} г
              </p>
              <p class="stats-line" v-if="stats">
                Среднее 7 дней: {{ stats.week_avg.protein_g.toFixed(1) }}/{{ stats.week_avg.fat_g.toFixed(1) }}/{{ stats.week_avg.carbs_g.toFixed(1) }} г
              </p>
              <p class="stats-line" v-if="stats">
                Среднее 30 дней: {{ stats.month_avg.protein_g.toFixed(1) }}/{{ stats.month_avg.fat_g.toFixed(1) }}/{{ stats.month_avg.carbs_g.toFixed(1) }} г
              </p>
            </section>

            <section class="nutrition-card">
              <p class="block-title">Приёмы пищи</p>
              <p v-if="!entries.length" class="empty">Пока нет записей</p>
              <div v-else class="entries-list">
                <article v-for="e in entries" :key="e.id" class="entry-item">
                  <div class="entry-main">
                    <p class="entry-title">{{ e.title }}</p>
                    <p class="entry-meta">{{ formatMacros(e) }}</p>
                    <p class="entry-date">{{ formatDate(e.consumed_at) }}</p>
                  </div>
                  <div class="entry-actions">
                    <ion-button size="small" fill="outline" @click="onStartEdit(e)">Редактировать</ion-button>
                    <ion-button size="small" fill="outline" @click="onCloneToFavorite(e)">В избранное</ion-button>
                    <ion-button size="small" color="danger" fill="clear" @click="onDeleteEntry(e.id)">Удалить</ion-button>
                  </div>
                </article>
              </div>
              <div class="pager-row">
                <ion-button fill="outline" :disabled="page <= 1" @click="changePage(page - 1)">Назад</ion-button>
                <p class="pager-text">Страница {{ page }} / {{ pageCount }}</p>
                <ion-button fill="outline" :disabled="page >= pageCount" @click="changePage(page + 1)">Вперёд</ion-button>
              </div>
            </section>

            <section class="nutrition-card">
              <p class="block-title">Избранное</p>
              <p v-if="!favorites.length" class="empty">Пока пусто</p>
              <div v-else class="favorites-list">
                <article v-for="f in favorites" :key="f.id" class="fav-item">
                  <div class="entry-main">
                    <p class="entry-title">{{ f.title }}</p>
                    <p class="entry-meta">{{ formatMacros(f) }}</p>
                  </div>
                  <div class="entry-actions">
                    <ion-button size="small" fill="outline" @click="onAddFromFavorite(f)">+ В дневник</ion-button>
                    <ion-button size="small" color="danger" fill="clear" @click="onDeleteFavorite(f.id)">Удалить</ion-button>
                  </div>
                </article>
              </div>
            </section>

            <section class="nutrition-card">
              <p class="block-title">База продуктов</p>
              <div class="catalog-list">
                <p v-for="(c, idx) in catalogPreview" :key="idx" class="catalog-item">
                  {{ c.title }} — {{ formatMacros(c) }}
                </p>
              </div>
            </section>

            <section v-if="editingId" class="nutrition-card">
              <p class="block-title">Редактирование записи</p>
              <div class="form-grid">
                <ion-input v-model="editForm.title" class="field" label="Название" label-placement="stacked" />
                <div class="triple-row">
                  <ion-input v-model="editForm.protein" class="field" inputmode="decimal" label="Белки, г" label-placement="stacked" />
                  <ion-input v-model="editForm.fat" class="field" inputmode="decimal" label="Жиры, г" label-placement="stacked" />
                  <ion-input v-model="editForm.carbs" class="field" inputmode="decimal" label="Углеводы, г" label-placement="stacked" />
                </div>
                <div class="pager-row">
                  <ion-button class="sportik-footer-btn" @click="onApplyEdit">Сохранить</ion-button>
                  <ion-button fill="clear" @click="onCancelEdit">Отмена</ion-button>
                </div>
              </div>
            </section>
          </div>
        </div>
      </div>

      <div class="nutrition-footer-stack">
        <app-tab-bar active-key="nutrition" />
      </div>
    </ion-content>
  </ion-page>
</template>

<script setup>
defineOptions({ name: 'NutritionPage' })

import { computed, onMounted, reactive, ref } from 'vue'
import { IonPage, IonContent, IonButton, IonInput, toastController } from '@ionic/vue'
import AppTabBar from '@/components/navigation/AppTabBar.vue'
import { useNutritionStore } from '@/stores/nutrition'

const nutrition = useNutritionStore()
const saving = ref(false)
const editingId = ref(null)
const form = reactive({
  title: '',
  protein: '',
  fat: '',
  carbs: ''
})
const editForm = reactive({
  title: '',
  protein: '',
  fat: '',
  carbs: ''
})
const filters = reactive({
  search: '',
  dateFrom: '',
  dateTo: ''
})

onMounted(async () => {
  await nutrition.hydrateAll({ page: 1, page_size: 20 })
})

const entries = computed(() => nutrition.entries)
const favorites = computed(() => nutrition.favorites)
const goal = computed(() => nutrition.goal)
const stats = computed(() => nutrition.stats)
const catalogPreview = computed(() => nutrition.catalog.slice(0, 20))
const page = computed(() => nutrition.entriesPage)
const pageCount = computed(() => Math.max(1, Math.ceil(nutrition.entriesTotal / nutrition.entriesPageSize)))

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
  return d.toLocaleString('ru-RU', {
    day: '2-digit',
    month: '2-digit',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

async function showToast(message, color = 'success') {
  const toast = await toastController.create({ message, duration: 1500, color })
  await toast.present()
}

function asIsoDate(dateOnly) {
  if (!dateOnly) return ''
  const t = new Date(`${dateOnly}T00:00:00.000Z`)
  return Number.isNaN(t.getTime()) ? '' : t.toISOString()
}

function asIsoDateEnd(dateOnly) {
  if (!dateOnly) return ''
  const t = new Date(`${dateOnly}T23:59:59.999Z`)
  return Number.isNaN(t.getTime()) ? '' : t.toISOString()
}

function currentFilters(extra = {}) {
  return {
    page: page.value,
    page_size: nutrition.entriesPageSize,
    search: filters.search.trim() || undefined,
    date_from: asIsoDate(filters.dateFrom) || undefined,
    date_to: asIsoDateEnd(filters.dateTo) || undefined,
    ...extra
  }
}

async function onAddEntry() {
  saving.value = true
  try {
    await nutrition.addEntry({
      title: form.title,
      protein_g: form.protein,
      fat_g: form.fat,
      carbs_g: form.carbs
    })
    form.title = ''
    form.protein = ''
    form.fat = ''
    form.carbs = ''
    await showToast('Приём пищи добавлен')
    await nutrition.hydrateAll(currentFilters())
  } catch (e) {
    await showToast(e?.response?.data || 'Ошибка добавления', 'danger')
  } finally {
    saving.value = false
  }
}

async function onDeleteEntry(id) {
  try {
    await nutrition.deleteEntry(id)
    await nutrition.hydrateAll(currentFilters())
    await showToast('Запись удалена')
  } catch {
    await showToast('Ошибка удаления', 'danger')
  }
}

async function onCloneToFavorite(entry) {
  try {
    await nutrition.addFavorite({
      title: entry.title,
      protein_g: entry.protein_g,
      fat_g: entry.fat_g,
      carbs_g: entry.carbs_g,
      unit_type: 'gram'
    })
    await showToast('Добавлено в избранное')
  } catch {
    await showToast('Ошибка добавления в избранное', 'danger')
  }
}

async function onAddFromFavorite(fav) {
  try {
    await nutrition.addEntry({
      title: fav.title,
      protein_g: fav.protein_g,
      fat_g: fav.fat_g,
      carbs_g: fav.carbs_g
    })
    await showToast('Добавлено в дневник')
    await nutrition.hydrateAll(currentFilters())
  } catch {
    await showToast('Ошибка добавления из избранного', 'danger')
  }
}

async function onDeleteFavorite(id) {
  try {
    await nutrition.deleteFavorite(id)
    await showToast('Избранное удалено')
  } catch {
    await showToast('Ошибка удаления избранного', 'danger')
  }
}

async function onRecalcGoal(target) {
  try {
    await nutrition.recalculateGoal(target)
    await showToast('Цель пересчитана по данным профиля')
  } catch (e) {
    await showToast(e?.response?.data || 'Ошибка пересчёта цели', 'danger')
  }
}

function onStartEdit(e) {
  editingId.value = e.id
  editForm.title = e.title || ''
  editForm.protein = String(e.protein_g ?? '')
  editForm.fat = String(e.fat_g ?? '')
  editForm.carbs = String(e.carbs_g ?? '')
}

async function onApplyEdit() {
  if (!editingId.value) return
  try {
    await nutrition.updateEntry(editingId.value, {
      title: editForm.title,
      protein_g: editForm.protein,
      fat_g: editForm.fat,
      carbs_g: editForm.carbs
    })
    editingId.value = null
    await nutrition.hydrateAll(currentFilters())
    await showToast('Запись обновлена')
  } catch {
    await showToast('Ошибка обновления', 'danger')
  }
}

function onCancelEdit() {
  editingId.value = null
}

async function changePage(nextPage) {
  await nutrition.hydrateAll(currentFilters({ page: nextPage }))
}

async function applyFilters() {
  await nutrition.hydrateAll(currentFilters({ page: 1 }))
}

async function resetFilters() {
  filters.search = ''
  filters.dateFrom = ''
  filters.dateTo = ''
  await nutrition.hydrateAll({ page: 1, page_size: nutrition.entriesPageSize })
}
</script>

<style scoped>
.nutrition-content { --background: var(--sportik-bg); }
.nutrition-frame { min-height: calc(100svh - env(safe-area-inset-bottom, 0px)); display: flex; flex-direction: column; }
.nutrition-strip { flex: 0 0 clamp(108px, 26vw, 152px); background: linear-gradient(145deg, var(--sportik-brand), var(--sportik-brand-2)); }
.nutrition-sheet { flex: 1; margin-top: -14px; border-radius: 28px 28px 0 0; background: var(--sportik-surface); padding-bottom: calc(120px + env(safe-area-inset-bottom, 0px)); color: var(--sportik-text); box-shadow: var(--sportik-shadow-lg); }
.nutrition-title { margin: 0; text-align: center; font-size: clamp(1.45rem, 5vw, 2rem); font-weight: 700; }
.nutrition-subtitle { margin: 0.4rem 0 1rem; text-align: center; font-size: 0.9rem; color: var(--sportik-text-muted); }
.nutrition-card { background: var(--sportik-surface-soft); border: 1px solid var(--sportik-border); border-radius: 14px; box-shadow: var(--sportik-shadow-md); padding: 12px; margin-bottom: 10px; }
.block-title { margin: 0 0 8px; font-weight: 700; }
.form-grid { display: grid; gap: 10px; }
.triple-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; }
.filters-row { display: grid; grid-template-columns: 1.3fr 1fr 1fr; gap: 8px; margin-bottom: 8px; }
.field { --background: var(--sportik-surface); --color: var(--sportik-text); border-radius: 10px; border: 1px solid var(--sportik-border); }
.goal-row { display: grid; grid-template-columns: repeat(3, 1fr); gap: 8px; margin-bottom: 8px; }
.stats-line { margin: 0 0 4px; font-size: 0.88rem; color: var(--sportik-text-soft); }
.empty { margin: 0; font-size: 0.86rem; color: var(--sportik-text-muted); }
.entries-list, .favorites-list { display: grid; gap: 8px; }
.entry-item, .fav-item { display: flex; justify-content: space-between; gap: 8px; background: var(--sportik-surface); border: 1px solid var(--sportik-border); border-radius: 12px; padding: 10px; }
.entry-main { min-width: 0; }
.entry-title { margin: 0; font-weight: 700; }
.entry-meta, .entry-date { margin: 2px 0 0; font-size: 0.82rem; color: var(--sportik-text-muted); }
.entry-actions { display: grid; align-content: start; gap: 4px; }
.pager-row { margin-top: 8px; display: flex; align-items: center; gap: 8px; }
.pager-text { margin: 0; font-size: 0.86rem; color: var(--sportik-text-muted); }
.catalog-list { display: grid; gap: 4px; max-height: 260px; overflow: auto; }
.catalog-item { margin: 0; font-size: 0.82rem; color: var(--sportik-text-soft); }
.nutrition-footer-stack { position: fixed; left: 0; right: 0; bottom: 0; z-index: 10; background: var(--sportik-surface-glass); box-shadow: 0 -8px 22px rgba(0, 0, 0, 0.1); backdrop-filter: blur(12px); padding-bottom: env(safe-area-inset-bottom, 0px); }

@media (max-width: 760px) {
  .filters-row { grid-template-columns: 1fr; }
  .triple-row, .goal-row { grid-template-columns: 1fr; }
}
</style>

