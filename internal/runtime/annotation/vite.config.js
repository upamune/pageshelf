import { defineConfig } from 'vite';

export default defineConfig({
  build: {
    emptyOutDir: true,
    modulePreload: false,
    rollupOptions: {
      output: {
        entryFileNames: 'annotation.js',
        assetFileNames: 'annotation.[ext]',
      },
    },
  },
});
