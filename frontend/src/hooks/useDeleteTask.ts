import { taskApi } from '@/api/taskApi'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'

/**
 * タスクを削除するカスタムフック。
 */
export function useDeleteTask() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: number) => taskApi.delete(id),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['tasks'] })
      toast.success('タスクを削除しました')
    },
    onError: () => {
      toast.error('タスクの削除に失敗しました')
    },
  })
}
