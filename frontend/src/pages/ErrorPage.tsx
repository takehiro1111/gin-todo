import { useNavigate } from 'react-router-dom'

/**
 * 予期せぬエラー発生時に表示する汎用エラーページ。
 */
export function ErrorPage({ message }: { message?: string }) {
  const navigate = useNavigate()

  return (
    <div className="flex flex-col items-center justify-center min-h-screen gap-4">
      <h1 className="text-4xl font-bold text-red-600">エラーが発生しました</h1>
      <p className="text-gray-500">{message ?? '予期せぬエラーが発生しました'}</p>
      <button
        type="button"
        onClick={() => navigate('/')}
        className="px-4 py-2 bg-blue-500 text-white rounded hover:bg-blue-600"
      >
        トップへ戻る
      </button>
    </div>
  )
}
