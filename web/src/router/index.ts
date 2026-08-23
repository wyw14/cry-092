import { createRouter, createWebHistory } from 'vue-router'

declare module 'vue-router' {
  interface RouteMeta { title: string; roles: string[] }
}

export const router = createRouter({
  history: createWebHistory('/'),
  routes: [
    { path: '/', name: 'handling-desk', component: () => import('../pages/WorkbenchPage.vue'), meta: { title: '建议办理工作台', roles: ['representative', 'unit_officer', 'supervisor'] } },
    { path: '/archives', name: 'historical-performance', component: () => import('../pages/ArchivePage.vue'), meta: { title: '历年档案与单位排名', roles: ['supervisor', 'administrator'] } }
  ],
  scrollBehavior: () => ({ top: 0 })
})

router.afterEach((route) => { document.title = `${String(route.meta.title)} | 代表建议平台` })
