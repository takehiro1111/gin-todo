---
paths:
  - "infra/**/*.ts"
---

# Infrastructure Architecture Rules

## 全体構成

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

## CDK プロジェクト構成

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
└── vitest.config.ts                # Vitest 設定 (@/ エイリアス解決)
```

### Stack と Module の役割分担

| 層 | 役割 | 例 |
|----|------|-----|
| **Stack** (`lib/stacks/`) | Module を組み合わせてデプロイ単位を構成。Module 間の依存を配線する | `NetworkStack` が `VpcModule` を使い、`vpc` を public プロパティで公開 |
| **Module** (`lib/modules/`) | 単一の責務を持つ再利用可能な Construct。Props で設定を受け取る | `RdsModule` は VPC と Security Group を Props で受け取り、RDS インスタンスを作る |

### Module 設計の原則

- **1 Module = 1 責務**: VPC、RDS、ECS など AWS サービス単位で分ける
- **Props で依存を注入**: 他の Module のリソースは Props で受け取る (import しない)
- **public プロパティで公開**: 他 Module が参照するリソースは public readonly で公開
- **Module 内で完結させる**: Security Group やサブネットグループなど付随リソースは Module 内で作る
- **Module 同士は直接参照しない**: Stack が配線役を担う

## Construct レベルの方針

**L2/L3 Construct を積極的に使用する。** L1 (`Cfn*`) は L2/L3 で対応できない場合のみ。

### 現在使用している L2/L3

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

## Stack 分割方針

**3つの軸** で分割を判断する:

1. **ライフサイクル (変更頻度)** — 変更頻度が近いリソースを同じ Stack に
2. **依存方向** — 循環依存を作らない。必ず上→下の一方向
3. **破壊の影響範囲** — `cdk destroy` でデータが消えるリソース (RDS, S3) は Compute と分離する

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

**原則:**
- Stack 間の参照は props で渡す
- Security Group はそれを使うリソースと同じ Stack に入れる (ライフサイクルで判断)
- ステートフルリソース (RDS, S3, ECR) には `removalPolicy: RemovalPolicy.RETAIN` を設定する

## SSM パラメータ管理

| Stack | 用途 | RDS ホスト |
|-------|------|-----------|
| `SsmLocalStack` | ローカル開発 | `localhost` |
| `SsmStack` | 本番 | RDS エンドポイント (動的取得) |

## import パス

- `@/` エイリアスを使用 (例: `import { ENV } from '@/env/env.js'`)
- `tsconfig.json` の `paths` で TypeScript に認識させる (`baseUrl` 不要、TS 5.x+)
- `vitest.config.ts` の `resolve.alias` で Vitest 実行時に解決

## 命名規約

| 対象 | パターン | 例 |
|------|---------|-----|
| Stack クラス名 | `{Category}Stack` | `NetworkStack` |
| Stack ID (bin/cdk.ts) | `GinTodo` + クラス名 | `'GinTodoNetworkStack'` |
| Construct 論理 ID | PascalCase、リソース種別を含める | `'ApiTaskDefinition'` |
| SSM パラメータ名 | `/gin-todo/{category}/{name}` | `/gin-todo/db/postgres-user` |

## 環境管理

- `env/env.ts` にパブリックなパラメータを集約 (Git 管理)
- `env/secret.ts` にシークレット値を分離 (.gitignore で Git 管理外)
- RDS パスワードは `Credentials.fromPassword()` で SSM 管理のパスワードを使用
- 本番 SSM は RDS エンドポイントを動的取得 (`rdsInstance.dbInstanceEndpointAddress`)
- 複数環境 (dev/staging/prod) が必要になったら `env/` を分割する

## セキュリティ

- ALB: internal (VPC Origin 経由のみ)
- RDS: パブリックアクセス無効、Private Subnet 配置、ストレージ暗号化有効
- ECS: パブリック IP なし、ALB 経由のみ
- S3: パブリックアクセス全ブロック、CloudFront OAC 経由のみ
- Security Group: ECS SG は Frontend/Backend 共有、RDS SG は ECS からの 5432 のみ
- IAM Task Role: SSM 読み取り (`/gin-todo/*`) + SES 送信 (ドメイン Identity に制限)
- CloudFront: HTTPS リダイレクト、WebSocket パス (`/api/ws/*`) 対応
- SSL: 本番 RDS 接続は `ssl-mode=require`

## 監視

CloudWatch Alarms (6件) → SNS Topic:
- ECS: CPU 使用率 > 80%、メモリ使用率 > 80%
- ALB: 5xx エラー > 10/min
- RDS: CPU 使用率 > 80%、空きストレージ <= 5GB、接続数 > 50

## テスト

- Vitest でテストを実行する (`pnpm test`)
- `Template.fromStack()` で CloudFormation テンプレートのアサーションテストを書く
- クロススタック循環参照を避けるため、テストでは Module を同一スタック内に配置する
- ARN はハードコードせず `Stack.formatArn()` を使用する

## デプロイ時の注意

- `cdk diff` で差分を必ず確認してからデプロイする
- ステートフルリソース (RDS, S3, ECR) には `removalPolicy: RemovalPolicy.RETAIN` を設定する
- RDS には `deletionProtection: true` も設定する
- `cdk deploy` と `cdk destroy` は破壊的操作のため、Claude Code の自動許可には入れない (都度確認)
