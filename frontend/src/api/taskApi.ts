import { apiClient } from '@/api/client'
import type { CreateTaskBody, Task, UpdateTaskBody } from '@/types/task'

/**
 * CSRF トークンを取得する。
 * /api/auth/csrf-token は SuccessResponse 形式ではなく { token: string } を直接返す。
 * apiClient を使うことで Authorization ヘッダーが自動付与される。
 */
async function getCsrfToken(): Promise<string> {
  const response = await apiClient.get<{ token: string }>('/api/auth/csrf-token')
  return response.data.token
}

/**
 * "YYYY-MM-DD" を Go が受け付ける RFC 3339 形式に変換する。
 * 空文字や未入力の場合は undefined を返してフィールドを除外する。
 */
function toRFC3339OrUndefined(dateStr: string | undefined): string | undefined {
  if (!dateStr) return undefined
  // 既に RFC 3339 形式（T を含む）ならそのまま返す
  if (dateStr.includes('T')) return dateStr
  return `${dateStr}T00:00:00Z`
}

/**
 * タスク CRUD API
 */
export const taskApi = {
  getAll: async (): Promise<Task[]> => {
    const { data } = await apiClient.get('/api/task/')
    return data.data as Task[]
  },

  getById: async (id: number): Promise<Task> => {
    const { data } = await apiClient.get(`/api/task/${id}`)
    return data.data as Task
  },

  create: async (payload: CreateTaskBody): Promise<Task> => {
    const csrfToken = await getCsrfToken()
    const body = {
      status_id: 1, // デフォルト: todo
      ...payload,
      due_date: toRFC3339OrUndefined(payload.due_date),
    }
    const { data } = await apiClient.post('/api/task/', body, {
      headers: { 'X-CSRF-Token': csrfToken },
    })
    return data.data as Task
  },

  update: async (id: number, payload: UpdateTaskBody): Promise<Task> => {
    const csrfToken = await getCsrfToken()
    const body = { ...payload, due_date: toRFC3339OrUndefined(payload.due_date) }
    const { data } = await apiClient.put(`/api/task/${id}`, body, {
      headers: { 'X-CSRF-Token': csrfToken },
    })
    return data.data as Task
  },

  updateStatus: async (id: number, statusId: number): Promise<void> => {
    const csrfToken = await getCsrfToken()
    await apiClient.patch(
      `/api/task/${id}`,
      { status_id: statusId },
      { headers: { 'X-CSRF-Token': csrfToken } }
    )
  },

  delete: async (id: number): Promise<void> => {
    const csrfToken = await getCsrfToken()
    await apiClient.delete(`/api/task/${id}`, {
      headers: { 'X-CSRF-Token': csrfToken },
    })
  },

  exportCsv: async (): Promise<Blob> => {
    const { data } = await apiClient.get('/api/tasks/export/csv', {
      responseType: 'blob',
    })
    return data as Blob
  },
}
