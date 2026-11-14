<template>
  <platform-layout
    title="Ваш прогресс"
    subtitle="Фактически завершённые занятия, рабочий объём и соблюдение расписания."
  >
    <form class="platform-panel" @submit.prevent="refresh">
      <div class="platform-fields">
        <label
          >Период<select v-model="period">
            <option value="week">Неделя</option>
            <option value="month">Месяц</option>
            <option value="all">Всё время</option>
          </select></label
        ><label>Дата периода<input v-model="anchor" type="date" /></label
        ><label>Часовой пояс<input v-model="zone" required /></label>
      </div>
      <button class="primary">Показать прогресс</button>
    </form>
    <async-state :loading="load.loading.value" :error="load.error.value" @retry="refresh"
      ><template v-if="progress"
        ><div class="platform-grid">
          <article class="platform-card">
            <h2>{{ progress.workout_count }}</h2>
            <p>Завершённых тренировок</p>
          </article>
          <article class="platform-card">
            <h2>{{ progress.volume_kg.toLocaleString('ru-RU') }} кг</h2>
            <p>Рабочий объём</p>
          </article>
          <article class="platform-card">
            <h2>
              {{
                progress.adherence == null
                  ? 'Нет расписания'
                  : `${Math.round(progress.adherence * 100)}%`
              }}
            </h2>
            <p>Соблюдение плана</p>
          </article>
          <article class="platform-card">
            <h2>{{ progress.streak_weeks }}</h2>
            <p>Недель регулярности</p>
          </article>
          <article v-if="progress.weight_change_kg != null" class="platform-card">
            <h2>
              {{ progress.weight_change_kg > 0 ? '+' : '' }}{{ progress.weight_change_kg }} кг
            </h2>
            <p>Изменение веса за период</p>
          </article>
        </div>
        <p v-if="!progress.workout_count" class="platform-muted">
          За этот период занятий пока нет. После первой тренировки здесь появятся результаты.
        </p>
        <section v-if="progress.highlights.length" class="platform-panel">
          <h2>Личные результаты</h2>
          <p v-for="(item, index) in progress.highlights" :key="index">{{ item.message }}</p>
        </section></template
      ></async-state
    >
    <form class="platform-panel" @submit.prevent="saveWeight">
      <h2>Вес тела</h2>
      <div class="platform-fields">
        <label>Дата<input v-model="weight.day" type="date" required /></label
        ><label
          >Вес, кг<input
            v-model.number="weight.weight_kg"
            type="number"
            min="1"
            max="500"
            step="0.1"
            required
        /></label>
      </div>
      <button :disabled="action.busy.value">Записать вес</button>
      <p v-if="action.error.value" role="alert">{{ action.error.value }}</p>
      <p v-if="action.success.value" role="status">{{ action.success.value }}</p>
    </form>
    <section class="platform-panel">
      <h2>История тренировок</h2>
      <async-state
        :loading="history.loading.value"
        :error="history.error.value"
        :empty="!history.items.value.length"
        @retry="history.load()"
        ><template #empty>История пока пуста.</template>
        <article v-for="item in history.items.value" :key="item.id">
          <h3>
            <router-link :to="`/workouts/${item.id}`">{{
              formatDate(item.finished_at)
            }}</router-link>
          </h3>
          <p>{{ item.slots.length }} упражнений</p>
        </article></async-state
      ><button
        v-if="history.nextCursor.value"
        :disabled="history.loading.value"
        @click="history.load(true)"
      >
        Ещё занятия
      </button>
    </section>
    <section class="platform-panel">
      <h2>Итоги периодов</h2>
      <async-state
        :loading="summaries.loading.value"
        :error="summaries.error.value"
        :empty="!summaries.items.value.length"
        @retry="summaries.load()"
        ><template #empty>Итогов пока нет.</template>
        <article v-for="item in summaries.items.value" :key="item.id">
          <h3>{{ item.period_start }}</h3>
          <p v-for="message in item.messages" :key="message">{{ message }}</p>
          <p>{{ item.metrics.workout_count }} занятий · {{ item.metrics.volume_kg }} кг</p>
        </article></async-state
      ><button
        v-if="summaries.nextCursor.value"
        :disabled="summaries.loading.value"
        @click="summaries.load(true)"
      >
        Ещё итоги
      </button>
    </section>
  </platform-layout>
</template>
<script setup>
import { ref, computed } from 'vue'
import api from '@/api/api'
import { useRequest, usePage, useAction } from '@/composables/useAsync'
import { useViewLoad } from '@/composables/useViewLoad'
import { localDate } from '@/utils/platform'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const load = useRequest(),
  action = useAction(),
  period = ref('month'),
  anchor = ref(localDate()),
  zone = ref(Intl.DateTimeFormat().resolvedOptions().timeZone || 'Europe/Moscow')
const progress = computed(() => load.data.value),
  weight = ref({ day: localDate(), weight_kg: '' })
const history = usePage('/api/v2/workouts'),
  summaries = usePage('/api/v2/progress/summaries', () => ({
    period: period.value === 'all' ? 'month' : period.value,
    time_zone: zone.value,
  }))
function formatDate(value) {
  try {
    return new Date(value).toLocaleString('ru-RU', { timeZone: zone.value })
  } catch {
    return new Date(value).toLocaleString('ru-RU')
  }
}
async function refresh() {
  await load.run(
    async () =>
      (
        await api.get('/api/v2/progress', {
          params: { period: period.value, anchor_date: anchor.value, time_zone: zone.value },
        })
      ).data,
  )
  await Promise.all([history.load(), summaries.load()])
}
async function saveWeight() {
  if (await action.run('/api/v2/progress/weight', weight.value, { message: 'Вес записан' }))
    await refresh()
}
useViewLoad(refresh)
</script>
