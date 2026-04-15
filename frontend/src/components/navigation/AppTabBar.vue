<template>
  <nav class="app-tabbar" aria-label="˜˜˜˜˜˜ ˜˜˜˜">
    <button
      v-for="item in items"
      :key="item.key"
      type="button"
      class="tab-btn"
      :class="{ 'tab-btn--active': item.key === activeKey }"
      @click="onTabClick(item.key)"
    >
      <img v-if="item.src" class="tab-icon" :src="item.src" alt="" />
      <span v-else class="tab-fallback" aria-hidden="true">˜</span>
    </button>
  </nav>
</template>

<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { getHomeTabIconUrls } from '@/utils/localImages'

const props = defineProps({
  activeKey: {
    type: String,
    default: 'main'
  }
})

const router = useRouter()
const icons = getHomeTabIconUrls()

const items = computed(() => [
  { key: 'workout', src: icons.workout },
  { key: 'notes', src: icons.notes },
  { key: 'main', src: icons.main },
  { key: 'nutrition', src: icons.nutrition },
  { key: 'settings', src: icons.settings }
])

function onTabClick(key) {
  switch (key) {
    case 'workout':
      router.push('/workout-tools')
      return
    case 'main':
      router.push('/home')
      return
    case 'notes':
      router.push('/notes')
      return
    case 'settings':
      router.push('/settings')
      return
    case 'nutrition':
      router.push('/nutrition')
      return
    default:
  }
}
</script>

<style scoped>
.app-tabbar {
  display: flex;
  justify-content: space-around;
  align-items: center;
  padding: 8px 8px 10px;
  background: var(--sportik-cream);
  border-top: 1px solid rgba(0, 0, 0, 0.08);
}

.tab-btn {
  flex: 1;
  max-width: 64px;
  padding: 6px;
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0.72;
}

.tab-btn--active {
  opacity: 1;
}

.tab-icon {
  width: 36px;
  height: 36px;
  object-fit: contain;
  display: block;
}

.tab-fallback {
  width: 36px;
  height: 36px;
  line-height: 36px;
  text-align: center;
  color: var(--sportik-text-muted);
  font-size: 1.5rem;
}
</style>
