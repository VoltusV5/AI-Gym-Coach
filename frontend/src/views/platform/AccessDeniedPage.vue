<template>
  <platform-layout
    title="Этот раздел недоступен"
    subtitle="Проверьте учётную запись и доступную вам роль."
    ><p>
      Права определяет сервер. Если вам нужен доступ тренера или администратора, обратитесь к
      администратору проекта.
    </p>
    <router-link to="/catalog">Перейти в каталог</router-link></platform-layout
  >
</template>
<script setup>
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
</script>
