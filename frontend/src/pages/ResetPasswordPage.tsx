import { authApi } from '@/api/authApi'
import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { zodResolver } from '@hookform/resolvers/zod'
import axios from 'axios'
import { useForm } from 'react-hook-form'
import toast from 'react-hot-toast'
import { Link, useSearchParams } from 'react-router-dom'
import { z } from 'zod'

const schema = z
  .object({
    newPassword: z.string().min(8, '8文字以上で入力してください'),
    confirmPassword: z.string(),
  })
  .refine((d) => d.newPassword === d.confirmPassword, {
    message: 'パスワードが一致しません',
    path: ['confirmPassword'],
  })

type Fields = z.infer<typeof schema>

/**
 * パスワードリセット実行ページ。クエリパラメータの token を使う。
 */
export function ResetPasswordPage() {
  const [params] = useSearchParams()
  const resetToken = params.get('token') ?? ''

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting, isSubmitSuccessful },
  } = useForm<Fields>({ resolver: zodResolver(schema) })

  const onSubmit = async (data: Fields) => {
    if (!resetToken) {
      toast.error('リセットトークンが見つかりません')
      return
    }
    try {
      await authApi.resetPassword(resetToken, data.newPassword)
      toast.success('パスワードをリセットしました')
    } catch (error) {
      if (axios.isAxiosError(error)) {
        toast.error(error.response?.data?.message ?? 'リセットに失敗しました')
      }
    }
  }

  if (isSubmitSuccessful) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
        <div className="bg-white rounded-lg shadow p-8 w-full max-w-md text-center">
          <h1 className="text-2xl font-bold mb-4">パスワードをリセットしました</h1>
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
        <h1 className="text-2xl font-bold text-center mb-6">新しいパスワードを設定</h1>
        <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4">
          <Input
            label="新しいパスワード"
            type="password"
            error={errors.newPassword?.message}
            {...register('newPassword')}
          />
          <Input
            label="パスワード確認"
            type="password"
            error={errors.confirmPassword?.message}
            {...register('confirmPassword')}
          />
          <Button type="submit" isLoading={isSubmitting} className="w-full">
            パスワードをリセット
          </Button>
        </form>
      </div>
    </div>
  )
}
