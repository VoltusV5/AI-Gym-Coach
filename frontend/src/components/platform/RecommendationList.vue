<template>
  <div ref="container" class="platform-grid">
    <article v-for="item in items" :key="item.program_id" class="recommendation-item">
      <program-card
        :program="item.program"
        surface=""
        @click.capture="recordClick($event, item.program_id)"
      />
      <p class="platform-muted">{{ item.explanation }}</p>
      <p class="ranking-note">Порядок отражает соответствие программе, а не прогноз результата.</p>
    </article>
  </div>
</template>
<script setup>
import { ref, watch, nextTick, onScopeDispose } from 'vue'
import { sendProgramEvent } from '@/utils/events'
import ProgramCard from '@/components/platform/ProgramCard.vue'
const props = defineProps({
  items: { type: Array, required: true },
  requestId: { type: String, default: '' },
})
const container = ref(null)
let alive = true
let observer
const seen = new Set()
watch(
  () => props.items,
  async () => {
    await nextTick()
    if (!alive) return
    observer?.disconnect()
    observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          if (!entry.isIntersecting) continue
          const id = entry.target.dataset.programId
          if (!seen.has(id)) {
            seen.add(id)
            sendProgramEvent('impression', id, 'recommendations', props.requestId)
          }
          observer.unobserve(entry.target)
        }
      },
      { threshold: 0.5 },
    )
    container.value?.querySelectorAll('.recommendation-item').forEach((element, index) => {
      if (!props.items[index]) return
      element.dataset.programId = props.items[index].program_id
      observer.observe(element)
    })
  },
  { immediate: true },
)
function recordClick(event, id) {
  if (event.target.closest('a'))
    sendProgramEvent('recommendation_click', id, 'recommendations', props.requestId)
}
onScopeDispose(() => {
  alive = false
  observer?.disconnect()
})
</script>
<style scoped>
.ranking-note {
  font-size: 0.78rem;
  color: var(--sportik-text-muted);
}
</style>
