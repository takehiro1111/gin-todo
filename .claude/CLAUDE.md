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
│   ├── api/             # API通信層 (fetch/axios wrapper)
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
pnpm install         # 依存関係インストール
pnpm dev             # 開発サーバー起動 (http://localhost:3000)
pnpm build           # プロダクションビルド
pnpm preview         # ビルド結果のプレビュー
pnpm lint            # ESLint
pnpm typecheck       # TypeScript チェック
```

## API Base URL

```
Development: http://localhost:8080
```

Vite の proxy 設定で `/api` → `http://localhost:8080/api` にプロキシする。

## Authentication Flow

1. `POST /api/auth/login` → Access Token (レスポンスボディ) + Refresh Token (HttpOnly Cookie)
2. 以降のリクエストに `Authorization: Bearer <token>` ヘッダーを付与
3. トークン期限切れ時は `POST /api/auth/refresh` で更新
4. タスク操作前に `GET /api/auth/csrf-token` で CSRF トークン取得し `X-CSRF-Token` ヘッダーに付与

## Key Rules

- 詳細なコーディング規約は `.claude/rules/` 参照
- API エンドポイント一覧は `.claude/rules/api-endpoints.md` 参照
