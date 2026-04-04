#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { SsmLocalStack } from '@/lib/constructs/ssm-local.js';
import { NetworkStack } from '@/lib/stacks/network-stack.js';
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

// Network: VPC + Subnets + NAT Gateway
const network = new NetworkStack(app, 'GinTodoNetworkStack', { env });

// Data: RDS PostgreSQL + S3 (フロントエンド) + SES
const data = new DataStack(app, 'GinTodoDataStack', {
  env,
  vpcModule: network.vpcModule,
});

// 本番用 SSM パラメータ (RDS エンドポイントを動的に設定)
new SsmStack(app, 'GinTodoSsmProductionStack', {
  env,
  rdsInstance: data.rdsModule.instance,
});

// Compute: ALB (internal) + ECS Fargate
const compute = new ComputeStack(app, 'GinTodoComputeStack', {
  env,
  vpcModule: network.vpcModule,
  rdsModule: data.rdsModule,
  sesIdentityArn: data.sesIdentityArn,
});

// CDN: CloudFront + S3 Origin + VPC Origin (ALB)
new CdnStack(app, 'GinTodoCdnStack', {
  env,
  frontendBucket: data.frontendBucket,
  alb: compute.ecsModule.fargateService.loadBalancer,
});
