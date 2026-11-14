<template>
  <platform-layout
    title="Программы под ваши условия"
    subtitle="Подборка учитывает профиль, оборудование и вашу активность. Объяснения основаны на свойствах программ."
    ><template #actions><router-link to="/profile">Изменить профиль</router-link></template
    ><async-state
      :loading="load.loading.value"
      :error="load.error.value"
      :empty="!result?.items.length"
      @retry="refresh"
      ><template #empty
        >Пока нет совместимых программ. Проверьте профиль или вернитесь к каталогу позже.</template
      ><template v-if="result"
        ><p v-if="result.fallback" class="platform-muted">
          Подборка составлена по совместимости с вашим профилем.
        </p>
        <recommendation-list
          :items="result.items"
          :request-id="result.request_id" /></template></async-state
  ></platform-layout>
</template>
<script setup>
import { computed } from 'vue'
import api from '@/api/api'
import { useRequest } from '@/composables/useAsync'
import { useViewLoad } from '@/composables/useViewLoad'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
import RecommendationList from '@/components/platform/RecommendationList.vue'
const load = useRequest(),
  result = computed(() => load.data.value)
async function refresh() {
  await load.run(
    async () => (await api.get('/api/v2/recommendations/programs', { params: { limit: 20 } })).data,
  )
}
useViewLoad(refresh)
</script>
