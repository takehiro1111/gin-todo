import { Link } from 'react-router-dom'

/**
 * 存在しないパスへのアクセス時に表示する404ページ。
 */
export function NotFoundPage() {
  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-4">
      <h1 className="text-6xl font-bold text-gray-300">404</h1>
      <p className="text-xl text-gray-500">ページが見つかりません</p>
      <Link to="/" className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600">
        トップへ戻る
      </Link>
    </div>
  )
}
