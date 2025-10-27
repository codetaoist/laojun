import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import dts from 'vite-plugin-dts';
import { resolve } from 'path';

export default defineConfig({
  plugins: [
    react(),
    dts({
      insertTypesEntry: true,
      include: ['src/**/*'],
      exclude: ['src/**/*.test.*', 'src/**/*.spec.*'],
    }),
  ],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src'),
    },
  },
  build: {
    lib: {
      entry: {
        index: resolve(__dirname, 'src/index.ts'),
        api: resolve(__dirname, 'src/api/index.ts'),
        components: resolve(__dirname, 'src/components/index.ts'),
        router: resolve(__dirname, 'src/router/index.ts'),
        stores: resolve(__dirname, 'src/stores/index.ts'),
        utils: resolve(__dirname, 'src/utils/index.ts'),
      },
      name: 'LaojunFrontendShared',
      formats: ['es', 'cjs'],
      fileName: (format, entryName) => {
        const extension = format === 'es' ? 'esm.js' : 'js';
        return `${entryName}/index.${extension}`;
      },
    },
    rollupOptions: {
      external: [
        'react',
        'react-dom',
        'react-router-dom',
        'antd',
        '@ant-design/icons',
        'axios',
        'zustand',
        'dayjs',
        'lodash-es',
      ],
      output: {
        globals: {
          react: 'React',
          'react-dom': 'ReactDOM',
          'react-router-dom': 'ReactRouterDOM',
          antd: 'antd',
          '@ant-design/icons': 'AntdIcons',
          axios: 'axios',
          zustand: 'zustand',
          dayjs: 'dayjs',
          'lodash-es': 'lodash',
        },
      },
    },
    sourcemap: true,
    minify: 'terser',
  },
});