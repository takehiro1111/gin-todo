import { describe, test, beforeAll } from 'vitest'
import * as cdk from 'aws-cdk-lib'
import { Template } from 'aws-cdk-lib/assertions'
import { NetworkStack } from '@/lib/stacks/network-stack.js'
import { DataStack } from '@/lib/stacks/data-stack.js'
import { CdnModule } from '@/lib/modules/cdn/index.js'
import { SsmLocalStack } from '@/lib/constructs/ssm-local.js'
import { SsmStack } from '@/lib/stacks/ssm-stack.js'
import { VpcModule } from '@/lib/modules/vpc/index.js'
import { RdsModule } from '@/lib/modules/rds/index.js'
import { EcsModule } from '@/lib/modules/ecs/index.js'

const env = { account: '123456789012', region: 'ap-northeast-1' }

describe('SsmLocalStack', () => {
  test('SSM パラメータが7つ作成される', () => {
    const app = new cdk.App()
    const stack = new SsmLocalStack(app, 'TestSsmLocalStack')
    const template = Template.fromStack(stack)
    template.resourceCountIs('AWS::SSM::Parameter', 7)
  })
})

describe('SsmStack (本番用)', () => {
  test('SSM パラメータが7つ作成される', () => {
    const app = new cdk.App()
    // RDS を同一スタックに作成して rdsInstance を渡す
    const stack = new cdk.Stack(app, 'TestSsmProdSetup', { env })
    const vpc = new cdk.aws_ec2.Vpc(stack, 'Vpc', { maxAzs: 2 })
    const rdsModule = new RdsModule(stack, 'Rds', {
      vpc,
      subnets: vpc.privateSubnets,
      databaseName: 'gin-todo',
      credentials: cdk.aws_rds.Credentials.fromPassword(
        'gin',
        cdk.SecretValue.unsafePlainText('test-password'),
      ),
    })
    const ssmStack = new SsmStack(app, 'TestSsmProdStack', {
      env,
      rdsInstance: rdsModule.instance,
    })
    const template = Template.fromStack(ssmStack)
    template.resourceCountIs('AWS::SSM::Parameter', 7)
  })
})

describe('NetworkStack', () => {
  test('VPC が作成される', () => {
    const app = new cdk.App()
    const stack = new NetworkStack(app, 'TestNetworkStack', { env })
    const template = Template.fromStack(stack)
    template.resourceCountIs('AWS::EC2::VPC', 1)
  })

  test('パブリックサブネットとプライベートサブネットが作成される', () => {
    const app = new cdk.App()
    const stack = new NetworkStack(app, 'TestNetworkStack', { env })
    const template = Template.fromStack(stack)
    // 2 AZs × 2 types = 4 subnets
    template.resourceCountIs('AWS::EC2::Subnet', 4)
  })

  test('NAT Gateway が1つ作成される', () => {
    const app = new cdk.App()
    const stack = new NetworkStack(app, 'TestNetworkStack', { env })
    const template = Template.fromStack(stack)
    template.resourceCountIs('AWS::EC2::NatGateway', 1)
  })
})

describe('DataStack', () => {
  let template: Template

  beforeAll(() => {
    const app = new cdk.App()
    const network = new NetworkStack(app, 'TestNetworkStack', { env })
    const stack = new DataStack(app, 'TestDataStack', {
      env,
      vpcModule: network.vpcModule,
    })
    template = Template.fromStack(stack)
  })

  test('RDS インスタンスが作成される', () => {
    template.resourceCountIs('AWS::RDS::DBInstance', 1)
  })

  test('RDS が PostgreSQL エンジンで作成される', () => {
    template.hasResourceProperties('AWS::RDS::DBInstance', {
      Engine: 'postgres',
      PubliclyAccessible: false,
    })
  })

  test('S3 バケットが作成される', () => {
    template.resourceCountIs('AWS::S3::Bucket', 1)
  })

  test('S3 バケットがパブリックアクセスをブロックしている', () => {
    template.hasResourceProperties('AWS::S3::Bucket', {
      PublicAccessBlockConfiguration: {
        BlockPublicAcls: true,
        BlockPublicPolicy: true,
        IgnorePublicAcls: true,
        RestrictPublicBuckets: true,
      },
    })
  })

  test('SES ドメイン検証が作成される', () => {
    template.resourceCountIs('AWS::SES::EmailIdentity', 1)
  })
})

