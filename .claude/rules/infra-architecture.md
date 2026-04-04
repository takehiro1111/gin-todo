---
paths:
  - "infra/**/*.ts"
---

# Infrastructure Architecture Rules

## 全体構成

```
Route53 (todo.takehiro1111.com)
  └── CloudFront (CDN) + ACM Certificate
        ├── S3 (フロントエンド静的ファイル) — OAC 経由
        └── ALB (VPC Origin) — LoadBalancerV2Origin
              └── ECS Fargate (バックエンド Go API) ← ECR イメージ
                    ├── RDS for PostgreSQL ← Secrets Manager (DB パスワード)
                    ├── SSM Parameter Store (シークレット)
                    └── SES (パスワードリセットメール)

CloudWatch Alarms → SNS Topic (監視)
ECS Auto Scaling (CPU ターゲット追跡)
```

## CDK プロジェクト構成

```
infra/cdk/
├── bin/cdk.ts                          # App エントリポイント — Stack をインスタンス化・配線
├── lib/
│   ├── constructs/
│   │   └── ssm-parameter.ts            # 既存 SSM Stack (デプロイ済み・変更しない)
│   ├── stacks/
│   │   ├── network-stack.ts
│   │   ├── data-stack.ts               # RDS + S3 + SES
│   │   ├── compute-stack.ts            # ECS (L3) + ECR + CloudWatch Alarms
│   │   └── cdn-stack.ts                # CloudFront + Route53 + ACM
│   └── modules/
│       ├── vpc/index.ts                # VPC + Subnet + NAT Gateway
│       ├── rds/index.ts                # RDS PostgreSQL + Security Group
│       ├── ecs/index.ts                # ApplicationLoadBalancedFargateService (L3) + ECR + Auto Scaling
│       └── cdn/index.ts                # CloudFront + S3 OAC + LoadBalancerV2Origin + Route53
├── env/
│   ├── env.ts                          # パブリック値 (Git 管理)
│   └── secret.ts                       # シークレット (.gitignore)
├── test/cdk.test.ts                    # Vitest テスト
└── vitest.config.ts                    # Vitest 設定 (@/ エイリアス解決)
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
| `ecs_patterns.ApplicationLoadBalancedFargateService` | L3 | ALB + ECS Fargate を統合 |
| `rds.Credentials.fromGeneratedSecret()` | L2 | RDS パスワードを Secrets Manager で自動管理 |
| `origins.LoadBalancerV2Origin` | L2 | CloudFront → ALB の型安全な接続 |
| `origins.S3BucketOrigin.withOriginAccessControl()` | L2 | CloudFront OAC による S3 アクセス |
| `ecr.Repository` | L2 | コンテナイメージ管理 (ライフサイクルルール付き) |
| `ses.EmailIdentity` | L2 | SES ドメイン検証 |
| `route53.ARecord` / `AaaaRecord` | L2 | DNS レコード (CloudFront エイリアス) |
| `acm.Certificate` | L2 | SSL 証明書 (DNS 検証) |
| `cloudwatch.Alarm` | L2 | 監視アラーム (SNS 通知) |
| `autoScaleTaskCount` + `scaleOnCpuUtilization` | L2 | ECS Auto Scaling |
| `rds.DatabaseInstance` | L2 | RDS インスタンス |
| `ec2.Vpc` | L2 | VPC (サブネット・NAT Gateway 自動構成) |

## Stack 分割方針

**3つの軸** で分割を判断する:

1. **ライフサイクル (変更頻度)** — 変更頻度が近いリソースを同じ Stack に
2. **依存方向** — 循環依存を作らない。必ず上→下の一方向
3. **破壊の影響範囲** — `cdk destroy` でデータが消えるリソース (RDS, S3) は Compute と分離する

```
SsmStack          ─────────────────────────────┐
                                                ↓
NetworkStack  →  DataStack  →  ComputeStack  →  CdnStack
(ほぼ不変)       (たまに変更)    (頻繁に変更)      (頻繁に変更)
```

| Stack | リソース | 依存先 |
|-------|---------|--------|
| `SsmStack` | SSM Parameter Store | なし |
| `NetworkStack` | VPC, Subnet, NAT Gateway | なし |
| `DataStack` | RDS PostgreSQL, S3, SES EmailIdentity | Network |
| `ComputeStack` | ApplicationLoadBalancedFargateService, ECR, Auto Scaling, CloudWatch Alarms, SNS | Network, Data |
| `CdnStack` | CloudFront, Route53 A/AAAA, ACM Certificate | Compute, Data |

**原則:**
- Stack 間の参照は props で渡す
- Security Group はそれを使うリソースと同じ Stack に入れる (ライフサイクルで判断)
- ステートフルリソース (RDS, S3, ECR) には `removalPolicy: RemovalPolicy.RETAIN` を設定する

## import パス

- `@/` エイリアスを使用 (例: `import { ENV } from '@/env/env.js'`)
- `tsconfig.json` の `baseUrl` + `paths` で TypeScript に認識させる
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
- RDS パスワードは `Credentials.fromGeneratedSecret()` で Secrets Manager 自動管理
- RDS Secret は ECS に `DB_SECRET_ARN` 環境変数として注入
- 複数環境 (dev/staging/prod) が必要になったら `env/` を分割する

## セキュリティ

- RDS: パブリックアクセス無効、VPC 内部のみ、ストレージ暗号化有効
- ECS: パブリック IP なし、ALB 経由のみ
- Security Group: 最小権限 (ALB→ECS:8080, ECS→RDS:5432 のみ)
- S3: パブリックアクセスブロック有効、CloudFront OAC 経由のみ
- IAM Task Role: SSM 読み取り (`/gin-todo/*`) + SES 送信 + Secrets Manager 読み取り (RDS Secret)
- CloudFront: HTTPS リダイレクト、WebSocket パス (`/api/ws/*`) 対応

## テスト

- Vitest でテストを実行する (`pnpm test`)
- `Template.fromStack()` で CloudFormation テンプレートのアサーションテストを書く
- クロススタック循環参照を避けるため、テストでは Module を同一スタック内に配置する

## デプロイ時の注意

- `cdk diff` で差分を必ず確認してからデプロイする
- ステートフルリソース (RDS, S3, ECR) には `removalPolicy: RemovalPolicy.RETAIN` を設定する
- `cdk deploy` と `cdk destroy` は破壊的操作のため、Claude Code の自動許可には入れない (都度確認)
