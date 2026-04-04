import { taskApi } from '@/api/taskApi'
import { useQuery } from '@tanstack/react-query'

/**
 * タスク一覧を取得するカスタムフック。
 */
export function useTasks() {
  return useQuery({
    queryKey: ['tasks'],
    queryFn: taskApi.getAll,
  })
}
