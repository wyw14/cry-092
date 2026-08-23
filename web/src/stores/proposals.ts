import { defineStore } from 'pinia'
import { proposalApi } from '../services/api'
import type { SuggestionListRow } from '../types/proposal'

export const useProposalStore = defineStore('proposals', {
  state: () => ({ items: [] as SuggestionListRow[], total: 0, loading: false, error: '', page: 1, status: '', sort: 'submitted_at' }),
  actions: {
    async load() {
      this.loading = true; this.error = ''
      try {
        const params = new URLSearchParams({ page: String(this.page), page_size: '20', sort: this.sort })
        if (this.status) params.set('status', this.status)
        const page = await proposalApi.list(params)
        this.items = page.items; this.total = page.total
      } catch (error) { this.error = typeof error === 'object' && error && 'message' in error ? String(error.message) : '加载失败' }
      finally { this.loading = false }
    }
  }
})
