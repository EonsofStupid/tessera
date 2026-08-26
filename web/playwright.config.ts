import { defineConfig, devices } from '@playwright/test'

const baseURL = process.env.NOMEN_E2E_BASE_URL ?? 'https://127.0.0.1:8089'

export default defineConfig({
  testDir: './e2e',
  outputDir: './test-results',
  fullyParallel: false,
  forbidOnly: true,
  retries: 0,
  workers: 1,
  reporter: [['line']],
  use: {
    baseURL,
    // The local production topology uses a short-lived self-signed certificate
    // at its TLS terminator. Deployed environments use an operator-managed CA.
    ignoreHTTPSErrors: true,
    screenshot: 'off',
    trace: 'off',
    video: 'off',
  },
  projects: [
    {
      name: 'runtime-baseline',
      use: { ...devices['Desktop Chrome'] },
      testMatch: 'runtime-baseline.spec.ts',
    },
    {
      name: 'deployment-preflight',
      use: { ...devices['Desktop Chrome'] },
      testMatch: 'deployment-preflight.spec.ts',
    },
    {
      name: 'bootstrap-owner',
      use: { ...devices['Desktop Chrome'] },
      testMatch: 'bootstrap-owner.spec.ts',
    },
  ],
})
