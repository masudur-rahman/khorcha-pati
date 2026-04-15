import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
    plugins: [react(), tailwindcss()],
    server: {
        host: '0.0.0.0',
        port: 5173,
        strictPort: true,
        allowedHosts: ["xpense.mrahman.xyz"],
        proxy: {
            '/api': {
                target: 'https://xpense-api.mrahman.xyz',
                changeOrigin: true,
                secure: false,
            },
        },
    },
})