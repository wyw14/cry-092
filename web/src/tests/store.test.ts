import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useProposalStore } from '../stores/proposals'
import { proposalApi } from '../services/api'
vi.mock('../services/api', () => ({ proposalApi: { list: vi.fn() } }))
describe('proposal store', () => { beforeEach(() => setActivePinia(createPinia())); it('keeps server pagination totals', async () => { vi.mocked(proposalApi.list).mockResolvedValue({ items: [], page: 1, page_size: 20, total: 7 }); const store = useProposalStore(); await store.load(); expect(store.total).toBe(7); expect(store.error).toBe('') }) })
