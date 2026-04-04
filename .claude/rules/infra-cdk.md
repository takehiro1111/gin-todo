---
paths:
  - "infra/**/*.ts"
  - "infra/**/*.json"
---

# Infrastructure (AWS CDK) Rules

## 構成

```
infra/cdk/
├── bin/cdk.ts                      # エントリポイント — Stack 間の依存を配線
├── lib/
│   ├── constructs/
│   │   └── ssm-parameter.ts        # 既存 SSM Stack (デプロイ済み)
│   ├── stacks/                     # Stack 定義
│   │   ├── network-stack.ts        # VPC + Subnet
│   │   ├── data-stack.ts           # RDS + S3 + SES
│   │   ├── compute-stack.ts        # ECS Fargate (L3) + ECR + CloudWatch Alarms
│   │   └── cdn-stack.ts            # CloudFront + Route53 + ACM
│   └── modules/                    # 再利用可能な Construct モジュール
│       ├── vpc/index.ts            # VPC + Subnet + NAT Gateway
│       ├── rds/index.ts            # RDS PostgreSQL + Security Group
│       ├── ecs/index.ts            # ApplicationLoadBalancedFargateService (L3) + ECR + Auto Scaling
│       └── cdn/index.ts            # CloudFront + S3 OAC + LoadBalancerV2Origin + Route53
├── env/
│   ├── env.ts                      # パブリック値 (Git 管理)
│   └── secret.ts                   # シークレット (.gitignore)
├── test/cdk.test.ts                # Vitest テスト
├── vitest.config.ts                # Vitest 設定 (@/ エイリアス解決)
├── cdk.json                        # CDK 設定 + Feature Flags
└── package.json                    # pnpm 管理
```

## コマンド (`infra/cdk/` で実行)

```bash
pnpm install                 # 依存インストール
npx cdk synth                # CloudFormation テンプレート生成 (デプロイ前の確認)
npx cdk diff                 # 現在のスタックとの差分表示
npx cdk deploy --all         # 全スタックデプロイ
npx cdk destroy --all        # スタック削除
pnpm test                    # Vitest テスト
```

## デプロイ先

- リージョン: `ap-northeast-1` (東京)
- ドメイン: `todo.takehiro1111.com`

## Stack 一覧

| Stack | CloudFormation 名 | リソース | 依存先 |
|-------|-------------------|---------|--------|
| SsmStack | `GinTodoSsmStack` | SSM Parameter Store | なし |
| NetworkStack | `GinTodoNetworkStack` | VPC, Subnet (2AZ), NAT Gateway | なし |
| DataStack | `GinTodoDataStack` | RDS PostgreSQL 16.4, S3, SES EmailIdentity | Network |
| ComputeStack | `GinTodoComputeStack` | ALB + ECS Fargate (L3), ECR, Auto Scaling, CloudWatch Alarms, SNS | Network, Data |
| CdnStack | `GinTodoCdnStack` | CloudFront, Route53 A/AAAA, ACM Certificate | Compute, Data |

## Construct レベルの方針

- **L2/L3 Construct を積極的に使用する**
  - L3: `ApplicationLoadBalancedFargateService` (ALB + ECS Fargate 統合)
  - L2: `Credentials.fromGeneratedSecret()` (RDS パスワードの Secrets Manager 自動管理)
  - L2: `LoadBalancerV2Origin` (CloudFront → ALB の型安全な接続)
  - L2: `S3BucketOrigin.withOriginAccessControl()` (CloudFront OAC)
  - L2: `ecr.Repository` (コンテナイメージ管理)
  - L2: `ses.EmailIdentity` (SES ドメイン検証)
  - L2: `route53.ARecord` / `AaaaRecord` (DNS レコード)
  - L2: `acm.Certificate` (SSL 証明書、DNS 検証)
  - L2: `cloudwatch.Alarm` (監視アラーム)
  - L2: `autoScaleTaskCount` + `scaleOnCpuUtilization` (ECS Auto Scaling)
- **L1 (`Cfn*`) は L2/L3 で対応できない場合のみ使用**

## import パス

- `@/` エイリアスを使用 (例: `import { ENV } from '@/env/env.js'`)
- `tsconfig.json` の `paths` + `vitest.config.ts` の `resolve.alias` で解決
- 同一ディレクトリ内の参照 (`./secret.js` 等) はそのまま

## 規約

- **新規リソース追加**: `lib/modules/` に Module、`lib/stacks/` に Stack を作成
- **既存 SSM Stack**: `lib/constructs/ssm-parameter.ts` はデプロイ済みのため変更しない
- **環境固有の値**: `env/env.ts` (パブリック) + `env/secret.ts` (シークレット) に分離
- **命名**: Stack 名は `GinTodo` プレフィックス + リソースカテゴリ
- **テスト**: `Template.fromStack()` でリソースの存在と設定を検証する
- **パッケージ管理**: pnpm を使用 (npm/yarn は使わない)

## シークレット管理

- `env/env.ts`: パブリックなパラメータ (Git 管理)
- `env/secret.ts`: シークレット値 (.gitignore で Git 管理外)
- RDS パスワード: `Credentials.fromGeneratedSecret()` で Secrets Manager 自動管理
- RDS Secret は ECS に `DB_SECRET_ARN` 環境変数として注入
- 本番値は絶対にコミットしない

## CDK 新規リソース追加時の手順

1. `lib/modules/` に Module ファイルを作成
2. `lib/stacks/` に Stack ファイルを作成
3. `bin/cdk.ts` でインスタンス化・配線
4. `npx cdk synth` で CloudFormation テンプレートを確認
5. `npx cdk diff` で既存環境との差分を確認
6. テストを `test/cdk.test.ts` に追加
7. `npx cdk deploy` でデプロイ
