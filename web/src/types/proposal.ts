export const workflowStates = [
  'draft',
  'submitted',
  'assigned',
  'accepted',
  'handling',
  'answered',
  'evaluated',
  'archived'
] as const

export type WorkflowState = typeof workflowStates[number]

export type SuggestionListRow = Readonly<{
  id: string
  title: string
  category: string
  status: WorkflowState
  representative_id: string
  unit_id?: string
  submitted_at: string
  due_at?: string
  overdue: boolean
}>

export type PageEnvelope<Item> = Readonly<{
  items: Item[]
  next_cursor?: string
  page: number
  page_size: number
  total: number
}>

export type SuggestionPage = PageEnvelope<SuggestionListRow>

export type ServiceFailure = Readonly<{
  code: string
  message: string
  field_errors: ReadonlyArray<Readonly<{ field: string; message: string }>>
  request_id: string
}>
