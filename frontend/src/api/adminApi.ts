import { apiClient } from '@/api/client'
import type { Task } from '@/types/task'
import type { User } from '@/types/user'

/**
 * 管理者専用 API (role=admin のみ)
 */
export const adminApi = {
  getUsers: async (): Promise<User[]> => {
    const { data } = await apiClient.get('/api/admin/users')
    return data.data as User[]
  },

  getTasks: async (): Promise<Task[]> => {
    const { data } = await apiClient.get('/api/admin/tasks')
    return data.data as Task[]
  },
}
