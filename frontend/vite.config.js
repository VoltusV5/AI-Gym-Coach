import { fileURLToPath, URL } from 'node:url'
import crypto from 'node:crypto'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import vueDevTools from 'vite-plugin-vue-devtools'
import Components from 'unplugin-vue-components/vite'

function sportikBuildLabel() {
  const iso = process.env.SOURCE_DATE_EPOCH
    ? new Date(Number(process.env.SOURCE_DATE_EPOCH) * 1000).toISOString()
    : new Date().toISOString()
  const uniq = crypto.randomBytes(4).toString('hex')
  return `${iso}#${uniq}`
}

/** Один раз за запуск vite — и meta, и define совпадают */
const sportikBuildLabelOnce = sportikBuildLabel()

// https://vite.dev/config/
export default defineConfig({
  define: {
    __SPORTIK_BUILD_ID__: JSON.stringify(sportikBuildLabelOnce)
  },
  server: {
    host: true,
    port: 5173,
    strictPort: true,
    // В dev фронт ходит на тот же origin (пустой baseURL в api.js) — только /api/* на Go, без CORS.
    proxy: {
      '/api': { target: 'http://127.0.0.1:5050', changeOrigin: true }
    }
  },
  plugins: [
    {
      name: 'sportik-html-build-meta',
      transformIndexHtml(html) {
        const id = sportikBuildLabelOnce
        return html.replace(
          '<head>',
          `<head>\n    <meta name="sportik-build" content="${id}" />`
        )
      }
    },
    // Не помечать ion-* как native custom elements: в @ionic/vue это Vue-компоненты.
    // isCustomElement: ion-* ломает шаблоны — <ion-page> не связывается с IonPage → пустой экран.
    vue(),

    vueDevTools(),

    Components({
      dirs: ['src/components'],
      extensions: ['vue'],
      dts: 'src/components.d.ts',
      deep: true
    })
  ],

  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  }
})