import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { fileURLToPath, URL } from 'node:url'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    // Allow port override via environment variable for testing
    port: process.env.VITE_PORT ? parseInt(process.env.VITE_PORT) : 5173,
    host: '127.0.0.1',  // Explicitly bind to IPv4 localhost
    strictPort: false  // Allow fallback to alternative port if occupied
  }
})
