# Gin Todo API

タスク管理API（Go + Gin + React）のフルスタックアプリケーション

## 技術スタック

### Backend(概ね実装済み)

- Go
- Gin Web Framework
- PostgreSQL
- GORM (ORM)
- golang-migrate (DBマイグレーション)
- JWT認証 (Access Token + Refresh Token)
- WebSocket (リアルタイム通知)
- AWS SDK v2 (SSM)
- Swagger / OpenAPI (swaggo)
- Air (ホットリロード)

### Frontend(未実装)

- React
- TypeScript
- Vite
- React Router
- Context API (状態管理)

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
