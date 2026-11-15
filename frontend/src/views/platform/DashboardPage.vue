<template>
  <platform-layout
    title="Сегодня — шаг вперёд"
    subtitle="Ваш план, последние результаты и следующие шаги в одном месте."
  >
    <template #actions><router-link to="/plans">Все планы</router-link></template
    ><async-state :loading="load.loading.value" :error="load.error.value" @retry="refresh"
      ><template v-if="dashboard"
        ><section v-for="block in dashboard.blocks" :key="block.id" class="platform-panel">
          <h2>{{ titles[block.type] || 'Для вас' }}</h2>
          <p v-if="block.status === 'unavailable'" class="platform-muted">
            Этот блок временно недоступен. Остальные данные доступны.
          </p>
          <template v-else-if="block.type === 'recommendation'"
            ><p v-if="!block.data.items?.length" class="platform-muted">
              Пока нет подходящих программ.
            </p>
            <recommendation-list
              v-else
              :items="block.data.items"
              :request-id="block.data.request_id" /></template
          ><template v-else-if="block.type === 'followed_trainers'"
            ><p v-if="!block.data.items.length" class="platform-muted">
              Подпишитесь на тренеров, чтобы видеть их новые программы.
            </p>
            <p v-for="item in block.data.items" :key="item.program_id">
              <router-link :to="`/programs/${item.program_id}`">{{ item.title }}</router-link> ·
              {{ item.display_name }}
            </p></template
          ><template v-else
            ><p v-if="block.data.title">{{ block.data.title }}</p>
            <p v-if="block.data.message">{{ block.data.message }}</p>
            <p v-if="block.data.workout_count != null">Занятий: {{ block.data.workout_count }}</p>
            <p v-if="block.data.volume_kg != null">Рабочий объём: {{ block.data.volume_kg }} кг</p>
            <p v-if="block.data.metrics">
              Рабочий объём: {{ block.data.metrics.volume_kg }} кг · подходов:
              {{ block.data.metrics.working_sets }}
            </p></template
          >
          <div
            v-if="block.action?.route?.startsWith('/') && !block.action.route.startsWith('//')"
            class="platform-actions"
          >
            <router-link :to="block.action.route">{{ block.action.label }}</router-link>
          </div>
        </section></template
      ></async-state
    >
  </platform-layout>
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
  dashboard = computed(() => load.data.value)
const titles = {
  today_workout: 'Ваше следующее занятие',
  recent_progress: 'Последние результаты',
  ai_message: 'Подсказка на сегодня',
  recommendation: 'Подходящие программы',
  marketplace_program: 'Из каталога',
  achievement: 'Ваши достижения',
  followed_trainers: 'Новое у ваших тренеров',
}
async function refresh() {
  await load.run(async () => (await api.get('/api/v2/dashboard')).data)
}
useViewLoad(refresh)
</script>
