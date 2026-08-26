import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  base: '/ui/console/',
  plugins: [react()],
  build: {
    outDir: '../internal/api/ui/console/static',
    emptyOutDir: true,
    sourcemap: false,
  },
  server: {
    proxy: {
      '/nomen': 'http://127.0.0.1:8080',
      '/oauth': 'http://127.0.0.1:8080',
      '/ui/login': 'http://127.0.0.1:8080',
    },
  },
})
