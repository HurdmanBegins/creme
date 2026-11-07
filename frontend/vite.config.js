import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 5173,
    proxy: {
      // En dev : `npm run dev` ici + `go run .` dans backend/ (port 8080)
      '/api': { target: 'http://localhost:8080', changeOrigin: true },
    },
  },
})
