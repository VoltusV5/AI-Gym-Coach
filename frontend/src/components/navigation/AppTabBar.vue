<template>
  <nav class="app-tabbar" :aria-label="ui.ariaLabel">
    <button
      v-for="item in items"
      :key="item.key"
      type="button"
      class="tab-btn"
      :class="{ 'tab-btn--active': item.key === activeKey }"
      @click="onTabClick(item.key)"
    >
      <span class="tab-icon-wrap">
        <img v-if="item.src" class="tab-icon" :src="item.src" alt="" />
        <span v-else class="tab-fallback" aria-hidden="true">&#8226;</span>
      </span>
      <span class="tab-label">{{ item.label }}</span>
    </button>
  </nav>
</template>

<script setup>
import { computed } from 'vue'
import { useIonRouter } from '@ionic/vue'
import { MAIN_TABS } from '@/config/mainTabs'
import { getHomeTabIconUrls } from '@/utils/localImages'
import { noAnimation } from '@/utils/animations'

const props = defineProps({
  activeKey: {
    type: String,
    default: 'main'
  }
})

const ui = {
  ariaLabel: '\u041D\u0438\u0436\u043D\u0435\u0435 \u043C\u0435\u043D\u044E',
  workout: '\u0422\u0440\u0435\u043D\u0438\u0440\u043E\u0432\u043A\u0438',
  notes: '\u0417\u0430\u043C\u0435\u0442\u043A\u0438',
  main: '\u0413\u043B\u0430\u0432\u043D\u0430\u044F',
  nutrition: '\u041F\u0438\u0442\u0430\u043D\u0438\u0435',
  settings: '\u041D\u0430\u0441\u0442\u0440\u043E\u0439\u043A\u0438'
}

const ionRouter = useIonRouter()
const icons = getHomeTabIconUrls()

const items = computed(() =>
  MAIN_TABS.map((t) => ({
    key: t.key,
    src: icons[t.key],
    label: ui[t.key]
  }))
)

function onTabClick(key) {
  /** replace — не копим push-стек при переключении корневых вкладок.
   *  noAnimation — убираем анимацию при переключении табов. */
  switch (key) {
    case 'workout':
      ionRouter.navigate('/workout-tools', 'root', 'replace', noAnimation)
      return
    case 'main':
      ionRouter.navigate('/home', 'root', 'replace', noAnimation)
      return
    case 'notes':
      ionRouter.navigate('/notes', 'root', 'replace', noAnimation)
      return
    case 'settings':
      ionRouter.navigate('/settings', 'root', 'replace', noAnimation)
      return
    case 'nutrition':
      ionRouter.navigate('/nutrition', 'root', 'replace', noAnimation)
      return
    default:
  }
}
</script>

<style scoped>
.app-tabbar {
  display: flex;
  justify-content: space-between;
  align-items: stretch;
  gap: 6px;
  padding: 10px 10px 12px;
  background: var(--sportik-surface-glass);
  border-top: 1px solid color-mix(in srgb, var(--sportik-border) 70%, transparent);
  backdrop-filter: blur(14px);
}

.tab-btn {
  flex: 1;
  min-width: 0;
  border-radius: 14px;
  padding: 7px 4px;
  border: none;
  background: transparent;
  cursor: pointer;
  display: flex;
  flex-direction: column;
  gap: 4px;
  align-items: center;
  justify-content: center;
  opacity: 0.84;
  transition: all 0.2s ease;
}

.tab-btn--active {
  opacity: 1;
  background: color-mix(in srgb, var(--sportik-brand) 18%, transparent);
  box-shadow: inset 0 0 0 1px color-mix(in srgb, var(--sportik-brand) 40%, transparent);
}

.tab-icon-wrap {
  width: 38px;
  height: 38px;
  border-radius: 12px;
  display: grid;
  place-items: center;
}

.tab-icon {
  width: 28px;
  height: 28px;
  object-fit: contain;
  display: block;
}

.tab-fallback {
  width: 28px;
  height: 28px;
  line-height: 28px;
  text-align: center;
  color: var(--sportik-text-muted);
  font-size: 1.2rem;
}

.tab-label {
  font-size: 0.66rem;
  line-height: 1;
  color: var(--sportik-text-soft);
  font-weight: 600;
}

.tab-btn--active .tab-label {
  color: var(--sportik-text);
}
</style>
