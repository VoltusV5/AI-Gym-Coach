import { defineConfig, globalIgnores } from 'eslint/config'
import globals from 'globals'
import js from '@eslint/js'
import pluginVue from 'eslint-plugin-vue'
import skipFormatting from 'eslint-config-prettier/flat'

export default defineConfig([
  globalIgnores(['dist/**', 'dist-ssr/**', 'coverage/**', 'playwright-report/**', 'test-results/**']),
  js.configs.recommended,
  ...pluginVue.configs['flat/essential'],
  {
    files: ['src/**/*.{js,vue}'],
    languageOptions: {
      globals: { ...globals.browser, __SPORTIK_BUILD_ID__: 'readonly' },
    },
  },
  {
    files: ['*.config.js', 'scripts/**/*.mjs', 'tests/**/*.{js,mjs}'],
    languageOptions: { globals: globals.node },
  },
  {
    files: ['src/**/*.vue'],
    rules: { 'vue/no-deprecated-slot-attribute': ['error', { ignore: ['ion-buttons', 'ion-icon', 'ion-spinner'] }] },
  },
  skipFormatting,
])
