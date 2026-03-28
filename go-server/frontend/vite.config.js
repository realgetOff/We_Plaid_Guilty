import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig({
  plugins: [react()],
  base: '/', // Ensures absolute paths for assets
  build: {
    outDir: 'dist',
  }
})