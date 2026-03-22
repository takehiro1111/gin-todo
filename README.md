# Gin Todo API
- タスク管理API（Go + Gin + React）のフルスタックアプリケーション

## 目的
- Go言語の基礎的な実装力の向上

## 技術スタック

### Backend(実装済み)
- Go
- Gin Web Framework
- PostgreSQL
- GORM (ORM)
- golang-migrate (DBマイグレーション)
- JWT認証 (Access Token + Refresh Token)
- WebSocket (リアルタイム通知)
- AWS SDK v2 (SSM ParameterStore / SES)
- Swagger / OpenAPI (swaggo)
- Air (ホットリロード)

### Frontend(Claude Codeを使用したバイブコーディングで実装済み)
- React
- TypeScript
- Vite
- React Router
- Context API (状態管理)
- Tailwind CSS

### Infra(未実装 / バックエンドに必要なリソースを一部設定)
- AWS
  - VPC
  - Subnet
  - Internet Gateway
  - NAT Instance
  - Route53
  - CloudFront
    - VPC Origin
  - ALB
  - S3
  - ECS
  - RDS for PostgreSQL
  - SSM Parameter Store
  - CDK with TypeScript

### CI/CD(未実装)
- Github Actions
