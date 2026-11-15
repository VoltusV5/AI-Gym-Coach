<template>
  <platform-layout
    :title="admin ? 'Аналитика платформы' : 'Аналитика ваших программ'"
    subtitle="Период включает начальную дату и заканчивается перед конечной датой, по UTC. Продажи и возвраты учитываются отдельно."
  >
    <form class="platform-panel" @submit.prevent="refresh">
      <div class="platform-fields">
        <label>Начало периода<input v-model="from" type="date" required /></label
        ><label
          >Конец периода, не включая<input
            v-model="to"
            type="date"
            :required="!toNow"
            :disabled="toNow"
        /></label>
      </div>
      <label class="check-label"
        ><input v-model="toNow" type="checkbox" />По текущее время, включая сегодняшние
        операции</label
      ><button class="primary" :disabled="load.loading.value">Рассчитать</button>
    </form>
    <async-state :loading="load.loading.value" :error="load.error.value" @retry="refresh"
      ><template v-if="data"
        ><div class="platform-grid">
          <article v-for="metric in metrics" :key="metric.key" class="platform-card">
            <h2>
              {{
                metric.money
                  ? money(data[metric.key] ?? 0)
                  : (data[metric.key] ?? 0).toLocaleString('ru-RU')
              }}
            </h2>
            <p>{{ metric.label }}</p>
          </article>
        </div>
        <p v-if="!admin" class="platform-muted">
          Доступный баланс рассчитан за всё время с учётом возвратов и зарезервированных выплат.
          Остальные финансовые показатели относятся к выбранному периоду.
        </p>
        <p v-else class="platform-muted">
          Активность подтверждается завершённым занятием, покупкой или отзывом. Клиентские показы и
          клики не входят в доход или активность. Вывод комиссии платформы пока не подключён.
        </p>
        <section class="platform-panel">
          <h2>{{ admin ? 'Популярные программы' : 'Ваши программы' }}</h2>
          <p v-if="!programs.length">За выбранный период данных пока нет.</p>
          <div class="analytics-table">
            <table v-if="programs.length">
              <thead>
                <tr>
                  <th>Программа</th>
                  <th>Пользователи</th>
                  <th>Продажи</th>
                  <th v-if="!admin">Оборот</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in programs" :key="item.program_id">
                  <td>{{ item.title || 'Программа без публикации' }}</td>
                  <td>{{ item.users }}</td>
                  <td>{{ item.sales }}</td>
                  <td v-if="!admin">{{ money(item.gross_minor) }}</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>
        <section v-if="admin" class="platform-panel">
          <h2>Тренеры</h2>
          <p v-if="!data.top_trainers.length">Нет данных.</p>
          <p v-for="item in data.top_trainers" :key="item.trainer_id">
            {{ item.display_name }} · {{ item.users }} пользователей · {{ item.sales }} продаж
          </p>
        </section></template
      ></async-state
    >
  </platform-layout>
</template>
<script setup>
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/api/api'
import { money } from '@/utils/platform'
import { useRequest } from '@/composables/useAsync'
import { useViewLoad } from '@/composables/useViewLoad'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const route = useRoute(),
  admin = computed(() => route.path.startsWith('/admin/')),
  load = useRequest(),
  data = computed(() => load.data.value)
const toNow = ref(true)
const today = new Date(),
  to = ref(today.toISOString().slice(0, 10)),
  from = ref(new Date(today.getTime() - 30 * 86400000).toISOString().slice(0, 10))
const metrics = computed(() =>
  (admin.value
    ? [
        ['users', 'Зарегистрированные пользователи сейчас'],
        ['active_users', 'Активные пользователи за период'],
        ['workouts', 'Завершённые занятия'],
        ['published_programs', 'Опубликованные программы сейчас'],
        ['sales', 'Подтверждённые продажи'],
        ['gross_minor', 'Оборот до возвратов', true],
        ['refund_minor', 'Возвраты', true],
        ['net_minor', 'Оборот после возвратов', true],
        ['commission_accrued_minor', 'Начисленная комиссия за вычетом возвратов', true],
        ['commission_paid_minor', 'Выведенная комиссия платформы', true],
      ]
    : [
        ['program_users', 'Пользователи программ за период'],
        ['sales_count', 'Подтверждённые продажи'],
        ['gross_minor', 'Оборот до возвратов', true],
        ['refund_minor', 'Возвраты', true],
        ['commission_minor', 'Комиссия после возвратов', true],
        ['author_earned_minor', 'Начислено автору после возвратов', true],
        ['available_minor', 'Доступный баланс сейчас', true],
        ['paid_out_minor', 'Выплачено за период', true],
      ]
  ).map(([key, label, money]) => ({ key, label, money })),
)
const programs = computed(() =>
  admin.value ? data.value?.top_programs || [] : data.value?.programs || [],
)
async function refresh() {
  await load.run(
    async () =>
      (
        await api.get(`/api/v2/${admin.value ? 'admin' : 'trainer'}/analytics`, {
          params: { from: from.value, to: toNow.value ? undefined : to.value },
        })
      ).data,
  )
}
useViewLoad(refresh)
</script>
<style scoped>
.analytics-table {
  overflow-x: auto;
}
table {
  width: 100%;
  text-align: left;
  border-collapse: collapse;
}
th,
td {
  padding: 12px;
  border-bottom: 1px solid var(--sportik-border);
}
</style>
