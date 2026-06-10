import { defineConfig } from 'vitest/config';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  test: {
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    globals: true
  },
  build: {
    outDir: '../dist',
    emptyOutDir: true
  },
  server: {
    port: 5173
  }
});
