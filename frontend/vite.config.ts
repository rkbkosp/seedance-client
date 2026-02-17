import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': '/src'
    }
  },
  server: {
    port: 5173,
    strictPort: true,
    hmr: {
      overlay: false
    }
  },
  build: {
    outDir: 'dist',
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (/node_modules\/(vue|pinia|vue-router)\//.test(id)) {
            return 'framework'
          }
          if (/node_modules\/(naive-ui|vueuc|vooks|vdirs|css-render|date-fns|async-validator)\//.test(id)) {
            return 'ui'
          }
          if (id.includes('/src/storyboard_v2.js') || id.includes('/src/components/LegacyStoryboardV2Bridge.vue')) {
            return 'legacy-v2'
          }
          return undefined
        },
      },
    },
  }
})
