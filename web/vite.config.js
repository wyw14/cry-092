import { defineConfig, loadEnv } from 'vite';
import vue from '@vitejs/plugin-vue';
const domainChunks = (id) => {
    if (id.includes('element-plus'))
        return 'element-plus';
    if (id.includes('vue') || id.includes('pinia'))
        return 'vue-runtime';
    return undefined;
};
export default defineConfig(({ mode }) => {
    const environment = loadEnv(mode, '.', 'CRY092_');
    return {
        plugins: [vue()],
        server: { host: '127.0.0.1', proxy: { '/api': environment.CRY092_API_ORIGIN || 'http://localhost:8080' } },
        build: { target: 'es2022', sourcemap: false, rollupOptions: { output: { manualChunks: domainChunks } } },
        test: { environment: 'jsdom', include: ['src/tests/**/*.test.ts'], restoreMocks: true }
    };
});
