import { authApi } from '@/api/authApi'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { zodResolver } from '@hookform/resolvers/zod'
import axios from 'axios'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import toast from 'react-hot-toast'
import { Link } from 'react-router-dom'
import { z } from 'zod'

const schema = z.object({
  email: z.string().email('正しいメールアドレスを入力してください'),
})

type Fields = z.infer<typeof schema>

/**
 * パスワードリセットメール送信ページ。
 */
export function ForgotPasswordPage() {
  const [sent, setSent] = useState(false)
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<Fields>({ resolver: zodResolver(schema) })

  const onSubmit = async (data: Fields) => {
    try {
      await authApi.forgotPassword(data.email)
      setSent(true)
    } catch (error) {
      if (axios.isAxiosError(error)) {
        toast.error(error.response?.data?.message ?? 'メール送信に失敗しました')
      }
    }
  }

  if (sent) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
        <div className="bg-white rounded-lg shadow p-8 w-full max-w-md text-center">
          <h1 className="text-2xl font-bold mb-4">メールを送信しました</h1>
          <p className="text-gray-600 mb-6">
            パスワードリセット用のリンクをメールアドレスに送信しました。
          </p>
          <Link to="/login" className="text-blue-500 hover:underline">
            ログインページへ戻る
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
      <div className="bg-white rounded-lg shadow p-8 w-full max-w-md">
        <h1 className="text-2xl font-bold text-center mb-6">パスワードリセット</h1>
        <p className="text-sm text-gray-600 mb-4">
          登録済みのメールアドレスを入力してください。パスワードリセット用のリンクを送信します。
        </p>
        <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4">
          <Input
            label="メールアドレス"
            type="email"
            error={errors.email?.message}
            {...register('email')}
          />
          <Button type="submit" isLoading={isSubmitting} className="w-full">
            リセットメールを送信
          </Button>
        </form>
        <p className="mt-4 text-center text-sm text-gray-600">
          <Link to="/login" className="text-blue-500 hover:underline">
            ログインページへ戻る
          </Link>
        </p>
      </div>
    </div>
  )
}