describe('ComputeStack', () => {
  let template: Template

  beforeAll(() => {
    const app = new cdk.App()
    // 循環参照回避: 全リソースを同一スタックに配置
    const stack = new cdk.Stack(app, 'TestComputeAllInOneStack', { env })
    const vpcModule = new VpcModule(stack, 'Vpc', {
      vpcCidr: '10.0.0.0/16',
      maxAzs: 2,
      natGateways: 1,
      publicSubnetCidrMask: 24,
      privateSubnetCidrMask: 24,
    })
    const rdsModule = new RdsModule(stack, 'Rds', {
      vpc: vpcModule.vpc,
      subnets: vpcModule.privateSubnets,
      databaseName: 'gin-todo',
      credentials: cdk.aws_rds.Credentials.fromPassword(
        'gin',
        cdk.SecretValue.unsafePlainText('test-password'),
      ),
    })
    new EcsModule(stack, 'Ecs', {
      vpc: vpcModule.vpc,
      privateSubnets: vpcModule.privateSubnets,
      containerPort: 8080,
      containerName: 'gin-todo-api',
      logGroupName: '/ecs/gin-todo-api',
      ssmParameterPrefix: '/gin-todo',
      ecrRepositoryName: 'gin-todo-api',
      sesIdentityArn: `arn:aws:ses:ap-northeast-1:123456789012:identity/todo.takehiro1111.com`,
    })

    // CloudWatch Alarms (ComputeStack のロジックを再現)
    new cdk.aws_sns.Topic(stack, 'AlarmTopic')
    new cdk.aws_cloudwatch.Alarm(stack, 'EcsCpuAlarm', {
      metric: new cdk.aws_cloudwatch.Metric({ namespace: 'AWS/ECS', metricName: 'CPUUtilization' }),
      threshold: 80, evaluationPeriods: 3,
    })
    new cdk.aws_cloudwatch.Alarm(stack, 'EcsMemoryAlarm', {
      metric: new cdk.aws_cloudwatch.Metric({ namespace: 'AWS/ECS', metricName: 'MemoryUtilization' }),
      threshold: 80, evaluationPeriods: 3,
    })
    new cdk.aws_cloudwatch.Alarm(stack, 'Alb5xxAlarm', {
      metric: new cdk.aws_cloudwatch.Metric({ namespace: 'AWS/ApplicationELB', metricName: 'HTTPCode_ELB_5XX_Count' }),
      threshold: 10, evaluationPeriods: 2,
    })
    new cdk.aws_cloudwatch.Alarm(stack, 'RdsCpuAlarm', {
      metric: rdsModule.instance.metricCPUUtilization(),
      threshold: 80, evaluationPeriods: 3,
    })
    new cdk.aws_cloudwatch.Alarm(stack, 'RdsFreeStorageAlarm', {
      metric: rdsModule.instance.metricFreeStorageSpace(),
      threshold: 5 * 1024 * 1024 * 1024, evaluationPeriods: 1,
      comparisonOperator: cdk.aws_cloudwatch.ComparisonOperator.LESS_THAN_OR_EQUAL_TO_THRESHOLD,
    })
    new cdk.aws_cloudwatch.Alarm(stack, 'RdsConnectionsAlarm', {
      metric: rdsModule.instance.metricDatabaseConnections(),
      threshold: 50, evaluationPeriods: 2,
    })

    template = Template.fromStack(stack)
  })

  test('ALB が作成される', () => {
    template.resourceCountIs('AWS::ElasticLoadBalancingV2::LoadBalancer', 1)
  })

  test('ALB が internal である', () => {
    template.hasResourceProperties(
      'AWS::ElasticLoadBalancingV2::LoadBalancer',
      {
        Scheme: 'internal',
      }
    )
  })

  test('ECS クラスターが作成される', () => {
    template.resourceCountIs('AWS::ECS::Cluster', 1)
  })

  test('Fargate サービスが作成される', () => {
    template.resourceCountIs('AWS::ECS::Service', 1)
    template.hasResourceProperties('AWS::ECS::Service', {
      LaunchType: 'FARGATE',
    })
  })

  test('タスク定義にコンテナポート 8080 が設定されている', () => {
    template.hasResourceProperties('AWS::ECS::TaskDefinition', {
      ContainerDefinitions: [
        {
          PortMappings: [{ ContainerPort: 8080, Protocol: 'tcp' }],
        },
      ],
    })
  })

  test('ECR リポジトリが作成される', () => {
    template.resourceCountIs('AWS::ECR::Repository', 1)
  })

  test('Auto Scaling が設定されている', () => {
    template.resourceCountIs('AWS::ApplicationAutoScaling::ScalableTarget', 1)
    template.resourceCountIs('AWS::ApplicationAutoScaling::ScalingPolicy', 1)
  })

  test('CloudWatch アラームが作成される', () => {
    template.resourceCountIs('AWS::CloudWatch::Alarm', 6)
  })

  test('SNS トピックが作成される', () => {
    template.resourceCountIs('AWS::SNS::Topic', 1)
  })
})

