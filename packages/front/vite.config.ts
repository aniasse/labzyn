import react from '@vitejs/plugin-react';
import path from 'node:path';
import { defineConfig } from 'vite';

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { '@': path.resolve(__dirname, './src') },
  },
  server: {
    port: 4001,
    proxy: {
      // Proxy API calls to the Go server (labzyn-server on :4000).
      '/api': 'http://localhost:4000',
    },
  },
});
