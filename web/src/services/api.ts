import type { ServiceFailure, SuggestionPage } from '../types/proposal'

const base = '/api/v1'
let accessToken = ''
export const setAccessToken = (value: string) => { accessToken = value }

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  const controller = new AbortController()
  const timer = window.setTimeout(() => controller.abort(), 8000)
  try {
    const response = await fetch(`${base}${path}`, { ...init, signal: controller.signal, headers: { 'Content-Type': 'application/json', ...(accessToken ? { Authorization: `Bearer ${accessToken}` } : {}), ...init.headers } })
    if (!response.ok) throw await response.json() as ServiceFailure
    if (response.status === 204) return undefined as T
    return await response.json() as T
  } finally { window.clearTimeout(timer) }
}

export const proposalApi = {
  list: (params: URLSearchParams) => request<SuggestionPage>(`/proposals?${params}`),
  submit: (id: string) => request(`/proposals/${id}/submit`, { method: 'POST' }),
  accept: (id: string, unitId: string) => request(`/proposals/${id}/accept`, { method: 'POST', body: JSON.stringify({ unit_id: unitId }) }),
  evaluate: (id: string, replyId: string, rating: string, comment: string) => request(`/proposals/${id}/evaluations`, { method: 'POST', body: JSON.stringify({ reply_id: replyId, rating, comment }) })
}
