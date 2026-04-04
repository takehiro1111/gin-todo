# Gin Todo Infrastructure (AWS CDK)

## Architecture

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

## Stack 構成

依存方向: `SSM` / `Network` → `Data` → `Compute` → `CDN`

| Stack | CloudFormation 名 | リソース | 変更頻度 |
|-------|-------------------|---------|---------|
| SsmStack | `GinTodoSsmStack` | SSM Parameter Store (DB接続情報, JWT鍵) | 低 |
| NetworkStack | `GinTodoNetworkStack` | VPC, Public/Private Subnet (2AZ), NAT Gateway | 低 |
| DataStack | `GinTodoDataStack` | RDS PostgreSQL 16.4, S3, SES EmailIdentity | 低 |
| ComputeStack | `GinTodoComputeStack` | ALB + ECS Fargate (L3), ECR, Auto Scaling, CloudWatch Alarms (6), SNS | 高 |
| CdnStack | `GinTodoCdnStack` | CloudFront, Route53 A/AAAA, ACM Certificate | 中 |

## プロジェクト構成

```
infra/cdk/
├── bin/cdk.ts                      # エントリポイント — Stack 間の依存を配線
├── lib/
│   ├── constructs/
│   │   └── ssm-parameter.ts        # 既存 SSM Stack (デプロイ済み)
│   ├── stacks/
│   │   ├── network-stack.ts        # VPC + Subnet
│   │   ├── data-stack.ts           # RDS + S3 + SES
│   │   ├── compute-stack.ts        # ECS (L3) + ECR + CloudWatch Alarms
│   │   └── cdn-stack.ts            # CloudFront + Route53 + ACM
│   └── modules/
│       ├── vpc/index.ts
│       ├── rds/index.ts
│       ├── ecs/index.ts            # ApplicationLoadBalancedFargateService (L3) + ECR + Auto Scaling
│       └── cdn/index.ts            # CloudFront + S3 OAC + LoadBalancerV2Origin + Route53
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
| `ApplicationLoadBalancedFargateService` | L3 | ALB + ECS Fargate を統合 |
| `Credentials.fromGeneratedSecret()` | L2 | RDS パスワードを Secrets Manager で自動管理 |
| `LoadBalancerV2Origin` | L2 | CloudFront → ALB の型安全な接続 |
| `S3BucketOrigin.withOriginAccessControl()` | L2 | CloudFront OAC による S3 アクセス |
| `ecr.Repository` | L2 | コンテナイメージ管理 |
| `ses.EmailIdentity` | L2 | SES ドメイン検証 |
| `route53.ARecord` / `AaaaRecord` | L2 | DNS レコード |
| `acm.Certificate` | L2 | SSL 証明書 (DNS 検証) |
| `cloudwatch.Alarm` | L2 | 監視アラーム |
| `autoScaleTaskCount` | L2 | ECS Auto Scaling |

L1 (`Cfn*`) は L2/L3 で対応できない場合のみ使用する。

### Stack と Module の分離

- **Stack** (`lib/stacks/`): Module を組み合わせてデプロイ単位を構成。Module 間の依存を配線する
- **Module** (`lib/modules/`): 単一責務の Construct。Props で依存を受け取り、public プロパティで公開する

### Stack 分割の判断基準

1. **ライフサイクル**: 変更頻度が近いリソースを同じ Stack に
2. **依存方向**: 循環依存を作らない (上→下の一方向)
3. **破壊の影響範囲**: ステートフルリソース (RDS, S3, ECR) は Compute と分離

### セキュリティ

- RDS: パブリックアクセス無効、Private Subnet 配置、ストレージ暗号化有効
- ECS: パブリック IP なし、ALB 経由のみ
- S3: パブリックアクセス全ブロック、CloudFront OAC 経由のみ
- Security Group: 最小権限 (ALB→ECS:8080, ECS→RDS:5432 のみ)
- IAM Task Role: SSM 読み取り (`/gin-todo/*`) + SES 送信 + Secrets Manager 読み取り
- CloudFront: HTTPS リダイレクト、WebSocket パス (`/api/ws/*`) 対応

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
npx cdk deploy GinTodoNetworkStack
npx cdk deploy GinTodoDataStack
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

# 3. 差分確認
npx cdk diff

# 4. デプロイ (依存順に自動解決される)
npx cdk deploy --all

# 5. ECS コンテナイメージの差し替え
#    初回は amazon/amazon-ecs-sample がデプロイされる
#    CI/CD パイプラインで ECR イメージに差し替える
```

## 環境変数・シークレット

| ファイル | Git 管理 | 内容 |
|---------|---------|------|
| `env/env.ts` | する | パブリックなパラメータ (リージョン, ドメイン, ECS スペック等) |
| `env/secret.ts` | しない (.gitignore) | パスワード, JWT キー, RDS ホスト等 |

RDS パスワードは `Credentials.fromGeneratedSecret()` で Secrets Manager が自動管理し、
ECS に `DB_SECRET_ARN` 環境変数として注入する。

| SSM パラメータ | 用途 |
|---------------|------|
| `/gin-todo/db/postgres-user` | DB ユーザー名 |
| `/gin-todo/db/postgres-password` | DB パスワード |
| `/gin-todo/db/postgres-db-name` | DB 名 |
| `/gin-todo/db/postgres-db-host` | DB ホスト |
| `/gin-todo/db/postgres-db-port` | DB ポート |
| `/gin-todo/db/postgres-ssl-mode` | SSL モード |
| `/gin-todo/db/jwt-secret-key` | JWT 署名キー |
