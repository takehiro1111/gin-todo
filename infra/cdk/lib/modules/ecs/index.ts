import * as ec2 from 'aws-cdk-lib/aws-ec2'
import * as ecr from 'aws-cdk-lib/aws-ecr'
import * as ecs from 'aws-cdk-lib/aws-ecs'
import * as ecs_patterns from 'aws-cdk-lib/aws-ecs-patterns'
import * as elbv2 from 'aws-cdk-lib/aws-elasticloadbalancingv2'
import * as iam from 'aws-cdk-lib/aws-iam'
import * as logs from 'aws-cdk-lib/aws-logs'
import { Duration, RemovalPolicy } from 'aws-cdk-lib'
import { Construct } from 'constructs'

export interface EcsModuleProps {
  vpc: ec2.IVpc
  privateSubnets: ec2.ISubnet[]
  albSecurityGroup: ec2.ISecurityGroup
  ecsSecurityGroup: ec2.ISecurityGroup
  ssmParameterPrefix: string
  sesIdentityArn: string

  // Backend
  backendContainerPort: number
  backendContainerName: string
  backendLogGroupName: string
  backendEcrRepositoryName: string
  backendCpu?: number
  backendMemoryLimitMiB?: number
  backendDesiredCount?: number
  backendEnvironment?: Record<string, string>

  // Frontend
  frontendContainerPort: number
  frontendContainerName: string
  frontendLogGroupName: string
  frontendEcrRepositoryName: string
  frontendCpu?: number
  frontendMemoryLimitMiB?: number
  frontendDesiredCount?: number

  // Auto Scaling
  minCapacity?: number
  maxCapacity?: number
  cpuTargetUtilization?: number
}

export class EcsModule extends Construct {
  public readonly fargateService: ecs_patterns.ApplicationLoadBalancedFargateService
  public readonly backendRepository: ecr.Repository
  public readonly frontendRepository: ecr.Repository

