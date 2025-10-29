import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import { VitePWA } from "vite-plugin-pwa"

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd());
  const port = `${env.VITE_PORT ?? '5173'}`;

  return {
    plugins: [
        react(),
        VitePWA({
            registerType: "autoUpdate",
            devOptions: {
                enabled: true,
            },
            manifest: {
                name: "Igaku",
                background_color: "#15161e",
                theme_color: "#15161e",
                icons: [
                    {
                        src: "/icons/512.png",
                        sizes: "512x512",
                        type: "image/png",
                    },
                    {
                        src: "/icons/512-maskable.png",
                        sizes: "512x512",
                        type: "image/png",
                        purpose: "maskable",
                    },
                ],
            },
            workbox: {
                globPatterns: [
                    "**/*.{css,html,js,png,svg}",
                ],
            },
        }),
    ],
    server: {
      host: true,
      port: +port,
    }
  };
});
