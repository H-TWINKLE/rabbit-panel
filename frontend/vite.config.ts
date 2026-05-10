import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

// https://vite.dev/config/
export default defineConfig({
  plugins: [
    vue(),
    AutoImport({
      resolvers: [ElementPlusResolver()],
      imports: ['vue', 'vue-router', 'pinia'],
      dts: 'src/auto-imports.d.ts',
    }),
    Components({
      resolvers: [ElementPlusResolver()],
      dts: 'src/components.d.ts',
    }),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  server: {
    port: 3000,
    proxy: {
      '/api': {
        target: 'http://localhost:3958',
        changeOrigin: true,
      },
      '/ws': {
        target: 'ws://localhost:3958',
        ws: true,
      },
    },
  },
  build: {
    // 输出到 backend/dist 目录，供 Go 后端嵌入
    outDir: '../backend/.dist',
    emptyOutDir: true,
    // 优化打包体积
    rollupOptions: {
      output: {
        // 分离第三方库，优化缓存
        manualChunks: {
          'vue-vendor': ['vue', 'vue-router', 'pinia'],
          'element-plus': ['element-plus'],
          'xterm': ['@xterm/xterm', '@xterm/addon-fit', '@xterm/addon-web-links'],
        },
        // 优化资源文件命名
        chunkFileNames: 'assets/[name]-[hash].js',
        entryFileNames: 'assets/[name]-[hash].js',
        assetFileNames: 'assets/[name]-[hash].[ext]',
      },
    },
    // 启用 CSS 代码分割
    cssCodeSplit: true,
    // 设置 chunk 大小警告阈值 (500KB)
    chunkSizeWarningLimit: 500,
    // 压缩选项
    minify: 'esbuild',
    // 生成 sourcemap 用于调试（生产环境可设为 false）
    sourcemap: false,
    // 目标浏览器
    target: 'es2015',
  },
})

