# Gin Todo Frontend

React + TypeScript によるタスク管理 SPA。

## 技術スタック

- React 18 + TypeScript + Vite
- React Router v6 / Context API (状態管理)
- Tailwind CSS + clsx (スタイリング)
- axios + TanStack Query (API通信・キャッシュ)
- react-hook-form + zod (フォーム・バリデーション)
- Biome (Linter + Formatter) / husky + lint-staged
- Vitest + Testing Library (テスト)

## セットアップ

```bash
cd frontend

# 依存関係インストール
pnpm install

# 開発サーバー起動 (http://localhost:3000)
pnpm dev

# プロダクションビルド
pnpm build
```

## コマンド

| コマンド | 説明 |
|---------|------|
| `pnpm dev` | 開発サーバー起動 |
| `pnpm build` | プロダクションビルド |
| `pnpm typecheck` | TypeScript 型チェック |
| `pnpm test` | テスト (watch モード) |
| `pnpm test --run` | テスト (1回実行) |
| `pnpm biome check src/` | Lint + Format チェック |
| `pnpm biome check --write src/` | Lint + Format 自動修正 |

## 注意事項

- 開発サーバーは **必ず `http://localhost:3000`** で起動すること
  - CORS と Secure Cookie の localhost 例外に依存しているため
  - `127.0.0.1` や別ホスト名は不可
