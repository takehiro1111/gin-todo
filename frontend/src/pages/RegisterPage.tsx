import { Button } from '@/components/ui/Button'
import { Input } from '@/components/ui/Input'
import { useAuth } from '@/contexts/AuthContext'
import { zodResolver } from '@hookform/resolvers/zod'
import axios from 'axios'
import { useForm } from 'react-hook-form'
import toast from 'react-hot-toast'
import { Link, useNavigate } from 'react-router-dom'
import { z } from 'zod'

const registerSchema = z
  .object({
    name: z.string().min(1, '名前を入力してください'),
    email: z.string().email('正しいメールアドレスを入力してください'),
    password: z.string().min(8, '8文字以上で入力してください'),
    confirmPassword: z.string(),
  })
  .refine((d) => d.password === d.confirmPassword, {
    message: 'パスワードが一致しません',
    path: ['confirmPassword'],
  })

type RegisterFields = z.infer<typeof registerSchema>

/**
 * ユーザー登録ページ。
 */
export function RegisterPage() {
  const { register: registerUser } = useAuth()
  const navigate = useNavigate()

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<RegisterFields>({ resolver: zodResolver(registerSchema) })

  const onSubmit = async (data: RegisterFields) => {
    try {
      await registerUser(data.name, data.email, data.password)
      navigate('/')
      toast.success('登録が完了しました')
    } catch (error) {
      if (axios.isAxiosError(error)) {
        toast.error(error.response?.data?.message ?? '登録に失敗しました')
      }
    }
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center px-4">
      <div className="bg-white rounded-lg shadow p-8 w-full max-w-md">
        <h1 className="text-2xl font-bold text-center mb-6">新規登録</h1>
        <form onSubmit={handleSubmit(onSubmit)} className="flex flex-col gap-4">
          <Input label="名前" error={errors.name?.message} {...register('name')} />
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
          <Input
            label="パスワード確認"
            type="password"
            error={errors.confirmPassword?.message}
            {...register('confirmPassword')}
          />
          <Button type="submit" isLoading={isSubmitting} className="w-full">
            登録
          </Button>
        </form>
        <p className="mt-4 text-center text-sm text-gray-600">
          アカウントをお持ちの方は{' '}
          <Link to="/login" className="text-blue-500 hover:underline">
            ログイン
          </Link>
        </p>
      </div>
    </div>
  )
}
