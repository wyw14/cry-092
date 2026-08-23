import { computed, type Ref } from 'vue'
export function useTimezone(value: Ref<string | undefined>, timezone = 'Asia/Shanghai') {
  return computed(() => value.value ? new Intl.DateTimeFormat('zh-CN', { timeZone: timezone, dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value.value)) : '-')
}
