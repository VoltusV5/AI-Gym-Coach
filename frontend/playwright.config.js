import { defineConfig } from '@playwright/test'

const baseURL = 'http://127.0.0.1:55173'
export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  timeout: 20000,
  use: {
    baseURL,
    trace: 'off',
    video: process.env.SPORTIK_E2E_VIDEO === '1' ? 'on' : 'off',
    launchOptions: process.env.SPORTIK_CHROMIUM_PATH ? { executablePath: process.env.SPORTIK_CHROMIUM_PATH } : {},
  },
  webServer: {
    command: 'node node_modules/vite/bin/vite.js --host 127.0.0.1 --port 55173',
    url: baseURL,
    env: { VITE_USE_MOCK: 'false', VITE_USE_WORKOUT_MOCKS: 'false', VITE_API_URL: '' },
    reuseExistingServer: false,
  },
})