  constructor(scope: Construct, id: string, props: EcsModuleProps) {
    super(scope, id)

    // --- ECR Repositories ---

    this.backendRepository = new ecr.Repository(this, 'BackendRepository', {
      repositoryName: props.backendEcrRepositoryName,
      removalPolicy: RemovalPolicy.RETAIN,
      lifecycleRules: [{ maxImageCount: 10, description: 'Keep last 10 images' }],
    })

    this.frontendRepository = new ecr.Repository(this, 'FrontendRepository', {
      repositoryName: props.frontendEcrRepositoryName,
      removalPolicy: RemovalPolicy.RETAIN,
      lifecycleRules: [{ maxImageCount: 10, description: 'Keep last 10 images' }],
    })

    // --- Cluster ---

    const cluster = new ecs.Cluster(this, 'Cluster', {
      vpc: props.vpc,
      defaultCloudMapNamespace: {
        name: 'gin-todo.local',
      },
    })

    // --- Frontend (L3: default route) ---

    const frontendLogGroup = new logs.LogGroup(this, 'FrontendLogGroup', {
      logGroupName: props.frontendLogGroupName,
      retention: logs.RetentionDays.TWO_WEEKS,
      removalPolicy: RemovalPolicy.DESTROY,
    })

    // ALB を事前に作成 (SecurityStack の SG を使用し、L3 の自動生成 SG による循環依存を防止)
    const alb = new elbv2.ApplicationLoadBalancer(this, 'Alb', {
      vpc: props.vpc,
      internetFacing: false,
      securityGroup: props.albSecurityGroup,
      vpcSubnets: { subnets: props.privateSubnets },
    })

    // L3 で Frontend Service を作成 (default route → frontend)
    this.fargateService =
      new ecs_patterns.ApplicationLoadBalancedFargateService(
        this,
        'FargateService',
        {
          cluster,
          loadBalancer: alb,
          cpu: props.frontendCpu ?? 256,
          memoryLimitMiB: props.frontendMemoryLimitMiB ?? 512,
          desiredCount: props.frontendDesiredCount ?? 1,
          assignPublicIp: false,
          taskSubnets: { subnets: props.privateSubnets },
          taskImageOptions: {
            image: ecs.ContainerImage.fromRegistry('amazon/amazon-ecs-sample'),
            containerName: props.frontendContainerName,
            containerPort: props.frontendContainerPort,
            logDriver: ecs.LogDrivers.awsLogs({
              logGroup: frontendLogGroup,
              streamPrefix: 'frontend',
            }),
          },
          securityGroups: [props.ecsSecurityGroup],
          // ヘルスチェックは ALB ターゲットグループで実施
        }
      )

    // Service Connect: Frontend は client-only (Backend を呼ぶ側)
    this.fargateService.service.enableServiceConnect({
      namespace: cluster.defaultCloudMapNamespace!.namespaceName,
    })

    // Frontend ヘルスチェック
    this.fargateService.targetGroup.configureHealthCheck({
      path: '/',
      interval: Duration.seconds(30),
      timeout: Duration.seconds(5),
      healthyThresholdCount: 2,
      unhealthyThresholdCount: 3,
    })

    // --- Backend ---

    const backendLogGroup = new logs.LogGroup(this, 'BackendLogGroup', {
      logGroupName: props.backendLogGroupName,
      retention: logs.RetentionDays.TWO_WEEKS,
      removalPolicy: RemovalPolicy.DESTROY,
    })

    const backendTaskDef = new ecs.FargateTaskDefinition(this, 'BackendTaskDef', {
      cpu: props.backendCpu ?? 256,
      memoryLimitMiB: props.backendMemoryLimitMiB ?? 512,
    })

    backendTaskDef.addContainer('BackendContainer', {
      image: ecs.ContainerImage.fromRegistry('amazon/amazon-ecs-sample'),
      containerName: props.backendContainerName,
      portMappings: [{
        containerPort: props.backendContainerPort,
        name: 'backend',
      }],
      environment: {
        GIN_MODE: 'release',
        PORT: String(props.backendContainerPort),
        ENV: 'production',
        ...props.backendEnvironment,
      },
      logging: ecs.LogDrivers.awsLogs({
        logGroup: backendLogGroup,
        streamPrefix: 'api',
      }),
    })

    // Backend IAM: SSM 読み取り
    backendTaskDef.taskRole.addToPrincipalPolicy(
      new iam.PolicyStatement({
        actions: ['ssm:GetParameter', 'ssm:GetParameters'],
        resources: [`arn:aws:ssm:*:*:parameter${props.ssmParameterPrefix}/*`],
      })
    )

    // Backend IAM: SES メール送信 (ドメイン Identity に制限)
    backendTaskDef.taskRole.addToPrincipalPolicy(
      new iam.PolicyStatement({
        actions: ['ses:SendEmail', 'ses:SendRawEmail'],
        resources: [props.sesIdentityArn],
      })
    )

    const backendService = new ecs.FargateService(this, 'BackendService', {
      cluster,
      taskDefinition: backendTaskDef,
      desiredCount: props.backendDesiredCount ?? 1,
      assignPublicIp: false,
      vpcSubnets: { subnets: props.privateSubnets },
      securityGroups: [props.ecsSecurityGroup],
      serviceConnectConfiguration: {
        namespace: cluster.defaultCloudMapNamespace!.namespaceName,
        services: [{
          portMappingName: 'backend',
          dnsName: 'backend',
          port: props.backendContainerPort,
        }],
      },
    })

    // Backend ターゲットグループ
    const backendTargetGroup = new elbv2.ApplicationTargetGroup(this, 'BackendTargetGroup', {
      vpc: props.vpc,
      port: props.backendContainerPort,
      protocol: elbv2.ApplicationProtocol.HTTP,
      targetType: elbv2.TargetType.IP,
      healthCheck: {
        path: '/api/health',
        interval: Duration.seconds(30),
        timeout: Duration.seconds(5),
        healthyThresholdCount: 2,
        unhealthyThresholdCount: 3,
      },
    })

    backendService.attachToApplicationTargetGroup(backendTargetGroup)

    // WebSocket 用のスティッキーセッション
    backendTargetGroup.setAttribute('stickiness.enabled', 'true')
    backendTargetGroup.setAttribute('stickiness.type', 'lb_cookie')
    backendTargetGroup.setAttribute('stickiness.lb_cookie.duration_seconds', '86400')

    // ALB リスナールール: /api/* → Backend
    const listener = this.fargateService.listener

    listener.addAction('ApiRoute', {
      priority: 10,
      conditions: [elbv2.ListenerCondition.pathPatterns(['/api/*'])],
      action: elbv2.ListenerAction.forward([backendTargetGroup]),
    })

    // --- Auto Scaling (Backend) ---

    const backendScaling = backendService.autoScaleTaskCount({
      minCapacity: props.minCapacity ?? 1,
      maxCapacity: props.maxCapacity ?? 4,
    })

    backendScaling.scaleOnCpuUtilization('BackendCpuScaling', {
      targetUtilizationPercent: props.cpuTargetUtilization ?? 70,
      scaleInCooldown: Duration.seconds(60),
      scaleOutCooldown: Duration.seconds(60),
    })
  }
}
