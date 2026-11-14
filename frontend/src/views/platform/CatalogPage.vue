<template>
  <platform-layout
    title="Найдите свой тренировочный план"
    subtitle="Программы тренеров под вашу цель, подготовку и доступное оборудование."
  >
    <form class="platform-panel" @submit.prevent="page.load()">
      <div class="platform-fields">
        <label
          >Цель<select v-model="filter.goal">
            <option value="">Все цели</option>
            <option v-for="(label, key) in goalNames" :key="key" :value="key">{{ label }}</option>
          </select></label
        >
        <label
          >Уровень<select v-model="filter.level">
            <option value="">Все уровни</option>
            <option v-for="(label, key) in levelNames" :key="key" :value="key">{{ label }}</option>
          </select></label
        >
        <label
          >Длительность, недель<input
            v-model="filter.duration_weeks"
            type="number"
            min="1"
            max="104"
            placeholder="Любая"
        /></label>
        <label
          >Занятий в неделю<input
            v-model="filter.sessions_per_week"
            type="number"
            min="1"
            max="7"
            placeholder="Любая частота"
        /></label>
        <label
          >Стоимость<select v-model="filter.price_type">
            <option value="">Любая</option>
            <option value="free">Бесплатные</option>
            <option value="paid">Платные</option>
          </select></label
        >
        <label
          >Сортировка<select v-model="filter.sort">
            <option value="newest">Новые</option>
            <option value="price_asc">Сначала дешевле</option>
            <option value="price_desc">Сначала дороже</option>
          </select></label
        >
      </div>
      <fieldset>
        <legend>Доступное оборудование</legend>
        <p v-if="equipment.error.value" role="alert">{{ equipment.error.value }}</p>
        <label v-for="item in equipment.data.value || []" :key="item.id" class="check-label"
          ><input v-model="selectedEquipment" type="checkbox" :value="item.id" />{{
            item.name
          }}</label
        ><label class="check-label"
          ><input v-model="equipmentFilter" type="checkbox" />Применить ограничение
          оборудования</label
        >
      </fieldset>
      <div class="platform-actions">
        <button class="primary">Показать программы</button
        ><button type="button" @click="reset">Сбросить фильтры</button>
      </div>
    </form>
    <async-state
      :loading="page.loading.value && !page.items.value.length"
      :error="page.error.value"
      :empty="!page.items.value.length"
      @retry="page.load()"
      ><template #empty>Подходящих программ пока нет. Попробуйте изменить фильтры.</template>
      <div class="platform-grid">
        <program-card
          v-for="program in page.items.value"
          :key="program.id"
          :program="program"
        /></div
    ></async-state>
    <div v-if="page.nextCursor.value" class="platform-actions">
      <button :disabled="page.loading.value" @click="page.load(true)">Ещё программы</button>
    </div>
  </platform-layout>
</template>
<script setup>
import { referenceItems } from '@/api/referenceLists'
import { useViewLoad } from '@/composables/useViewLoad'
import { reactive, ref } from 'vue'
import { useRoute } from 'vue-router'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
import ProgramCard from '@/components/platform/ProgramCard.vue'
import { usePage, useRequest } from '@/composables/useAsync'
import { goalNames, levelNames } from '@/utils/platform'
const route = useRoute()
const defaults = {
  goal: '',
  level: '',
  duration_weeks: '',
  sessions_per_week: '',
  price_type: '',
  sort: 'newest',
}
const filter = reactive({ ...defaults }),
  selectedEquipment = ref([]),
  equipmentFilter = ref(false)
const equipment = useRequest()
const page = usePage('/api/v2/catalog/programs', () => ({
  ...(route.query.trainer_id ? { trainer_id: route.query.trainer_id } : {}),
  ...Object.fromEntries(Object.entries(filter).filter(([, value]) => value !== '')),
  ...(equipmentFilter.value
    ? { equipment: selectedEquipment.value.length ? selectedEquipment.value : '' }
    : {}),
}))
async function reset() {
  Object.assign(filter, defaults)
  selectedEquipment.value = []
  equipmentFilter.value = false
  await page.load()
}
useViewLoad(() => {
  page.load()
  equipment.run(async () => {
    return referenceItems('/api/v2/equipment')
  })
})
</script>
