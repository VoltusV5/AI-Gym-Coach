import { useRoute } from 'vue-router'
import { onMounted, onScopeDispose } from 'vue'
import { onIonViewWillEnter } from '@ionic/vue'

export function useViewLoad(load, canRefresh = () => true) {
  const route = useRoute(),
    pagePath = route.path
  let started = false
  let pending = false
  let alive = true
  onScopeDispose(() => {
    alive = false
  })
  async function refresh() {
    if (route.path !== pagePath || !alive || pending || (started && !canRefresh())) return
    started = true
    pending = true
    try {
      await load()
    } finally {
      pending = false
    }
  }
  onMounted(() => {
    if (!started) void refresh()
  })
  onIonViewWillEnter(refresh)
  return refresh
}
