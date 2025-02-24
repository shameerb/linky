import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'url'

// https://vitejs.dev/config/
export default defineConfig(({ command, mode }) => {
  const config = {
    plugins: [vue()],
    resolve: {
      alias: {
        '@': fileURLToPath(new URL('./src', import.meta.url))
      }
    }
  }

  if (command === 'serve') {
    // Development specific settings
    config.server = {
      port: 3000,
      proxy: {
        '/api': {
          target: 'http://localhost:8080',
          changeOrigin: true
        }
      }
    }
  } else {
    // Production build settings
    config.build = {
      outDir: '../backend/dist',
      emptyOutDir: true,
      manifest: true
    }
  }

  return config
})
