import { adminApi } from '@/api/adminApi'
import { Layout } from '@/components/layout/Layout'
import { useQuery } from '@tanstack/react-query'
import { format } from 'date-fns'
import { ja } from 'date-fns/locale'
import Skeleton from 'react-loading-skeleton'
import 'react-loading-skeleton/dist/skeleton.css'

/**
 * 管理者向けページ。全ユーザーと全タスクを表示する。
 */
export function AdminPage() {
  const { data: users, isLoading: usersLoading } = useQuery({
    queryKey: ['admin', 'users'],
    queryFn: adminApi.getUsers,
  })

  const { data: tasks, isLoading: tasksLoading } = useQuery({
    queryKey: ['admin', 'tasks'],
    queryFn: adminApi.getTasks,
  })

  return (
    <Layout>
      <h1 className="text-2xl font-bold mb-6">管理画面</h1>

      <section className="mb-8">
        <h2 className="text-xl font-semibold mb-4">ユーザー一覧</h2>
        {usersLoading ? (
          <Skeleton count={3} height={40} />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm border-collapse">
              <thead>
                <tr className="bg-gray-100">
                  <th className="px-4 py-2 text-left border">ID</th>
                  <th className="px-4 py-2 text-left border">名前</th>
                  <th className="px-4 py-2 text-left border">メール</th>
                  <th className="px-4 py-2 text-left border">ロール</th>
                  <th className="px-4 py-2 text-left border">登録日</th>
                </tr>
              </thead>
              <tbody>
                {users?.map((user) => (
                  <tr key={user.id} className="hover:bg-gray-50">
                    <td className="px-4 py-2 border">{user.id}</td>
                    <td className="px-4 py-2 border">{user.name}</td>
                    <td className="px-4 py-2 border">{user.email}</td>
                    <td className="px-4 py-2 border">{user.role_id === 1 ? 'admin' : 'writer'}</td>
                    <td className="px-4 py-2 border">
                      {format(new Date(user.created_at), 'yyyy/MM/dd', { locale: ja })}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>

      <section>
        <h2 className="text-xl font-semibold mb-4">タスク一覧 (全ユーザー)</h2>
        {tasksLoading ? (
          <Skeleton count={5} height={40} />
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full text-sm border-collapse">
              <thead>
                <tr className="bg-gray-100">
                  <th className="px-4 py-2 text-left border">ID</th>
                  <th className="px-4 py-2 text-left border">タイトル</th>
                  <th className="px-4 py-2 text-left border">ユーザーID</th>
                  <th className="px-4 py-2 text-left border">ステータス</th>
                  <th className="px-4 py-2 text-left border">優先度</th>
                </tr>
              </thead>
              <tbody>
                {tasks?.map((task) => (
                  <tr key={task.id} className="hover:bg-gray-50">
                    <td className="px-4 py-2 border">{task.id}</td>
                    <td className="px-4 py-2 border">{task.title}</td>
                    <td className="px-4 py-2 border">{task.user_id}</td>
                    <td className="px-4 py-2 border">{task.status_id}</td>
                    <td className="px-4 py-2 border">{task.priority}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </section>
    </Layout>
  )
}
