<template>
  <nutrition-chrome title="Мои блюда" subtitle="на 100 г" :show-back="true">
    <!-- Сначала форма создания/редактирования -->
    <section class="nutrition-card">
      <div class="list-head">
        <p class="block-title">{{ formTitle }}</p>
        <ion-button size="small" fill="outline" class="add-inline" @click="startCreate">
          Добавить
        </ion-button>
      </div>
      <div class="form-grid">
        <ion-input v-model="form.title" class="field" label="Название" label-placement="stacked" />
        <p class="kcal-hint">Ккал по БЖУ (на 100 г): <strong>{{ kcalPreview }}</strong></p>
        <div class="triple-row">
          <ion-input v-model="form.protein" class="field" inputmode="decimal" label="Белки, г" label-placement="stacked" />
          <ion-input v-model="form.fat" class="field" inputmode="decimal" label="Жиры, г" label-placement="stacked" />
          <ion-input v-model="form.carbs" class="field" inputmode="decimal" label="Углеводы, г" label-placement="stacked" />
        </div>
        <ion-input
          v-model="form.base_grams"
          class="field"
          inputmode="decimal"
          label="Порция для расчёта, г (часто 100)"
          label-placement="stacked"
        />
        <div class="actions-row">
          <ion-button class="sportik-footer-btn" expand="block" :disabled="saving" @click="onSave">
            Сохранить блюдо
          </ion-button>
          <ion-button expand="block" fill="outline" color="medium" :disabled="saving" @click="onCancel">
            Отмена
          </ion-button>
        </div>
      </div>
    </section>

    <!-- Ниже список сохранённых блюд -->
    <section class="nutrition-card">
      <p class="block-title">Список</p>
      <p v-if="!store.myDishes.length" class="empty">Пока пусто — заполните форму выше и нажмите «Сохранить блюдо».</p>
      <div v-else class="dish-list">
        <div v-for="d in store.myDishes" :key="d.id" class="dish-row-wrap">
          <button
            type="button"
            class="dish-row"
            :class="{ active: editingId === d.id }"
            @click="loadIntoForm(d)"
          >
            <span class="dish-title">{{ d.title }}</span>
            <span class="dish-meta">{{ lineMeta(d) }}</span>
          </button>
          <ion-button
            class="row-del"
            fill="clear"
            size="small"
            color="danger"
            aria-label="Удалить"
            @click.stop="onDeleteAsk(d)"
          >
            ✕
          </ion-button>
        </div>
      </div>
    </section>
  </nutrition-chrome>
</template>

<script setup>
defineOptions({ name: 'NutritionMyDishes' })

import { computed, onMounted, reactive, ref } from 'vue'
import { IonButton, IonInput, toastController, alertController } from '@ionic/vue'
import NutritionChrome from '@/components/nutrition/NutritionChrome.vue'
import { useNutritionStore } from '@/stores/nutrition'

const store = useNutritionStore()
const editingId = ref(null)
const saving = ref(false)
const form = reactive({
  title: '',
  protein: '',
  fat: '',
  carbs: '',
  base_grams: '100'
})

const formTitle = computed(() => {
  if (!editingId.value) return 'Новое блюдо'
  const t = String(form.title || '').trim()
  return t ? `Редактирование: ${t}` : 'Редактирование'
})

function toNum(v) {
  const n = Number(String(v ?? '').replace(',', '.'))
  return Number.isFinite(n) ? n : 0
}

const kcalPreview = computed(() => {
  const p = toNum(form.protein)
  const f = toNum(form.fat)
  const c = toNum(form.carbs)
  return (p * 4 + f * 9 + c * 4).toFixed(0)
})

function lineMeta(d) {
  const p = toNum(d.protein_g)
  const f = toNum(d.fat_g)
  const c = toNum(d.carbs_g)
  const k = toNum(d.calories) || p * 4 + f * 9 + c * 4
  return `Б ${p.toFixed(1)} / Ж ${f.toFixed(1)} / У ${c.toFixed(1)} · ${k.toFixed(0)} ккал / 100 г`
}

function extractApiError(e) {
  const st = e?.response?.status
  const raw = e?.response?.data
  if (typeof raw === 'string' && raw.trim()) return raw.trim().slice(0, 240)
  if (st === 409) return 'Блюдо с таким названием уже есть в справочнике'
  if (st === 400) return 'Проверьте название и числа БЖУ'
  if (st === 401) return 'Сессия истекла — войдите снова'
  if (st === 404) return 'Блюдо не найдено или чужое'
  if (st === 500) return 'Ошибка сервера при сохранении'
  return e?.message ? String(e.message).slice(0, 200) : 'Не удалось выполнить запрос'
}

