import { defineConfig, devices } from "@playwright/test";

// E2E against a production build served by `vite preview` (proxying /v1 to the mock API in
// e2e/mock-api.mjs): no real backend needed. First run: npx playwright install chromium
export default defineConfig({
  testDir: "e2e",
  workers: 1, // the mock API keeps a single session: specs must not log in at the same time
  use: { baseURL: "http://localhost:4173" },
  projects: [{ name: "chromium", use: { ...devices["Desktop Chrome"] } }],
  webServer: [
    {
      command: "node e2e/mock-api.mjs",
      port: 3998,
      reuseExistingServer: false,
    },
    {
      command: "npm run build && npm run preview",
      url: "http://localhost:4173/login",
      reuseExistingServer: false,
      timeout: 180_000,
      env: { API_PROXY_TARGET: "http://localhost:3998" },
    },
  ],
});
