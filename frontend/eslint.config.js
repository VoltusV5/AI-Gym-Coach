import { defineConfig, globalIgnores } from 'eslint/config'
import globals from 'globals'
import js from '@eslint/js'
import pluginVue from 'eslint-plugin-vue'
import pluginOxlint from 'eslint-plugin-oxlint'
import skipFormatting from 'eslint-config-prettier/flat'

export default defineConfig([
  {
    name: 'app/files-to-lint',
    files: ['**/*.{vue,js,mjs,jsx}'],
  },

  globalIgnores(['**/dist/**', '**/dist-ssr/**', '**/coverage/**', '**/test-results/**', '**/playwright-report/**']),

  {
    languageOptions: {
      globals: {
        ...globals.browser,
        __SPORTIK_BUILD_ID__: 'readonly',
      },
    },
  },

  js.configs.recommended,
  ...pluginVue.configs['flat/essential'],

  ...pluginOxlint.buildFromOxlintConfigFile('.oxlintrc.json'),

  {
    files: ['*.config.js', 'tests/**/*.js'],
    languageOptions: { globals: globals.node }
  },
  {
    rules: {
      'no-empty': ['error', { allowEmptyCatch: true }],
      'vue/no-deprecated-slot-attribute': ['error', { ignore: ['/^ion-/'], ignoreParents: ['/^ion-/'] }]
    }
  },
  skipFormatting,
])
