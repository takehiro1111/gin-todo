import * as ec2 from 'aws-cdk-lib/aws-ec2'
import * as ecr from 'aws-cdk-lib/aws-ecr'
import * as ecs from 'aws-cdk-lib/aws-ecs'
import * as ecs_patterns from 'aws-cdk-lib/aws-ecs-patterns'
import * as iam from 'aws-cdk-lib/aws-iam'
import * as logs from 'aws-cdk-lib/aws-logs'
import { Duration, RemovalPolicy } from 'aws-cdk-lib'
import { Construct } from 'constructs'

export interface EcsModuleProps {
  vpc: ec2.IVpc
  privateSubnets: ec2.ISubnet[]
  containerPort: number
  containerName: string
  logGroupName: string
  ssmParameterPrefix: string
  ecrRepositoryName: string
  sesIdentityArn: string
  cpu?: number
  memoryLimitMiB?: number
  desiredCount?: number
  minCapacity?: number
  maxCapacity?: number
  cpuTargetUtilization?: number
  environment?: Record<string, string>
}

export class EcsModule extends Construct {
  public readonly fargateService: ecs_patterns.ApplicationLoadBalancedFargateService
  public readonly securityGroup: ec2.SecurityGroup
  public readonly repository: ecr.Repository

  constructor(scope: Construct, id: string, props: EcsModuleProps) {
    super(scope, id)

    // ECR Repository
    this.repository = new ecr.Repository(this, 'Repository', {
      repositoryName: props.ecrRepositoryName,
      removalPolicy: RemovalPolicy.RETAIN,
      lifecycleRules: [
        {
          maxImageCount: 10,
          description: 'Keep last 10 images',
        },
      ],
    })

    const cluster = new ecs.Cluster(this, 'Cluster', {
      vpc: props.vpc,
    })

    const logGroup = new logs.LogGroup(this, 'LogGroup', {
      logGroupName: props.logGroupName,
      retention: logs.RetentionDays.TWO_WEEKS,
      removalPolicy: RemovalPolicy.DESTROY,
    })

    this.fargateService =
      new ecs_patterns.ApplicationLoadBalancedFargateService(
        this,
        'FargateService',
        {
          cluster,
          cpu: props.cpu ?? 256,
          memoryLimitMiB: props.memoryLimitMiB ?? 512,
          desiredCount: props.desiredCount ?? 1,
          assignPublicIp: false,
          taskSubnets: { subnets: props.privateSubnets },
          // ALB は internal — CloudFront VPC Origin 経由でのみアクセス
          publicLoadBalancer: false,
          taskImageOptions: {
            // 初回デプロイ時はプレースホルダーイメージを使用
            // CI/CD パイプラインで ECR イメージに差し替える
            image: ecs.ContainerImage.fromRegistry(
              'amazon/amazon-ecs-sample'
            ),
            containerName: props.containerName,
            containerPort: props.containerPort,
            environment: {
              GIN_MODE: 'release',
              PORT: String(props.containerPort),
              ENV: 'production',
              ...props.environment,
            },
            // DB 接続情報は SSM Parameter Store から Go アプリが直接取得
            logDriver: ecs.LogDrivers.awsLogs({
              logGroup,
              streamPrefix: 'api',
            }),
          },
          // ヘルスチェックは ALB ターゲットグループで実施
        }
      )

    // ALB ターゲットグループのヘルスチェック
    this.fargateService.targetGroup.configureHealthCheck({
      path: '/api/health',
      interval: Duration.seconds(30),
      timeout: Duration.seconds(5),
      healthyThresholdCount: 2,
      unhealthyThresholdCount: 3,
    })

    // ALB: WebSocket 用のスティッキーセッションを有効化
    this.fargateService.targetGroup.setAttribute(
      'stickiness.enabled',
      'true'
    )
    this.fargateService.targetGroup.setAttribute(
      'stickiness.type',
      'lb_cookie'
    )
    this.fargateService.targetGroup.setAttribute(
      'stickiness.lb_cookie.duration_seconds',
      '86400'
    )

    // --- IAM Permissions ---

    // SSM Parameter Store 読み取り
    this.fargateService.taskDefinition.taskRole.addToPrincipalPolicy(
      new iam.PolicyStatement({
        actions: ['ssm:GetParameter', 'ssm:GetParameters'],
        resources: [`arn:aws:ssm:*:*:parameter${props.ssmParameterPrefix}/*`],
      })
    )

    // SES メール送信 (ドメイン Identity に制限)
    this.fargateService.taskDefinition.taskRole.addToPrincipalPolicy(
      new iam.PolicyStatement({
        actions: ['ses:SendEmail', 'ses:SendRawEmail'],
        resources: [props.sesIdentityArn],
      })
    )

    // --- Auto Scaling ---
    const scaling = this.fargateService.service.autoScaleTaskCount({
      minCapacity: props.minCapacity ?? 1,
      maxCapacity: props.maxCapacity ?? 4,
    })

    scaling.scaleOnCpuUtilization('CpuScaling', {
      targetUtilizationPercent: props.cpuTargetUtilization ?? 70,
      scaleInCooldown: Duration.seconds(60),
      scaleOutCooldown: Duration.seconds(60),
    })

    this.securityGroup =
      this.fargateService.service.connections.securityGroups[0] as ec2.SecurityGroup

    // コンテナ間通信のためセルフ参照を許可
    this.securityGroup.addIngressRule(
      this.securityGroup,
      ec2.Port.allTraffic(),
      'Allow ECS container-to-container communication'
    )
  }
}
