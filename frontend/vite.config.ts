import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd());
  const port = `${env.VITE_PORT ?? '5173'}`;

  return {
    plugins: [react()],
    server: {
      host: true,
      port: port,
    }
  };
});
