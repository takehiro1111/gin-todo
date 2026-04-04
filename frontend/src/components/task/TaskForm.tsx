import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z } from 'zod'

const taskSchema = z.object({
  title: z.string().min(1, '必須項目です').max(100, '100文字以内で入力してください'),
  description: z.string().max(300, '300文字以内で入力してください').optional(),
  priority: z.enum(['low', 'medium', 'high']).default('medium'),
  due_date: z.string().optional(),
})

type TaskFields = z.infer<typeof taskSchema>

interface TaskFormProps {
  defaultValues?: Partial<TaskFields>
  onSubmit: (data: TaskFields) => Promise<void>
  isLoading?: boolean
  submitLabel?: string
}

/**
 * タスク作成・編集共通フォーム。
 */
export function TaskForm({
  defaultValues,
  onSubmit,
  isLoading = false,
  submitLabel = '保存',
}: TaskFormProps) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<TaskFields>({
    resolver: zodResolver(taskSchema),
    defaultValues,
  })

  return (
    <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4">
      <Input label="タイトル *" error={errors.title?.message} {...register('title')} />
      <div className="flex flex-col gap-1">
        <label htmlFor="task-description" className="text-sm font-medium text-gray-700">
          説明
        </label>
        <textarea
          id="task-description"
          {...register('description')}
          rows={3}
          className="px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
        />
        {errors.description && <p className="text-sm text-red-600">{errors.description.message}</p>}
      </div>
      <div className="flex flex-col gap-1">
        <label htmlFor="task-priority" className="text-sm font-medium text-gray-700">
          優先度
        </label>
        <select
          id="task-priority"
          {...register('priority')}
          className="px-3 py-2 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
        >
          <option value="low">低</option>
          <option value="medium">中</option>
          <option value="high">高</option>
        </select>
      </div>
      <Input label="期限" type="date" error={errors.due_date?.message} {...register('due_date')} />
      <Button type="submit" isLoading={isLoading}>
        {submitLabel}
      </Button>
    </form>
  )
}
