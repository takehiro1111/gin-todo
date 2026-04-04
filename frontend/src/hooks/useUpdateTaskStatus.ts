import { taskApi } from '@/api/taskApi'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'

/**
 * タスクのステータスのみを更新するカスタムフック。
 */
export function useUpdateTaskStatus() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, statusId }: { id: number; statusId: number }) =>
      taskApi.updateStatus(id, statusId),
    onSuccess: async (_data, { id }) => {
      await Promise.all([
        queryClient.invalidateQueries({ queryKey: ['tasks'] }),
        queryClient.invalidateQueries({ queryKey: ['tasks', id] }),
      ])
      toast.success('ステータスを更新しました')
    },
    onError: () => {
      toast.error('ステータスの更新に失敗しました')
    },
  })
}
