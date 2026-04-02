import { Header } from '@/components/layout/Header'
import type { ReactNode } from 'react'

/**
 * 認証済みページの共通レイアウト。
 */
export function Layout({ children }: { children: ReactNode }) {
  return (
    <div className="min-h-screen bg-gray-50">
      <Header />
      <main className="max-w-6xl mx-auto px-4 py-6">{children}</main>
    </div>
  )
}
