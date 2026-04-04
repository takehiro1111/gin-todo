# Gin Todo Infrastructure (AWS CDK)

## Architecture

```
Route53 (todo.takehiro1111.com)
  └── CloudFront (CDN) + ACM Certificate (us-east-1)
        ├── /assets/*  → S3 (静的アセット) — OAC 経由
        └── /* (default), /api/*, /api/ws/*
              └── ALB (internal, VPC Origin) — PrivateLink 経由
                    ├── /* (default)  → Frontend ECS (Node serve, SPA)
                    └── /api/*        → Backend ECS (Go API)
                                        ├── RDS for PostgreSQL ← SSM (パスワード)
                                        ├── SSM Parameter Store (DB接続情報, JWT鍵)
                                        └── SES (パスワードリセットメール)

ECS Service Connect (gin-todo.local 名前空間)
  Frontend → http://backend:8080 (Envoy サイドカー経由)

CloudWatch Alarms → SNS Topic (監視)
ECS Auto Scaling (CPU ターゲット追跡)
```

### 通信フロー

```
ブラウザ
  │
  ▼
CloudFront (HTTPS)
  │
  ├─ /assets/*     ─→ S3 OAC (静的アセット)
  │
  └─ それ以外       ─→ VPC Origin (PrivateLink)
                        │
                        ▼
                    ALB (internal, Port 80)
                        │
                        ├─ /* (default) ─→ Frontend ECS :3000 (SPA)
                        │
                        └─ /api/*       ─→ Backend ECS :8080 (Go API)
                                              │
                                              ├─→ RDS PostgreSQL :5432
                                              ├─→ SSM Parameter Store
                                              └─→ SES v2

Frontend ECS ──(Service Connect)──→ Backend ECS :8080
                http://backend:8080
```

### Security Group

```
┌─────────────────────────────────────────────┐
│ ECS SG (Frontend/Backend 共有)              │
│  Inbound:                                   │
│    - ALB SG → 全ポート (L3 自動設定)        │
│    - Self   → 全ポート (コンテナ間通信)     │
│  Outbound:                                  │
│    - 全許可 (NAT Gateway 経由で AWS API)    │
└─────────────────────────────────────────────┘
        │ :5432
        ▼
┌─────────────────────────────────────────────┐
│ RDS SG                                      │
│  Inbound:  ECS SG → :5432 のみ             │
│  Outbound: 全閉じ (不要)                    │
└─────────────────────────────────────────────┘

ALB SG:
  Inbound:  VPC Origin SG (CloudFront デプロイ後に自動作成)
  Outbound: ECS SG (L3 自動設定)
```

### ネットワーク

```
VPC 10.0.0.0/16
  ├── Public Subnet  10.0.0.0/24 (AZ-a) — NAT Gateway
  ├── Public Subnet  10.0.1.0/24 (AZ-c)
  ├── Private Subnet 10.0.2.0/24 (AZ-a) — ECS, RDS
  └── Private Subnet 10.0.3.0/24 (AZ-c) — ECS, RDS
```

## Stack 構成

依存方向: `SsmLocal` / `Dns` / `Network` → `Data` → `SsmProduction` / `Compute` → `CDN`

| Stack | CloudFormation 名 | リソース | 変更頻度 |
|-------|-------------------|---------|---------|
| SsmLocalStack | `GinTodoSsmStack` | SSM Parameter Store (ローカル開発用) | 低 |
| DnsStack | `GinTodoDnsStack` | Route53 Public Hosted Zone | 低 |
| NetworkStack | `GinTodoNetworkStack` | VPC, Public/Private Subnet (2AZ), NAT Gateway | 低 |
| DataStack | `GinTodoDataStack` | RDS PostgreSQL 16.4, S3 (静的アセット), SES EmailIdentity | 低 |
| SsmStack | `GinTodoSsmProductionStack` | SSM Parameter Store (本番用, RDS エンドポイント動的設定) | 低 |
| ComputeStack | `GinTodoComputeStack` | ALB (internal) + ECS Fargate (Frontend + Backend), ECR x2, Service Connect, Auto Scaling, CloudWatch Alarms (6), SNS | 高 |
| CdnStack | `GinTodoCdnStack` | CloudFront (VPC Origin + S3 OAC), Route53 A/AAAA, ACM Certificate (us-east-1) | 中 |

## プロジェクト構成

```
infra/cdk/
├── bin/cdk.ts                      # エントリポイント — Stack 間の依存を配線
├── lib/
│   ├── constructs/
│   │   └── ssm-local.ts            # ローカル開発用 SSM Stack (デプロイ済み)
│   ├── stacks/
│   │   ├── dns-stack.ts            # Route53 Public Hosted Zone
│   │   ├── network-stack.ts        # VPC + Subnet
│   │   ├── data-stack.ts           # RDS + S3 + SES
│   │   ├── ssm-stack.ts            # 本番用 SSM (RDS エンドポイント動的取得)
│   │   ├── compute-stack.ts        # ECS (Frontend + Backend) + CloudWatch Alarms
│   │   └── cdn-stack.ts            # CloudFront + Route53 + ACM
│   └── modules/
│       ├── vpc/index.ts
│       ├── rds/index.ts
│       ├── ecs/index.ts            # ALB (internal) + Frontend/Backend ECS + Service Connect
│       └── cdn/index.ts            # CloudFront + S3 OAC + VPC Origin + Route53
├── env/
│   ├── env.ts                      # パブリック値 (Git 管理)
│   └── secret.ts                   # シークレット (.gitignore)
├── test/cdk.test.ts                # Vitest テスト
├── vitest.config.ts                # Vitest 設定 (@/ エイリアス解決)
├── cdk.json
└── package.json                    # pnpm 管理
```