async function toast(message, color = 'success') {
  const t = await toastController.create({ message, duration: 2200, color })
  await t.present()
}

async function refresh() {
  await store.fetchMyDishes()
}

function loadIntoForm(d) {
  editingId.value = d.id
  form.title = String(d.title || '')
  form.protein = d.protein_g != null ? String(d.protein_g) : ''
  form.fat = d.fat_g != null ? String(d.fat_g) : ''
  form.carbs = d.carbs_g != null ? String(d.carbs_g) : ''
  form.base_grams = d.base_grams != null && Number(d.base_grams) > 0 ? String(d.base_grams) : '100'
}

function resetForm() {
  editingId.value = null
  form.title = ''
  form.protein = ''
  form.fat = ''
  form.carbs = ''
  form.base_grams = '100'
}

function startCreate() {
  resetForm()
}

function onCancel() {
  resetForm()
}

onMounted(async () => {
  try {
    await refresh()
  } catch (e) {
    await toast(extractApiError(e), 'danger')
  }
})

function buildPayload() {
  return {
    title: String(form.title || '').trim(),
    protein_g: toNum(form.protein),
    fat_g: toNum(form.fat),
    carbs_g: toNum(form.carbs),
    base_grams: Math.max(1, toNum(form.base_grams) || 100)
  }
}

async function onSave() {
  const title = String(form.title || '').trim()
  if (!title) {
    await toast('Укажите название блюда', 'warning')
    return
  }
  saving.value = true
  try {
    const payload = buildPayload()
    if (editingId.value) {
      await store.updateDish(editingId.value, payload)
      await toast('Блюдо сохранено')
    } else {
      const created = await store.createDish(payload)
      if (created?.id) {
        editingId.value = created.id
        loadIntoForm(created)
      }
      await toast('Блюдо добавлено в список')
    }
    await refresh()
    if (editingId.value) {
      const row = store.myDishes.find((x) => Number(x.id) === Number(editingId.value))
      if (row) loadIntoForm(row)
    }
  } catch (e) {
    await toast(extractApiError(e), 'danger')
  } finally {
    saving.value = false
  }
}

async function onDeleteAsk(d) {
  const id = d?.id
  if (id == null) return
  const name = String(d.title || 'блюдо')
  const alert = await alertController.create({
    header: 'Удалить блюдо?',
    message: `«${name}» будет удалено из справочника. Записи в дневнике останутся, связь с блюдом снимется.`,
    buttons: [
      { text: 'Отмена', role: 'cancel' },
      {
        text: 'Удалить',
        role: 'destructive',
        handler: () => doDelete(id)
      }
    ]
  })
  await alert.present()
}

async function doDelete(id) {
  saving.value = true
  try {
    await store.deleteDish(id)
    if (Number(editingId.value) === Number(id)) resetForm()
    await refresh()
    await toast('Удалено')
  } catch (e) {
    await toast(extractApiError(e), 'danger')
  } finally {
    saving.value = false
  }
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
.list-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 8px;
}
.add-inline {
  margin: 0;
  flex-shrink: 0;
}
.block-title {
  margin: 0;
  font-weight: 700;
  color: var(--sportik-text);
}
.empty {
  margin: 0;
  color: var(--sportik-text-muted);
  font-size: 0.9rem;
}
.dish-list {
  display: grid;
  gap: 8px;
  max-height: 42vh;
  overflow: auto;
}
.dish-row-wrap {
  display: flex;
  align-items: stretch;
  gap: 4px;
}
.dish-row {
  flex: 1;
  min-width: 0;
  text-align: left;
  border-radius: 12px;
  border: 1px solid var(--sportik-border);
  background: var(--sportik-surface);
  padding: 10px 12px;
  color: var(--sportik-text);
  cursor: pointer;
  display: grid;
  gap: 4px;
}
.dish-row.active {
  border-color: var(--sportik-brand);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--sportik-brand) 45%, transparent);
}
.row-del {
  margin: 0;
  align-self: center;
  flex-shrink: 0;
  --padding-start: 8px;
  --padding-end: 8px;
}
.dish-title {
  font-weight: 700;
  font-size: 0.95rem;
}
.dish-meta {
  font-size: 0.8rem;
  color: var(--sportik-text-muted);
}
.form-grid {
  display: grid;
  gap: 10px;
}
.kcal-hint {
  margin: 0;
  font-size: 0.88rem;
  color: var(--sportik-text-muted);
}
.triple-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}
.field {
  --background: var(--sportik-surface);
  --color: var(--sportik-text);
  border-radius: 10px;
  border: 1px solid var(--sportik-border);
}
.actions-row {
  display: grid;
  gap: 8px;
}
</style>
