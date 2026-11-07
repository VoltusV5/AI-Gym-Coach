<script setup>
import { useRoute, RouterLink } from 'vue-router'

defineProps({
  loading: Boolean,
  error: { type: String, default: '' },
  status: { type: Number, default: 0 },
  empty: { type: String, default: '' },
  retry: { type: Function, default: null },
})
const route = useRoute()
</script>

<template>
  <div v-if="loading" class="feature-state" role="status">
    <span class="feature-spinner" aria-hidden="true" /> Загрузка…
  </div>
  <div v-else-if="error" class="feature-state feature-error" role="alert">
    <p>{{ error }}</p>
    <RouterLink v-if="status === 401" class="feature-button" :to="{ path: '/auth/login', query: { return_to: route.fullPath } }">
      Войти
    </RouterLink>
    <button v-else-if="retry && ![403, 404].includes(status)" class="feature-button secondary" type="button" @click="retry()">
      Повторить
    </button>
  </div>
  <div v-else-if="empty" class="feature-state">{{ empty }}</div>
</template>
