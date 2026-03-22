---
paths:
  - "frontend/src/components/**/*.tsx"
  - "frontend/src/pages/**/*.tsx"
---

# Frontend Component Rules

## コンポーネント基本方針

- 関数コンポーネント + hooks のみ使用 (クラスコンポーネント禁止)
- `export default` ではなく named export を使う
- 1ファイル 1コンポーネント
- Props は `interface Props` で定義する

```typescript
// Good
interface Props {
  title: string
  onDelete: (id: number) => void
}

export function TaskCard({ title, onDelete }: Props) {
  return <div>{title}</div>
}

// Bad
export default function TaskCard(props: any) { ... }
```

## ディレクトリ構成

```
components/
├── ui/          # Button, Input, Modal 等の汎用UI
├── layout/      # Header, Sidebar, Layout 等
├── task/        # TaskCard, TaskForm, TaskList 等
└── auth/        # LoginForm, RegisterForm 等

pages/
├── LoginPage.tsx
├── RegisterPage.tsx
├── DashboardPage.tsx
├── TaskDetailPage.tsx
└── AdminPage.tsx
```

## ファイル命名

- コンポーネント: PascalCase (`TaskCard.tsx`)
- hooks: camelCase with `use` prefix (`useAuth.ts`)
- 型定義: camelCase (`task.ts`)
- API: camelCase (`taskApi.ts`)

## State 管理方針

- グローバル状態 → Context API (`contexts/`)
- ローカル状態 → `useState`
- 副作用・データ取得 → カスタムフック (`hooks/`)
- Redux 等の外部状態管理ライブラリは使わない

## Context の使い方

```typescript
// contexts/AuthContext.tsx
interface AuthContextValue {
  user: User | null
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => Promise<void>
}

export const AuthContext = createContext<AuthContextValue | null>(null)

export function useAuth() {
  const ctx = useContext(AuthContext)
  if (!ctx) throw new Error('useAuth must be used within AuthProvider')
  return ctx
}
```

## Error / Loading の扱い

- API コールは try/catch で囲む
- ローディング中は `isLoading` フラグで表示制御
- エラーはユーザーに分かりやすいメッセージを表示 (生のエラーメッセージをそのまま表示しない)
