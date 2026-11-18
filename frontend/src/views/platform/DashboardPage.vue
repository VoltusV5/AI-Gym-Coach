<template>
  <platform-layout
    title="Сегодня — шаг вперёд"
    subtitle="Ваш план, последние результаты и следующие шаги в одном месте."
  >
    <template #actions><router-link to="/plans">Все планы</router-link></template>
    <async-state :loading="load.loading.value" :error="load.error.value" @retry="refresh">
      <div v-if="dashboard" class="dashboard-feed">
        <section
          v-for="block in dashboard.blocks"
          :key="block.id"
          class="platform-panel dashboard-block"
          :class="`dashboard-block--${block.type}`"
        >
          <div class="dashboard-block-heading">
            <h2>{{ titles[block.type] || 'Для вас' }}</h2>
            <span v-if="block.type === 'today_workout'" class="platform-badge">Сегодня</span>
          </div>
          <p v-if="block.status === 'unavailable'" class="platform-muted">
            Этот блок временно недоступен. Остальные данные доступны.
          </p>
          <template v-else-if="block.type === 'recommendation'">
            <p v-if="!block.data.items?.length" class="platform-muted">
              Пока нет подходящих программ.
            </p>
            <recommendation-list
              v-else
              :items="block.data.items"
              :request-id="block.data.request_id"
            />
          </template>
          <template v-else-if="block.type === 'followed_trainers'">
            <p v-if="!block.data.items.length" class="platform-muted">
              Подпишитесь на тренеров, чтобы видеть их новые программы.
            </p>
            <p
              v-for="item in block.data.items"
              :key="item.program_id"
              class="dashboard-publication"
            >
              <router-link :to="`/programs/${item.program_id}`">{{ item.title }}</router-link
              ><span>{{ item.display_name }}</span>
            </p>
          </template>
          <template v-else-if="block.type === 'recent_progress'">
            <p class="dashboard-period">
              {{ block.data.metrics ? 'Последняя тренировка' : metricPeriod(block.data)
              }}<span v-if="metricRange(block.data)"> · {{ metricRange(block.data) }}</span>
            </p>
            <dl class="dashboard-stats">
              <div v-if="block.data.workout_count != null">
                <dt>Тренировок</dt>
                <dd>{{ metricNumber(block.data.workout_count) }}</dd>
              </div>
              <div>
                <dt>{{ block.data.metrics ? 'Объём тренировки' : 'Объём за период' }}</dt>
                <dd>
                  {{ metricNumber((block.data.metrics || block.data).volume_kg) }} <small>кг</small>
                </dd>
              </div>
              <div v-if="block.data.metrics">
                <dt>Рабочих подходов</dt>
                <dd>{{ metricNumber(block.data.metrics.working_sets) }}</dd>
              </div>
            </dl>
            <volume-help />
          </template>
          <template v-else>
            <h3 v-if="block.data.title" class="dashboard-plan-title">{{ block.data.title }}</h3>
            <p v-if="block.data.message" class="dashboard-message">{{ block.data.message }}</p>
            <p v-if="block.data.workout_count != null" class="platform-muted">
              Занятий: {{ metricNumber(block.data.workout_count) }}
            </p>
          </template>
          <div
            v-if="block.action?.route?.startsWith('/') && !block.action.route.startsWith('//')"
            class="platform-actions"
          >
            <router-link
              :to="block.action.route"
              :class="
                block.type === 'today_workout' ? 'platform-button primary' : 'dashboard-action'
              "
              >{{ block.action.label }}</router-link
            >
          </div>
        </section>
      </div>
    </async-state>
  </platform-layout>
</template>
<script setup>
import { computed } from 'vue'
import { metricNumber, metricPeriod, metricRange } from '@/utils/metrics'
import VolumeHelp from '@/components/ui/VolumeHelp.vue'
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
