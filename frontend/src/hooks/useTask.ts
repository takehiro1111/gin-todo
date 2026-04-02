import { taskApi } from '@/api/taskApi'
import { useQuery } from '@tanstack/react-query'

/**
 * 単一タスクを取得するカスタムフック。
 */
export function useTask(id: number, options?: { enabled?: boolean }) {
  return useQuery({
    queryKey: ['tasks', id],
    queryFn: () => taskApi.getById(id),
    enabled: options?.enabled ?? true,
  })
}
