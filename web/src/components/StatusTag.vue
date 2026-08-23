<script setup lang="ts">
import { computed } from 'vue'
import type { WorkflowState } from '../types/proposal'
const props = defineProps<{ status: WorkflowState }>()
const labels: Record<WorkflowState, string> = { draft: '草稿', submitted: '待分办', assigned: '待签收', accepted: '已签收', handling: '办理中', answered: '已答复', evaluated: '已评价', archived: '已归档' }
const type = computed(() => props.status === 'handling' ? 'warning' : props.status === 'answered' || props.status === 'evaluated' ? 'success' : 'info')
</script>
<template><el-tag :type="type" effect="plain">{{ labels[status] }}</el-tag></template>
