import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { ElAlert, ElAside, ElButton, ElContainer, ElEmpty, ElLoading, ElMain, ElMenu, ElMenuItem, ElOption, ElSelect, ElTable, ElTableColumn, ElTag } from 'element-plus'
import 'element-plus/dist/index.css'
import './style.css'
import App from './App.vue'
import { router } from './router'

const handlingDesk = createApp(App)
const controls = [ElAlert, ElAside, ElButton, ElContainer, ElEmpty, ElMain, ElMenu, ElMenuItem, ElOption, ElSelect, ElTable, ElTableColumn, ElTag]
for (const control of controls) handlingDesk.component(control.name!, control)
handlingDesk.directive('loading', ElLoading.directive)
handlingDesk.use(createPinia())
handlingDesk.use(router)
handlingDesk.mount('#app')
