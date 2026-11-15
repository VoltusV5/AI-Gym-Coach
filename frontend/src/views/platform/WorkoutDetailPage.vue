<template>
  <platform-layout
    title="Результат тренировки"
    subtitle="Фактически выполненные упражнения и подходы."
    ><template #actions><router-link to="/progress">К прогрессу</router-link></template
    ><async-state :loading="load.loading.value" :error="load.error.value" @retry="refresh"
      ><template v-if="session"
        ><p>{{ new Date(session.finished_at).toLocaleString('ru-RU') }}</p>
        <section v-for="slot in session.slots" :key="slot.position" class="platform-panel">
          <h2>{{ names.name(slot.exercise_id) }}</h2>
          <p v-for="(set, index) in slot.sets" :key="index">
            Подход {{ index + 1 }}: {{ set.reps }} повторений × {{ set.weight_kg }} кг{{
              set.is_warmup ? ' · разминка' : ''
            }}
          </p>
          <router-link :to="`/progress/exercises/${slot.exercise_id}`"
            >История упражнения и рекорды</router-link
          >
        </section></template
      ></async-state
    ></platform-layout
  >
</template>
<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import api from '@/api/api'
import { useRequest } from '@/composables/useAsync'
import { useViewLoad } from '@/composables/useViewLoad'
import { useExerciseNames } from '@/composables/useExerciseNames'
import PlatformLayout from '@/components/platform/PlatformLayout.vue'
import AsyncState from '@/components/platform/AsyncState.vue'
const route = useRoute(),
  load = useRequest(),
  names = useExerciseNames(),
  session = computed(() => load.data.value)
async function refresh() {
  const value = await load.run(
    async () => (await api.get(`/api/v2/workouts/${route.params.id}`)).data,
  )
  if (value) await names.load(value.slots.map((slot) => slot.exercise_id))
}
useViewLoad(refresh)
</script>
