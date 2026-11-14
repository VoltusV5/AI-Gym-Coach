<template>
  <platform-layout
    :title="admin ? 'Заказы платформы' : 'Мои покупки'"
    subtitle="Состояние каждого заказа подтверждается сервером."
  >
    <async-state
      :loading="page.loading.value"
      :error="page.error.value"
      :empty="!page.items.value.length"
      @retry="page.load()"
      ><template #empty
        >У вас пока нет покупок.
        <router-link to="/catalog">Посмотреть программы</router-link>.</template
      >
      <div class="platform-grid">
        <article v-for="order in page.items.value" :key="order.id" class="platform-card">
          <h2>{{ money(order.amount_minor, order.currency) }}</h2>
          <p>{{ statuses[order.status] || order.status }}</p>
          <router-link v-if="!admin" :to="`/checkout/${order.id}`">Проверить заказ</router-link>
          <p>
            <router-link v-if="order.status === 'paid'" :to="`/programs/${order.program_id}`"
              >Открыть программу</router-link
            >
          </p>
        </article>
      </div></async-state
    >
    <button v-if="page.nextCursor.value" :disabled="page.loading.value" @click="page.load(true)">
      Ещё покупки
    </button>
  </platform-layout>
</template>
<script setup>
import { useViewLoad } from '@/composables/useViewLoad'
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { usePage } from '@/composables/useAsync'
import { money } from '@/utils/platform'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const route = useRoute(),
  admin = computed(() => route.path.startsWith('/admin'))
const page = usePage(() => (admin.value ? '/api/v2/admin/orders' : '/api/v2/me/orders'))
const statuses = {
  pending: 'Ожидает оплаты',
  paid: 'Оплачено',
  failed: 'Не оплачено',
  refunded: 'Возврат',
  cancelled: 'Отменено',
  expired: 'Срок истёк',
}
useViewLoad(() => page.load())
</script>
