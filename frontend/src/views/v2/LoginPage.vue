<script setup>
import { useRoute, useRouter } from 'vue-router'
import FeaturePage from '@/components/layout/FeaturePage.vue'
import ProfileAuthPanel from '@/components/settings/ProfileAuthPanel.vue'
import { safeReturnPath } from '@/utils/session'

const route = useRoute()
const router = useRouter()

/** Resume the requested screen after authentication without replaying its action. */
function resume() {
  return router.replace(safeReturnPath(route.query.return_to))
}
</script>

<template>
  <FeaturePage eyebrow="Аккаунт" title="Вход в Спортик" subtitle="Войдите, чтобы открыть свои данные или получить программу.">
    <ProfileAuthPanel @signed-in="resume" />
    <RouterLink class="feature-back" to="/marketplace">Вернуться в каталог</RouterLink>
  </FeaturePage>
</template>
