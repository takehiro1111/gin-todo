import clsx from 'clsx'
import type { ButtonHTMLAttributes } from 'react'

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'danger'
  isLoading?: boolean
}

/**
 * 汎用ボタンコンポーネント。
 */
export function Button({
  variant = 'primary',
  isLoading = false,
  className,
  children,
  disabled,
  ...props
}: ButtonProps) {
  return (
    <button
      {...props}
      disabled={disabled || isLoading}
      className={clsx(
        'px-4 py-2 rounded font-medium transition-colors focus:outline-none focus:ring-2 focus:ring-offset-2',
        {
          'bg-blue-500 text-white hover:bg-blue-600 focus:ring-blue-500': variant === 'primary',
          'bg-gray-200 text-gray-800 hover:bg-gray-300 focus:ring-gray-400':
            variant === 'secondary',
          'bg-red-500 text-white hover:bg-red-600 focus:ring-red-500': variant === 'danger',
          'opacity-50 cursor-not-allowed': disabled || isLoading,
        },
        className
      )}
    >
      {isLoading ? (
        <span className="flex items-center gap-2">
          <span className="animate-spin rounded-full h-4 w-4 border-b-2 border-current" />
          処理中...
        </span>
      ) : (
        children
      )}
    </button>
  )
}
