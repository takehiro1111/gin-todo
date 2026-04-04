import { taskApi } from '@/api/taskApi'
import type { CreateTaskBody } from '@/types/task'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import toast from 'react-hot-toast'

/**
 * タスクを新規作成するカスタムフック。
 */
export function useCreateTask() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (payload: CreateTaskBody) => taskApi.create(payload),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['tasks'] })
      toast.success('タスクを作成しました')
    },
    onError: () => {
      toast.error('タスクの作成に失敗しました')
    },
  })
}
