import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

export default defineConfig({
    plugins: [react(), tailwindcss()],
    server: {
        host: process.env.VITE_HOST || '0.0.0.0',
        port: Number(process.env.VITE_PORT) || 65262,
        strictPort: true,
        allowedHosts: (process.env.VITE_ALLOWED_HOSTS || 'khorcha-pati.mrahman.xyz').split(','),
        proxy: {
            '/api': {
                target: process.env.VITE_BACKEND_URL || 'https://khorcha-pati-api.mrahman.xyz',
                changeOrigin: true,
                secure: false,
            },
        },
    },
})
