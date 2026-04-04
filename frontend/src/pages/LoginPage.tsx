import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useAuth } from '@/contexts/AuthContext'
import { zodResolver } from '@hookform/resolvers/zod'
import axios from 'axios'
import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import toast from 'react-hot-toast'
import { Link, useLocation, useNavigate } from 'react-router-dom'
import { z } from 'zod'

const loginSchema = z.object({
  email: z.string().email('正しいメールアドレスを入力してください'),
  password: z.string().min(1, 'パスワードを入力してください'),
})

type LoginFields = z.infer<typeof loginSchema>

interface LocationState {
  from?: { pathname: string }
}

/**
 * ログインページ。
 */
export function LoginPage() {
  const { login, user } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const from = (location.state as LocationState)?.from?.pathname ?? '/'

  // ログイン済みならリダイレクト
  useEffect(() => {
    if (user) navigate('/', { replace: true })
  }, [user, navigate])

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFields>({ resolver: zodResolver(loginSchema) })

  const onSubmit = async (data: LoginFields) => {
    try {
      await login(data.email, data.password)
      navigate(from, { replace: true })
    } catch (error) {
      if (axios.isAxiosError(error)) {
        toast.error(error.response?.data?.message ?? 'ログインに失敗しました')
      }
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
      <div className="bg-white rounded-lg shadow p-8 w-full max-w-md">
        <h1 className="text-2xl font-bold text-center mb-6">ログイン</h1>
        <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4">
          <Input
            label="メールアドレス"
            type="email"
            error={errors.email?.message}
            {...register('email')}
          />
          <Input
            label="パスワード"
            type="password"
            error={errors.password?.message}
            {...register('password')}
          />
          <Button type="submit" isLoading={isSubmitting} className="w-full">
            ログイン
          </Button>
        </form>
        <div className="mt-4 text-center text-sm text-gray-600 flex flex-col gap-2">
          <Link to="/forgot-password" className="text-blue-500 hover:underline">
            パスワードを忘れた方はこちら
          </Link>
          <div>
            アカウントをお持ちでない方は{' '}
            <Link to="/register" className="text-blue-500 hover:underline">
              新規登録
            </Link>
          </div>
        </div>
      </div>
    </div>
  )
}
