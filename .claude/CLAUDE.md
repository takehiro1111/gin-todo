# Gin Todo - Frontend Development Guide

## Project Overview
Go + Gin バックエンドに対応する React フロントエンドを実装するプロジェクト。
バックエンドは実装済み。フロントエンドはこれから実装する。

## Tech Stack

### Frontend
- React 18 + TypeScript
- Vite (ビルドツール)
- React Router v6 (ルーティング)
- Context API (状態管理 — Redux等は使わない)
- Tailwind CSS (スタイリング — style属性・CSS Modules不使用)
- clsx (条件付きクラス名)
- react-hook-form + zod (フォーム管理 + バリデーション)
- axios (HTTP クライアント)
- TanStack Query (サーバー状態キャッシュ・データ取得)
- date-fns (日付操作・フォーマット)
- react-hot-toast (Toast通知)
- react-loading-skeleton (スケルトンローディング)
- nprogress (ページ遷移のプログレスバー)
- Biome (Linter + Formatter)
- husky + lint-staged (コミット前自動チェック)
- Vitest + Testing Library (単体・結合テスト)

### Backend (実装済み)
- Go + Gin、ポート 8080
- JWT 認証 (Access Token 15分 / Refresh Token 7日)
- CSRF トークン保護 (タスク操作に必要)
- WebSocket (チャット: `/api/ws/chat`)

## Directory Structure (frontend/)

```
frontend/
├── src/
│   ├── components/      # 再利用可能なUIコンポーネント
│   ├── pages/           # ページコンポーネント (Router対応)
│   ├── contexts/        # Context API (AuthContext, TaskContext等)
│   ├── hooks/           # カスタムフック
│   ├── api/             # API通信層 (axios)
│   ├── types/           # TypeScript型定義
│   └── utils/           # ユーティリティ関数
├── public/
├── index.html
├── vite.config.ts
└── package.json
```

## Commands

```bash
cd frontend
pnpm install                   # 依存関係インストール
pnpm dev                       # 開発サーバー起動 (http://localhost:3000)
pnpm build                     # プロダクションビルド
pnpm preview                   # ビルド結果のプレビュー
pnpm biome check src/          # lint + format チェック
pnpm biome check --write src/  # 自動修正
pnpm typecheck                 # TypeScript チェック
pnpm test                      # テスト (ウォッチモード)
pnpm test --run                # テスト (1回実行)
```

## API Base URL

```
Development: http://localhost:8080
```

バックエンドは `AllowOrigins: ["http://localhost:3000"]` で CORS を許可済みのためプロキシ不要。
`VITE_API_BASE_URL=http://localhost:8080` を `.env` に設定して axios の `baseURL` に渡す。

## Authentication Flow

1. `POST /api/auth/login` → `{ access_token, refresh_token }` がレスポンスボディで返る (Cookie ではない)
2. 両トークンを `localStorage` に保存する
3. 以降のリクエストに `Authorization: Bearer <access_token>` ヘッダーを付与
4. Access Token 期限切れ (401) → `POST /api/auth/refresh` を呼ぶ。このエンドポイントは `VerifyUser` ミドルウェア配下のため `Authorization: Bearer <refresh_token>` を付与し、ボディに `{ refresh_token }` を送る
5. タスク操作前に `GET /api/auth/csrf-token` で CSRF トークン取得し `X-CSRF-Token` ヘッダーに付与

## Vite 設定

`vite.config.ts` でパスエイリアスを設定する。

```typescript
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import path from 'path'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: { '@': path.resolve(__dirname, './src') },
  },
  server: { port: 3000 },
})
```

`tsconfig.json` にも追加する。

```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": { "@/*": ["src/*"] }
  }
}
```

以降のインポートはすべて `@/` を使う。

```typescript
// Bad
import { useAuth } from '../../../contexts/AuthContext'
// Good
import { useAuth } from '@/contexts/AuthContext'
```

## 環境変数

`.env` ファイルで管理。`VITE_` プレフィックス必須。

```
VITE_API_BASE_URL=http://localhost:8080
```

## Key Rules

- 詳細なコーディング規約は `.claude/rules/` 参照
- API エンドポイント一覧は `.claude/rules/api-endpoints.md` 参照
