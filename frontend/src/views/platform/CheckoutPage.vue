<template>
  <platform-layout
    title="Тестовая оплата"
    subtitle="Деньги не списываются. Этот режим проверяет покупку программы внутри приложения."
  >
    <async-state :loading="load.loading.value" :error="load.error.value" @retry="fetchOrder"
      ><section v-if="order" class="platform-panel">
        <h2>{{ money(order.amount_minor, order.currency) }}</h2>
        <p>Состояние: {{ statuses[order.status] || order.status }}</p>
        <p v-if="order.status === 'pending'">
          Выберите исход тестового платежа. Возврат на страницу сам по себе не подтверждает оплату.
        </p>
        <div class="platform-actions">
          <button
            v-if="order.status === 'pending'"
            class="primary"
            :disabled="action.busy.value"
            @click="decide('pay')"
          >
            Подтвердить тестовую оплату</button
          ><button
            v-if="order.status === 'pending'"
            :disabled="action.busy.value"
            @click="decide('fail')"
          >
            Отклонить тестовый платёж</button
          ><router-link
            v-if="order.status === 'paid'"
            class="platform-button primary"
            :to="`/programs/${order.program_id}`"
            >Открыть купленную программу</router-link
          ><button v-if="order.status === 'paid'" :disabled="action.busy.value" @click="refund">
            Тестовый возврат</button
          ><button :disabled="action.busy.value" @click="fetchOrder">
            Проверить серверный статус
          </button>
        </div>
        <p v-if="action.error.value" role="alert" class="platform-error">
          {{ action.error.value }}
        </p>
        <p v-if="order.status === 'failed' || order.status === 'refunded'" class="platform-muted">
          Доступ этим заказом не предоставлен.
        </p>
      </section></async-state
    >
  </platform-layout>
</template>
<script setup>
import { useViewLoad } from '@/composables/useViewLoad'
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/api/api'
import { useRequest, useAction } from '@/composables/useAsync'
import { money } from '@/utils/platform'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const route = useRoute(),
  load = useRequest(),
  action = useAction(),
  order = computed(() => load.data.value)
const statuses = {
  pending: 'Ожидает подтверждения',
  paid: 'Подтверждён',
  failed: 'Отклонён',
  refunded: 'Возвращён',
  cancelled: 'Отменён',
  expired: 'Срок истёк',
}
async function fetchOrder() {
  await load.run(async () => (await api.get(`/api/v2/orders/${route.params.id}`)).data)
}
async function decide(actionName) {
  if (
    await action.run(
      `/api/v2/payments/sandbox/${route.params.id}`,
      { action: actionName },
      { message: 'Статус обновлён' },
    )
  )
    await fetchOrder()
}
async function refund() {
  if (confirm('Тестовый возврат отзовёт доступ к программе. Продолжить?')) await decide('refund')
}
useViewLoad(fetchOrder)
</script>
