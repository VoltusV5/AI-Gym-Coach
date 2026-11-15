<template>
  <platform-layout
    title="Управление платформой"
    subtitle="Работа с пользователями, публикациями и финансовым учётом."
    ><div class="platform-grid">
      <article v-for="item in items" :key="item.to" class="platform-card">
        <h2>
          <router-link :to="item.to">{{ item.title }}</router-link>
        </h2>
        <p class="platform-muted">{{ item.description }}</p>
      </article>
    </div></platform-layout
  >
</template>
<script setup>
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
const items = [
  { to: '/admin/users', title: 'Пользователи', description: 'Роли, блокировки и поиск.' },
  {
    to: '/admin/verifications',
    title: 'Тренеры',
    description: 'Заявки на подтверждение квалификации.',
  },
  { to: '/admin/programs', title: 'Программы', description: 'Содержание и решения о публикации.' },
  { to: '/admin/orders', title: 'Заказы', description: 'Подтверждённые покупки и возвраты.' },
  {
    to: '/admin/finance',
    title: 'Выплаты и комиссия',
    description: 'Учёт доходов и тестовых выплат.',
  },
  { to: '/admin/exercises', title: 'Упражнения', description: 'Редактирование справочника.' },
  { to: '/admin/categories', title: 'Категории', description: 'Категории программ.' },
  { to: '/admin/reviews', title: 'Отзывы', description: 'Видимость отзывов и оценки.' },
  { to: '/admin/reports', title: 'Жалобы', description: 'Очередь обращений и решения.' },
  {
    to: '/admin/analytics',
    title: 'Аналитика',
    description: 'Активность, продажи и популярность.',
  },
]
</script>
