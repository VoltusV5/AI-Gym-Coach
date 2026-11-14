<template>
  <platform-layout
    :title="admin ? 'Финансы платформы' : 'Начисления и выплаты'"
    subtitle="Тестовый финансовый учёт. Реальные деньги не переводятся."
  >
    <template v-if="admin" #actions><router-link to="/admin">Управление</router-link></template>
    <section v-if="admin" class="platform-panel">
      <h2>Комиссия платформы</h2>
      <p v-if="commission.error.value" class="platform-error" role="alert">
        {{ commission.error.value }}
      </p>
      <form v-if="commission.data.value" @submit.prevent="saveCommission">
        <div class="platform-fields">
          <label
            >Комиссия, %<input
              v-model="percentage"
              type="number"
              min="0"
              max="100"
              step="0.01"
              required /></label
          ><label
            >Начало действия<input v-model="effectiveAt" type="datetime-local" required
          /></label>
        </div>
        <button :disabled="action.busy.value">Сохранить новую ставку</button>
        <p class="platform-muted">
          Ставка сохраняется в каждом новом заказе. Прошлые продажи не пересчитываются.
        </p>
      </form>
    </section>
    <section v-if="admin" class="platform-panel">
      <h2>Создать тестовую выплату</h2>
      <form @submit.prevent="createPayout">
        <div class="platform-fields">
          <label>ID тренера<input v-model.number="trainer" type="number" min="1" required /></label
          ><label
            >Сумма, руб.<input v-model="amount" type="number" min="0.01" step="0.01" required
          /></label>
        </div>
        <button class="primary" :disabled="action.busy.value">Записать тестовую выплату</button>
      </form>
    </section>
    <p v-if="action.error.value" role="alert" class="platform-error">{{ action.error.value }}</p>
    <p v-if="action.success.value" role="status" class="platform-success">
      {{ action.success.value }}
    </p>
    <h2>Выплаты</h2>
    <async-state
      :loading="page.loading.value"
      :error="page.error.value"
      :empty="!page.items.value.length"
      @retry="page.load()"
      ><div class="platform-grid">
        <article v-for="payout in page.items.value" :key="payout.id" class="platform-card">
          <h3>{{ money(payout.amount_minor, payout.currency) }}</h3>
          <p v-if="admin">Тренер {{ payout.trainer_id }}</p>
          <p>{{ statuses[payout.status] || payout.status }}</p>
          <div v-if="admin" class="platform-actions">
            <button :disabled="action.busy.value" @click="decide(payout, 'reconcile')">
              Проверить состояние</button
            ><button
              v-if="payout.status === 'failed'"
              :disabled="action.busy.value"
              @click="decide(payout, 'retry')"
            >
              Повторить</button
            ><button
              v-if="payout.status === 'requested' || payout.status === 'failed'"
              :disabled="action.busy.value"
              @click="decide(payout, 'cancel')"
            >
              Отменить
            </button>
          </div>
        </article>
      </div></async-state
    ><button v-if="page.nextCursor.value" :disabled="page.loading.value" @click="page.load(true)">
      Ещё выплаты
    </button>
  </platform-layout>
</template>
<script setup>
import { useViewLoad } from '@/composables/useViewLoad'
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/api/api'
import { usePage, useRequest, useAction } from '@/composables/useAsync'
import { money } from '@/utils/platform'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const route = useRoute(),
  admin = computed(() => route.path.startsWith('/admin')),
  action = useAction(),
  commission = useRequest()
const page = usePage(() => (admin.value ? '/api/v2/admin/payouts' : '/api/v2/trainer/payouts'))
const percentage = ref('20'),
  effectiveAt = ref(''),
  trainer = ref(null),
  amount = ref('')
const statuses = {
  requested: 'Запрошено',
  processing: 'Обрабатывается',
  paid: 'Тестовая выплата записана',
  failed: 'Ошибка',
  cancelled: 'Отменено',
}
function hundredths(value) {
  const match = String(value).match(/^(\d+)(?:[.,](\d{1,2}))?$/)
  return match ? Number(match[1]) * 100 + Number((match[2] || '').padEnd(2, '0')) : null
}
async function fetchCommission() {
  const data = await commission.run(
    async () => (await api.get('/api/v2/admin/settings/commission')).data,
  )
  if (data) {
    percentage.value = (data.commission_bps / 100).toFixed(2)
    effectiveAt.value = new Date(Date.now() - new Date().getTimezoneOffset() * 60000)
      .toISOString()
      .slice(0, 16)
  }
}
async function saveCommission() {
  const bps = hundredths(percentage.value)
  if (bps == null || bps > 10000) {
    action.error.value = 'Укажите ставку от 0 до 100% с точностью до 0,01%.'
    return
  }
  if (!confirm('Применить новую ставку к будущим заказам?')) return
  if (
    await action.run(
      '/api/v2/admin/settings/commission',
      {
        commission_bps: bps,
        effective_at: new Date(effectiveAt.value).toISOString(),
        version: commission.data.value.version,
      },
      { method: 'patch' },
    )
  )
    await fetchCommission()
}
async function createPayout() {
  const minor = hundredths(amount.value)
  if (minor == null || minor <= 0) return
  if (!confirm('Создать запись выплаты в тестовом учёте? Деньги не переводятся.')) return
  if (
    await action.run(
      '/api/v2/admin/payouts',
      { trainer_id: trainer.value, amount_minor: minor, currency: 'RUB' },
      { idempotent: true, message: 'Тестовая выплата записана' },
    )
  )
    await page.load()
}
async function decide(payout, choice) {
  const reason = prompt('Укажите причину действия')
  if (!reason?.trim()) return
  if (
    await action.run(
      `/api/v2/admin/payouts/${payout.id}`,
      { action: choice, reason },
      { method: 'patch' },
    )
  )
    await page.load()
}
useViewLoad(() => {
  page.load()
  if (admin.value) fetchCommission()
})
</script>
