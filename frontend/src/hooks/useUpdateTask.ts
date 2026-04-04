import { taskApi } from '@/api/taskApi'
import type { UpdateTaskBody } from '@/types/task'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'

/**
 * タスクを更新するカスタムフック。
 */
export function useUpdateTask(id: number) {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (payload: UpdateTaskBody) => taskApi.update(id, payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['tasks'] })
      await queryClient.invalidateQueries({ queryKey: ['tasks', id] })
      toast.success('タスクを更新しました')
    },
    onError: () => {
      toast.error('タスクの更新に失敗しました')
    },
  })
}
