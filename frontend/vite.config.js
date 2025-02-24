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
    },
    build: {
      assetsInlineLimit: 0, // Disable inlining assets
      rollupOptions: {
        output: {
          assetFileNames: (assetInfo) => {
            const info = assetInfo.name.split('.')
            const ext = info[info.length - 1]
            if (/\.(woff2?|eot|ttf|otf)(\?.*)?$/i.test(assetInfo.name)) {
              return `assets/fonts/[name].[hash].[ext]`
            }
            return `assets/[name].[hash].[ext]`
          }
        }
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
    config.build.outDir = '../backend/dist'
    config.build.emptyOutDir = true
    config.build.manifest = true
  }

  return config
})
