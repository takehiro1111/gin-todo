import type { Task } from '@/types/task'
import clsx from 'clsx'
import { format, isPast } from 'date-fns'
import { ja } from 'date-fns/locale'
import { Link } from 'react-router-dom'

const STATUS_LABEL: Record<number, string> = {
  1: '未着手',
  2: '進行中',
  3: '完了',
  4: 'アーカイブ',
}

const PRIORITY_LABEL: Record<string, string> = {
  low: '低',
  medium: '中',
  high: '高',
}

const PRIORITY_COLOR: Record<string, string> = {
  low: 'bg-green-100 text-green-700',
  medium: 'bg-yellow-100 text-yellow-700',
  high: 'bg-red-100 text-red-700',
}

/**
 * タスクの概要を表示するカード。
 * 完了・アーカイブ済みの場合は半透明で表示される。
 */
export function TaskCard({ task }: { task: Task }) {
  const isDone = task.status_id === 3 || task.status_id === 4
  const isOverdue = task.due_date && isPast(new Date(task.due_date)) && !isDone

  return (
    <article
      className={clsx(
        'bg-white rounded-lg border p-4 hover:shadow-md transition-shadow',
        isDone && 'opacity-50'
      )}
    >
      <Link to={`/tasks/${task.id}`} className="block">
        <div className="flex items-start justify-between gap-2">
          <h3 className="font-medium text-gray-900 line-clamp-2">{task.title}</h3>
          <span
            className={clsx(
              'shrink-0 text-xs px-2 py-0.5 rounded-full',
              PRIORITY_COLOR[task.priority]
            )}
          >
            {PRIORITY_LABEL[task.priority]}
          </span>
        </div>
        {task.description && (
          <p className="mt-1 text-sm text-gray-500 line-clamp-2">{task.description}</p>
        )}
        <div className="mt-3 flex items-center gap-3 text-xs text-gray-500">
          <span className="px-2 py-0.5 rounded-full bg-gray-100">
            {STATUS_LABEL[task.status_id]}
          </span>
          {task.due_date && (
            <span className={clsx(isOverdue && 'text-red-600 font-medium')}>
              期限: {format(new Date(task.due_date), 'yyyy/MM/dd', { locale: ja })}
            </span>
          )}
        </div>
      </Link>
    </article>
  )
}