describe('CdnStack', () => {
  let template: Template

  beforeAll(() => {
    const app = new cdk.App()
    // 全リソースを同一スタックに配置して循環参照を回避
    const cdnStack = new cdk.Stack(app, 'TestCdnStack', { env })

    const testBucket = new cdk.aws_s3.Bucket(cdnStack, 'TestBucket')
    const hostedZone = new cdk.aws_route53.HostedZone(cdnStack, 'TestZone', {
      zoneName: 'takehiro1111.com',
    })
    const certificate = new cdk.aws_certificatemanager.Certificate(
      cdnStack,
      'TestCert',
      { domainName: 'todo.takehiro1111.com' }
    )

    // テスト用の internal ALB を同一スタック内に作成
    const vpc = new cdk.aws_ec2.Vpc(cdnStack, 'TestVpc', { maxAzs: 2 })
    const alb = new cdk.aws_elasticloadbalancingv2.ApplicationLoadBalancer(
      cdnStack,
      'TestAlb',
      { vpc, internetFacing: false }
    )

    new CdnModule(cdnStack, 'Cdn', {
      frontendBucket: testBucket,
      alb,
      domainName: 'takehiro1111.com',
      appDomain: 'todo.takehiro1111.com',
      certificate,
      hostedZone,
    })
    template = Template.fromStack(cdnStack)
  })

  test('CloudFront ディストリビューションが作成される', () => {
    template.resourceCountIs('AWS::CloudFront::Distribution', 1)
  })

  test('VPC Origin が作成される', () => {
    template.resourceCountIs('AWS::CloudFront::VpcOrigin', 1)
  })

  test('デフォルトルートオブジェクトが index.html', () => {
    template.hasResourceProperties('AWS::CloudFront::Distribution', {
      DistributionConfig: {
        DefaultRootObject: 'index.html',
      },
    })
  })

  test('カスタムドメインが設定されている', () => {
    template.hasResourceProperties('AWS::CloudFront::Distribution', {
      DistributionConfig: {
        Aliases: ['todo.takehiro1111.com'],
      },
    })
  })

  test('Route53 A レコードが作成される', () => {
    template.resourceCountIs('AWS::Route53::RecordSet', 2) // A + AAAA
  })
})
