<template>
  <article ref="card" class="platform-card">
    <span class="platform-badge">{{ levelNames[program.level] || program.level }}</span>
    <h2 class="program-title">
      <router-link :to="`/programs/${program.id}`">{{ program.title }}</router-link>
    </h2>
    <p>
      <router-link :to="`/trainers/${program.trainer.id}`">{{
        program.trainer.display_name
      }}</router-link
      ><span v-if="program.trainer.verified"> · Проверенный тренер</span>
    </p>
    <p class="platform-muted">
      {{ program.duration_weeks }} нед. · {{ program.sessions_per_week }} занятий в неделю
    </p>
    <p>
      {{ program.price_minor === 0 ? 'Бесплатно' : money(program.price_minor, program.currency) }}
    </p>
    <p v-if="program.rating != null" class="platform-muted">
      Оценка {{ program.rating.toFixed(1) }} · отзывов {{ program.review_count }}
    </p>
    <slot />
  </article>
</template>
<script setup>
import { ref, onMounted, onScopeDispose } from 'vue'
import { sendProgramEvent } from '@/utils/events'
import { money, levelNames } from '@/utils/platform'
const props = defineProps({
  program: { type: Object, required: true },
  surface: { type: String, default: 'catalog' },
})
const card = ref(null)
let observer
onMounted(() => {
  if (!props.surface) return
  observer = new IntersectionObserver(
    (entries) => {
      if (entries.some((entry) => entry.isIntersecting)) {
        sendProgramEvent('impression', props.program.id, props.surface)
        observer.disconnect()
      }
    },
    { threshold: 0.5 },
  )
  if (card.value) observer.observe(card.value)
})
onScopeDispose(() => observer?.disconnect())
</script>
<style scoped>
.program-title {
  margin-top: 15px;
  font-size: 1.15rem;
  line-height: 1.4;
}
.program-title a {
  text-decoration: none;
}
</style>
