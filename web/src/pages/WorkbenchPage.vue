<script setup lang="ts">
import { onMounted } from 'vue'
import { Refresh, Plus } from '@element-plus/icons-vue'
import ProposalTable from '../features/proposal/ProposalTable.vue'
import { useProposalStore } from '../stores/proposals'
const store = useProposalStore()
onMounted(() => store.load())
</script>
<template>
  <main class="page">
    <header><div><h1>建议办理工作台</h1><p>提交、分办、办理、答复与督办</p></div><el-button type="primary" :icon="Plus">新建议</el-button></header>
    <section class="toolbar">
      <el-select v-model="store.status" clearable placeholder="全部状态" @change="store.load"><el-option label="待分办" value="submitted"/><el-option label="办理中" value="handling"/><el-option label="已答复" value="answered"/></el-select>
      <el-button :icon="Refresh" circle title="刷新" @click="store.load" />
      <span class="count">共 {{ store.total }} 条</span>
    </section>
    <el-alert v-if="store.error" :title="store.error" type="error" show-icon :closable="false" />
    <ProposalTable :rows="store.items" :loading="store.loading" />
  </main>
</template>
<style scoped>.page{padding:24px 28px}header{display:flex;justify-content:space-between;align-items:center;margin-bottom:20px}h1{font-size:24px;margin:0 0 4px}p{margin:0;color:#667085}.toolbar{display:flex;gap:10px;align-items:center;margin-bottom:14px}.toolbar .el-select{width:180px}.count{margin-left:auto;color:#667085}</style>
