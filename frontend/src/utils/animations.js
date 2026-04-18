import { createAnimation } from '@ionic/vue'

export const noAnimation = (baseEl, opts) => {
  const anim = createAnimation().duration(0)
  if (baseEl) {
    anim.addElement(baseEl)
  }
  return anim
}