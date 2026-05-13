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


const sportikBuildLabelOnce = sportikBuildLabel()
export default defineConfig({
  define: {
    __SPORTIK_BUILD_ID__: JSON.stringify(sportikBuildLabelOnce)
  },
  server: {
    host: true,
    port: 5173,
    strictPort: true,
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