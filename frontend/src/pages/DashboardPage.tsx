import { Layout } from '@/components/layout/Layout'
import { TaskForm } from '@/components/task/TaskForm'
import { TaskList } from '@/components/task/TaskList'
import { Button } from '@/components/ui/Button'
import { Modal } from '@/components/ui/Modal'
import { useCreateTask } from '@/hooks/useCreateTask'
import { useTasks } from '@/hooks/useTasks'
import type { CreateTaskBody } from '@/types/task'
import { useState } from 'react'

/**
 * タスク一覧ダッシュボードページ。
 */
export function DashboardPage() {
  const { data: tasks, isLoading } = useTasks()
  const createTask = useCreateTask()
  const [isModalOpen, setIsModalOpen] = useState(false)

  const handleCreate = async (data: CreateTaskBody) => {
    await createTask.mutateAsync(data)
    setIsModalOpen(false)
  }

  return (
    <Layout>
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-2xl font-bold">タスク一覧</h1>
        <Button onClick={() => setIsModalOpen(true)}>+ 新規作成</Button>
      </div>
      <TaskList tasks={tasks} isLoading={isLoading} />
      <Modal isOpen={isModalOpen} onClose={() => setIsModalOpen(false)} title="タスクを作成">
        <TaskForm onSubmit={handleCreate} isLoading={createTask.isPending} submitLabel="作成" />
      </Modal>
    </Layout>
  )
}
