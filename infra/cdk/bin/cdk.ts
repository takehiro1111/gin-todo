#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { SsmLocalStack } from '@/lib/constructs/ssm-local.js';
import { DnsStack } from '@/lib/stacks/dns-stack.js';
import { NetworkStack } from '@/lib/stacks/network-stack.js';
import { SecurityStack } from '@/lib/stacks/security-stack.js';
import { DataStack } from '@/lib/stacks/data-stack.js';
import { SsmStack } from '@/lib/stacks/ssm-stack.js';
import { ComputeStack } from '@/lib/stacks/compute-stack.js';
import { CdnStack } from '@/lib/stacks/cdn-stack.js';
import { ENV } from '@/env/env.js';

const app = new cdk.App();

const env = {
  account: ENV.accountID,
  region: ENV.defaultRegion,
};

// ローカル開発用 SSM パラメータ (デプロイ済み)
new SsmLocalStack(app, 'GinTodoSsmStack', { env });

// DNS: Route53 Public Hosted Zone
const dns = new DnsStack(app, 'GinTodoDnsStack', { env });

// Network: VPC + Subnets + NAT Gateway
const network = new NetworkStack(app, 'GinTodoNetworkStack', { env });

// Security: ALB SG + ECS SG + RDS SG + ルール
const security = new SecurityStack(app, 'GinTodoSecurityStack', {
  env,
  vpcModule: network.vpcModule,
});

// Data: RDS PostgreSQL + S3 (静的アセット) + SES
const data = new DataStack(app, 'GinTodoDataStack', {
  env,
  vpcModule: network.vpcModule,
  rdsSecurityGroup: security.securityModule.rdsSecurityGroup,
});

// 本番用 SSM パラメータ (RDS エンドポイントを動的に設定)
new SsmStack(app, 'GinTodoSsmProductionStack', {
  env,
  rdsInstance: data.rdsModule.instance,
});

// SES Identity ARN を文字列で構築 (循環依存を防止)
const sesIdentityArn = `arn:aws:ses:${ENV.defaultRegion}:${ENV.accountID}:identity/${ENV.appDomain}`;

// Compute: ALB (internal) + ECS Fargate (Frontend + Backend)
const compute = new ComputeStack(app, 'GinTodoComputeStack', {
  env,
  vpcModule: network.vpcModule,
  rdsModule: data.rdsModule,
  albSecurityGroup: security.securityModule.albSecurityGroup,
  ecsSecurityGroup: security.securityModule.ecsSecurityGroup,
  sesIdentityArn,
});

// CDN: CloudFront + S3 (静的アセット) + VPC Origin (ALB) + Route53 レコード
new CdnStack(app, 'GinTodoCdnStack', {
  env,
  alb: compute.ecsModule.fargateService.loadBalancer,
  hostedZone: dns.hostedZone,
});
