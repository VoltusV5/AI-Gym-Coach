import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './tests/e2e',
  testMatch: process.env.SPORTIK_REAL_API ? '**/real-api.spec.js' : '**/platform.spec.js',
  outputDir: './test-results-v2',
  timeout: 30000,
  expect: { timeout: 8000 },
  fullyParallel: false,
  workers: 1,
  use: {
    baseURL: 'http://127.0.0.1:5179',
    channel: process.env.PLAYWRIGHT_CHANNEL || 'chromium',
    trace: 'retain-on-failure',
  },
  projects: [
    { name: 'desktop', use: { ...devices['Desktop Chrome'] } },
    { name: 'mobile', use: { ...devices['iPhone 13'], defaultBrowserType: 'chromium' } },
  ],
  webServer: {
    env: { VITE_API_URL: '', VITE_USE_MOCK: 'false', VITE_USE_WORKOUT_MOCKS: 'false' },
    command: 'npm run build && npm run preview -- --host 127.0.0.1 --port 5179',
    url: 'http://127.0.0.1:5179',
    reuseExistingServer: false,
  },
})
