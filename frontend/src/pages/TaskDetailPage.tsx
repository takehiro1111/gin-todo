import { Layout } from '@/components/layout/Layout'
import { TaskForm } from '@/components/task/TaskForm'
import { Button } from '@/components/ui/Button'
import { Modal } from '@/components/ui/Modal'
import { useDeleteTask } from '@/hooks/useDeleteTask'
import { useTask } from '@/hooks/useTask'
import { useUpdateTask } from '@/hooks/useUpdateTask'
import { useUpdateTaskStatus } from '@/hooks/useUpdateTaskStatus'
import { format } from 'date-fns'
import { ja } from 'date-fns/locale'
import { useState } from 'react'
import Skeleton from 'react-loading-skeleton'
import { useNavigate, useParams } from 'react-router-dom'
import 'react-loading-skeleton/dist/skeleton.css'
import type { UpdateTaskBody } from '@/types/task'

const STATUS_LABEL: Record<number, string> = {
  1: '未着手',
  2: '進行中',
  3: '完了',
  4: 'アーカイブ',
}

/**
 * タスク詳細ページ。編集・削除・ステータス変更が可能。
 */
export function TaskDetailPage() {
  const { id } = useParams<{ id: string }>()
  const taskId = Number(id)
  const navigate = useNavigate()

  const { data: task, isLoading } = useTask(taskId)
  const updateTask = useUpdateTask(taskId)
  const deleteTask = useDeleteTask()
  const updateStatus = useUpdateTaskStatus()

  const [isEditOpen, setIsEditOpen] = useState(false)

  const handleUpdate = async (data: UpdateTaskBody) => {
    await updateTask.mutateAsync(data)
    setIsEditOpen(false)
  }

  const handleDelete = async () => {
    if (!confirm('このタスクを削除しますか？')) return
    await deleteTask.mutateAsync(taskId)
    navigate('/')
  }

  if (isLoading) {
    return (
      <Layout>
        <Skeleton height={32} width={300} className="mb-4" />
        <Skeleton count={3} />
      </Layout>
    )
  }

  if (!task) {
    return (
      <Layout>
        <p className="text-gray-500">タスクが見つかりません</p>
      </Layout>
    )
  }

  return (
    <Layout>
      <div className="max-w-2xl">
        <div className="flex items-start justify-between mb-4 gap-4">
          <h1 className="text-2xl font-bold">{task.title}</h1>
          <div className="flex gap-2 shrink-0">
            <Button variant="secondary" onClick={() => setIsEditOpen(true)}>
              編集
            </Button>
            <Button variant="danger" isLoading={deleteTask.isPending} onClick={handleDelete}>
              削除
            </Button>
          </div>
        </div>

        {task.description && (
          <p className="text-gray-700 mb-4 whitespace-pre-wrap">{task.description}</p>
        )}

        <div className="grid grid-cols-2 gap-4 mb-6">
          <div>
            <p className="text-sm text-gray-500">ステータス</p>
            <select
              value={task.status_id}
              onChange={(e) =>
                updateStatus.mutate({ id: taskId, statusId: Number(e.target.value) })
              }
              className="mt-1 px-3 py-1.5 border border-gray-300 rounded focus:outline-none focus:ring-2 focus:ring-blue-500"
            >
              {Object.entries(STATUS_LABEL).map(([statusId, label]) => (
                <option key={statusId} value={statusId}>
                  {label}
                </option>
              ))}
            </select>
          </div>
          <div>
            <p className="text-sm text-gray-500">優先度</p>
            <p className="mt-1 font-medium">
              {{ low: '低', medium: '中', high: '高' }[task.priority]}
            </p>
          </div>
          {task.due_date && (
            <div>
              <p className="text-sm text-gray-500">期限</p>
              <p className="mt-1">
                {format(new Date(task.due_date), 'yyyy年MM月dd日', { locale: ja })}
              </p>
            </div>
          )}
          <div>
            <p className="text-sm text-gray-500">作成日</p>
            <p className="mt-1">
              {format(new Date(task.created_at), 'yyyy年MM月dd日', { locale: ja })}
            </p>
          </div>
        </div>

        <Button variant="secondary" onClick={() => navigate('/')}>
          一覧へ戻る
        </Button>
      </div>

      <Modal isOpen={isEditOpen} onClose={() => setIsEditOpen(false)} title="タスクを編集">
        <TaskForm
          defaultValues={{
            title: task.title,
            description: task.description,
            priority: task.priority,
            due_date: task.due_date ?? undefined,
          }}
          onSubmit={handleUpdate}
          isLoading={updateTask.isPending}
          submitLabel="更新"
        />
      </Modal>
    </Layout>
  )
}
