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
│   │   └── ssm-local.ts            # ローカル開発用 SSM Stack (デプロイ済み)
│   ├── stacks/
│   │   ├── dns-stack.ts            # Route53 Public Hosted Zone
│   │   ├── network-stack.ts        # VPC + Subnet
│   │   ├── data-stack.ts           # RDS + S3 + SES
│   │   ├── ssm-stack.ts            # 本番用 SSM (RDS エンドポイント動的取得)
│   │   ├── compute-stack.ts        # ECS (Frontend + Backend) + CloudWatch Alarms
│   │   └── cdn-stack.ts            # CloudFront + Route53 + ACM
│   └── modules/
│       ├── vpc/index.ts            # VPC + Subnet + NAT Gateway
│       ├── rds/index.ts            # RDS PostgreSQL + Security Group
│       ├── ecs/index.ts            # ALB (internal) + Frontend/Backend ECS + Service Connect
│       └── cdn/index.ts            # CloudFront + S3 OAC + VPC Origin + Route53
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
| SsmLocalStack | `GinTodoSsmStack` | SSM Parameter Store (ローカル開発用) | なし |
| DnsStack | `GinTodoDnsStack` | Route53 Public Hosted Zone | なし |
| NetworkStack | `GinTodoNetworkStack` | VPC, Subnet (2AZ), NAT Gateway | なし |
| DataStack | `GinTodoDataStack` | RDS PostgreSQL 16.4, S3 (静的アセット), SES EmailIdentity | Network |
| SsmStack | `GinTodoSsmProductionStack` | SSM Parameter Store (本番用, RDS エンドポイント動的) | Data |
| ComputeStack | `GinTodoComputeStack` | ALB (internal) + ECS Fargate (Frontend + Backend), ECR x2, Service Connect, Auto Scaling, CloudWatch Alarms, SNS | Network, Data |
| CdnStack | `GinTodoCdnStack` | CloudFront (VPC Origin + S3 OAC), Route53 A/AAAA, ACM Certificate (us-east-1) | Compute, Data, Dns |

## Construct レベルの方針

- **L2/L3 Construct を積極的に使用する**
  - L3: `ApplicationLoadBalancedFargateService` (ALB + Frontend ECS 統合)
  - L2: `origins.VpcOrigin.withApplicationLoadBalancer()` (CloudFront VPC Origin → internal ALB)
  - L2: `S3BucketOrigin.withOriginAccessControl()` (CloudFront OAC)
  - L2: `DnsValidatedCertificate` (ACM 証明書、us-east-1 指定)
  - L2: `Credentials.fromPassword()` (SSM 管理の RDS パスワード)
  - L2: `ecr.Repository` (コンテナイメージ管理)
  - L2: `ses.EmailIdentity` (SES ドメイン検証)
  - L2: `route53.PublicHostedZone` (DNS ゾーン)
  - L2: `route53.ARecord` / `AaaaRecord` (DNS レコード)
  - L2: `cloudwatch.Alarm` (監視アラーム)
  - L2: `autoScaleTaskCount` + `scaleOnCpuUtilization` (ECS Auto Scaling)
  - L2: `ec2.Vpc` + `IpAddresses.cidr()` (VPC CIDR 明示指定)
- **L1 (`Cfn*`) は L2/L3 で対応できない場合のみ使用**

## import パス

- `@/` エイリアスを使用 (例: `import { ENV } from '@/env/env.js'`)
- `tsconfig.json` の `paths` で解決 (`baseUrl` 不要、TS 5.x+)
- `vitest.config.ts` の `resolve.alias` で Vitest 実行時に解決
- 同一ディレクトリ内の参照 (`./secret.js` 等) はそのまま

## 規約

- **新規リソース追加**: `lib/modules/` に Module、`lib/stacks/` に Stack を作成
- **ローカル用 SSM**: `lib/constructs/ssm-local.ts` はデプロイ済みのため変更しない
- **本番用 SSM**: `lib/stacks/ssm-stack.ts` で RDS エンドポイントを動的取得
- **環境固有の値**: `env/env.ts` (パブリック) + `env/secret.ts` (シークレット) に分離
- **命名**: Stack 名は `GinTodo` プレフィックス + リソースカテゴリ
- **テスト**: `Template.fromStack()` でリソースの存在と設定を検証。ARN は `Stack.formatArn()` を使用
- **パッケージ管理**: pnpm を使用 (npm/yarn は使わない)

## シークレット管理

- `env/env.ts`: パブリックなパラメータ (Git 管理)
- `env/secret.ts`: シークレット値 (.gitignore で Git 管理外)
- RDS パスワード: `Credentials.fromPassword()` で SSM 管理のパスワードを使用
- 本番 SSM: RDS エンドポイントを `rdsInstance.dbInstanceEndpointAddress` で動的設定
- SSL モード: ローカル `disable` / 本番 `require`
- 本番値は絶対にコミットしない

## CDK 新規リソース追加時の手順

1. `lib/modules/` に Module ファイルを作成
2. `lib/stacks/` に Stack ファイルを作成
3. `bin/cdk.ts` でインスタンス化・配線
4. `npx cdk synth` で CloudFormation テンプレートを確認
5. `npx cdk diff` で既存環境との差分を確認
6. テストを `test/cdk.test.ts` に追加
7. `npx cdk deploy` でデプロイ
