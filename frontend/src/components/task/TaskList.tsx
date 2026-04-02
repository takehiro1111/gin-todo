import Skeleton from 'react-loading-skeleton'
import 'react-loading-skeleton/dist/skeleton.css'
import { TaskCard } from '@/components/task/TaskCard'
import type { Task } from '@/types/task'

interface TaskListProps {
  tasks: Task[] | undefined
  isLoading: boolean
}

/**
 * タスク一覧を表示するコンポーネント。ローディング中はスケルトンを表示する。
 */
export function TaskList({ tasks, isLoading }: TaskListProps) {
  if (isLoading) {
    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {Array.from({ length: 6 }, (_, i) => (
          // biome-ignore lint/suspicious/noArrayIndexKey: スケルトンは静的なプレースホルダーのため順序が変わらない
          <Skeleton key={`skeleton-${i}`} height={120} className="rounded-lg" />
        ))}
      </div>
    )
  }

  if (!tasks || tasks.length === 0) {
    return (
      <div className="text-center py-12 text-gray-500">
        タスクがありません。新しいタスクを作成してください。
      </div>
    )
  }

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      {tasks.map((task) => (
        <TaskCard key={task.id} task={task} />
      ))}
    </div>
  )
}
