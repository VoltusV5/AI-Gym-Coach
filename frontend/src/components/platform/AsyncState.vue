<template>
  <div v-if="loading" class="async-state" role="status" aria-live="polite">Загружаем…</div>
  <div v-else-if="error" class="async-state async-state--error" role="alert">
    <p>{{ error }}</p>
    <button type="button" @click="$emit('retry')">Повторить</button>
  </div>
  <div v-else-if="empty" class="async-state"><slot name="empty">Здесь пока нет данных.</slot></div>
  <slot v-else />
</template>

<script setup>
defineProps({ loading: Boolean, error: { type: String, default: '' }, empty: Boolean })
defineEmits(['retry'])
</script>

<style scoped>
.async-state {
  padding: 28px 16px;
  text-align: center;
  color: var(--sportik-text-muted);
}
.async-state--error {
  color: var(--sportik-danger);
}
button {
  border: 1px solid currentColor;
  border-radius: 12px;
  padding: 10px 18px;
  color: var(--sportik-brand);
  background: var(--sportik-surface);
  cursor: pointer;
}
</style>
