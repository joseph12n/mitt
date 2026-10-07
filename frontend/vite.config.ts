import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  build: {
    // The Go binary embeds this directory (internal/web/web.go); the bar LAN
    // may be offline, so the bundle must stay fully self-contained.
    outDir: '../internal/web/static/dist',
    emptyOutDir: true,
  },
})
