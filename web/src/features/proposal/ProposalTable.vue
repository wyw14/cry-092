<script setup lang="ts">
import type { SuggestionListRow } from '../../types/proposal'
import StatusTag from '../../components/StatusTag.vue'
defineProps<{ rows: SuggestionListRow[]; loading: boolean }>()
</script>
<template>
  <el-table :data="rows" v-loading="loading" height="calc(100vh - 230px)">
    <el-table-column prop="title" label="建议标题" min-width="260" show-overflow-tooltip />
    <el-table-column prop="category" label="分类" width="140" />
    <el-table-column label="状态" width="110"><template #default="scope"><StatusTag :status="scope.row.status" /></template></el-table-column>
    <el-table-column prop="unit_id" label="承办单位" min-width="160" />
    <el-table-column prop="submitted_at" label="提交时间" width="180" />
    <el-table-column label="期限" width="150"><template #default="scope"><span :class="{ overdue: scope.row.overdue }">{{ scope.row.due_at || '-' }}</span></template></el-table-column>
  </el-table>
</template>
<style scoped>.overdue{color:#c53030;font-weight:700}</style>