## 設計方針

### L2/L3 Construct の積極使用

| Construct | レベル | 用途 |
|-----------|-------|------|
| `ApplicationLoadBalancedFargateService` | L3 | ALB + Frontend ECS を統合 |
| `VpcOrigin.withApplicationLoadBalancer()` | L2 | CloudFront VPC Origin → internal ALB |
| `S3BucketOrigin.withOriginAccessControl()` | L2 | CloudFront OAC → S3 |
| `DnsValidatedCertificate` | L2 | ACM 証明書 (us-east-1 指定) |
| `ecr.Repository` | L2 | コンテナイメージ管理 |
| `ses.EmailIdentity` | L2 | SES ドメイン検証 |
| `route53.PublicHostedZone` | L2 | DNS ゾーン |
| `route53.ARecord` / `AaaaRecord` | L2 | DNS レコード |
| `cloudwatch.Alarm` | L2 | 監視アラーム |
| `autoScaleTaskCount` | L2 | ECS Auto Scaling |
| `rds.DatabaseInstance` | L2 | RDS インスタンス |
| `ec2.Vpc` + `IpAddresses.cidr()` | L2 | VPC (CIDR 明示指定) |

L1 (`Cfn*`) は L2/L3 で対応できない場合のみ使用する。

### Stack と Module の分離

- **Stack** (`lib/stacks/`): Module を組み合わせてデプロイ単位を構成。Module 間の依存を配線する
- **Module** (`lib/modules/`): 単一責務の Construct。Props で依存を受け取り、public プロパティで公開する

### Stack 分割の判断基準

1. **ライフサイクル**: 変更頻度が近いリソースを同じ Stack に
2. **依存方向**: 循環依存を作らない (上→下の一方向)
3. **破壊の影響範囲**: ステートフルリソース (RDS, S3, ECR) は Compute と分離

### SSM パラメータ管理

| Stack | 用途 | RDS ホスト |
|-------|------|-----------|
| `SsmLocalStack` | ローカル開発 | `localhost` |
| `SsmStack` | 本番 | RDS エンドポイント (動的取得) |

### セキュリティ

- ALB: internal (VPC Origin 経由のみ)
- RDS: パブリックアクセス無効、Private Subnet 配置、ストレージ暗号化有効
- ECS: パブリック IP なし、ALB 経由のみ
- S3: パブリックアクセス全ブロック、CloudFront OAC 経由のみ
- Security Group: ECS SG は Frontend/Backend 共有、RDS SG は ECS からの 5432 のみ
- IAM Task Role: SSM 読み取り (`/gin-todo/*`) + SES 送信 (ドメイン Identity に制限)
- CloudFront: HTTPS リダイレクト、WebSocket パス (`/api/ws/*`) 対応
- SSL: 本番 RDS 接続は `ssl-mode=require`

### 監視

CloudWatch Alarms (6件) → SNS Topic:
- ECS: CPU 使用率 > 80%、メモリ使用率 > 80%
- ALB: 5xx エラー > 10/min
- RDS: CPU 使用率 > 80%、空きストレージ <= 5GB、接続数 > 50

### ステートフルリソースの保護

RDS, S3, ECR には `removalPolicy: RETAIN` を設定。
RDS には `deletionProtection: true` も設定。
`cdk destroy` でもデータは削除されない。

## コマンド

```bash
# infra/cdk/ で実行

# 依存インストール
pnpm install

# CloudFormation テンプレート生成 (デプロイ前の確認)
npx cdk synth

# 現在のスタックとの差分表示
npx cdk diff

# 全スタックデプロイ
npx cdk deploy --all

# 個別スタックデプロイ (依存順)
npx cdk deploy GinTodoDnsStack
npx cdk deploy GinTodoNetworkStack
npx cdk deploy GinTodoDataStack
npx cdk deploy GinTodoSsmProductionStack
npx cdk deploy GinTodoComputeStack
npx cdk deploy GinTodoCdnStack

# テスト
pnpm test

# スタック削除 (ステートフルリソースは RETAIN)
npx cdk destroy --all
```

## 初回デプロイ手順

```bash
# 1. CDK Bootstrap (初回のみ)
npx cdk bootstrap aws://ACCOUNT_ID/ap-northeast-1

# 2. env/secret.ts を作成 (シークレット値を設定)
cp env/secret.example.ts env/secret.ts

# 3. DNS スタックを先にデプロイ
npx cdk deploy GinTodoDnsStack
# → 出力される NS レコードをドメインレジストラに設定

# 4. 差分確認
npx cdk diff

# 5. デプロイ (依存順に自動解決される)
npx cdk deploy --all

# 6. ECS コンテナイメージの差し替え
#    初回は amazon/amazon-ecs-sample がデプロイされる
#    CI/CD パイプラインで ECR イメージに差し替える
```

## 環境変数・シークレット

| ファイル | Git 管理 | 内容 |
|---------|---------|------|
| `env/env.ts` | する | パブリックなパラメータ (リージョン, ドメイン, ECS スペック等) |
| `env/secret.ts` | しない (.gitignore) | パスワード, JWT キー |

| SSM パラメータ (本番) | 用途 |
|----------------------|------|
| `/gin-todo/db/postgres-user` | DB ユーザー名 |
| `/gin-todo/db/postgres-password` | DB パスワード |
| `/gin-todo/db/postgres-db-name` | DB 名 |
| `/gin-todo/db/postgres-db-host` | DB ホスト (RDS エンドポイント, 動的設定) |
| `/gin-todo/db/postgres-db-port` | DB ポート |
| `/gin-todo/db/postgres-ssl-mode` | SSL モード (`require`) |
| `/gin-todo/db/jwt-secret-key` | JWT 署名キー |
