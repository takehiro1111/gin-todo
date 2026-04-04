import { Duration, Stack, StackProps } from 'aws-cdk-lib'
import * as cloudwatch from 'aws-cdk-lib/aws-cloudwatch'
import * as cloudwatch_actions from 'aws-cdk-lib/aws-cloudwatch-actions'
import * as ec2 from 'aws-cdk-lib/aws-ec2'
import * as sns from 'aws-cdk-lib/aws-sns'
import { Construct } from 'constructs'
import { VpcModule } from '@/lib/modules/vpc/index.js'
import { RdsModule } from '@/lib/modules/rds/index.js'
import { EcsModule } from '@/lib/modules/ecs/index.js'
import { ENV } from '@/env/env.js'

export interface ComputeStackProps extends StackProps {
  vpcModule: VpcModule
  rdsModule: RdsModule
  sesIdentityArn: string
}

export class ComputeStack extends Stack {
  public readonly ecsModule: EcsModule

  constructor(scope: Construct, id: string, props: ComputeStackProps) {
    super(scope, id, props)

    this.ecsModule = new EcsModule(this, 'Ecs', {
      vpc: props.vpcModule.vpc,
      privateSubnets: props.vpcModule.privateSubnets,
      containerPort: ENV.containerPort,
      cpu: ENV.cpu,
      memoryLimitMiB: ENV.memoryLimitMiB,
      desiredCount: ENV.desiredCount,
      containerName: ENV.containerName,
      logGroupName: ENV.logGroupName,
      ssmParameterPrefix: ENV.ssmParameterPrefix,
      ecrRepositoryName: ENV.ecrRepositoryName,
      sesIdentityArn: props.sesIdentityArn,
      minCapacity: ENV.minCapacity,
      maxCapacity: ENV.maxCapacity,
      cpuTargetUtilization: ENV.cpuTargetUtilization,
    })

    // RDS へのアクセスを許可
    props.rdsModule.securityGroup.addIngressRule(
      this.ecsModule.securityGroup,
      ec2.Port.tcp(ENV.dbPort),
      'Allow PostgreSQL from ECS'
    )

    // --- CloudWatch Alarms ---

    const alarmTopic = new sns.Topic(this, 'AlarmTopic', {
      displayName: 'GinTodo Alarms',
    })

    // ECS CPU 使用率
    new cloudwatch.Alarm(this, 'EcsCpuAlarm', {
      metric:
        this.ecsModule.fargateService.service.metricCpuUtilization(),
      threshold: 80,
      evaluationPeriods: 3,
      alarmDescription: 'ECS CPU utilization > 80% for 3 periods',
    }).addAlarmAction(new cloudwatch_actions.SnsAction(alarmTopic))

    // ECS メモリ使用率
    new cloudwatch.Alarm(this, 'EcsMemoryAlarm', {
      metric:
        this.ecsModule.fargateService.service.metricMemoryUtilization(),
      threshold: 80,
      evaluationPeriods: 3,
      alarmDescription: 'ECS Memory utilization > 80% for 3 periods',
    }).addAlarmAction(new cloudwatch_actions.SnsAction(alarmTopic))

    // ALB 5xx エラー
    new cloudwatch.Alarm(this, 'Alb5xxAlarm', {
      metric: new cloudwatch.Metric({
        namespace: 'AWS/ApplicationELB',
        metricName: 'HTTPCode_ELB_5XX_Count',
        dimensionsMap: {
          LoadBalancer:
            this.ecsModule.fargateService.loadBalancer
              .loadBalancerFullName,
        },
        statistic: 'Sum',
        period: Duration.minutes(1),
      }),
      threshold: 10,
      evaluationPeriods: 2,
      alarmDescription: 'ALB 5xx errors > 10 for 2 periods',
    }).addAlarmAction(new cloudwatch_actions.SnsAction(alarmTopic))

    // RDS CPU 使用率
    new cloudwatch.Alarm(this, 'RdsCpuAlarm', {
      metric: props.rdsModule.instance.metricCPUUtilization(),
      threshold: 80,
      evaluationPeriods: 3,
      alarmDescription: 'RDS CPU utilization > 80% for 3 periods',
    }).addAlarmAction(new cloudwatch_actions.SnsAction(alarmTopic))

    // RDS 空きストレージ (5GB 以下)
    new cloudwatch.Alarm(this, 'RdsFreeStorageAlarm', {
      metric: props.rdsModule.instance.metricFreeStorageSpace(),
      threshold: 5 * 1024 * 1024 * 1024, // 5GB in bytes
      evaluationPeriods: 1,
      comparisonOperator:
        cloudwatch.ComparisonOperator.LESS_THAN_OR_EQUAL_TO_THRESHOLD,
      alarmDescription: 'RDS free storage <= 5GB',
    }).addAlarmAction(new cloudwatch_actions.SnsAction(alarmTopic))

    // RDS DB 接続数
    new cloudwatch.Alarm(this, 'RdsConnectionsAlarm', {
      metric: props.rdsModule.instance.metricDatabaseConnections(),
      threshold: 50,
      evaluationPeriods: 2,
      alarmDescription: 'RDS connections > 50 for 2 periods',
    }).addAlarmAction(new cloudwatch_actions.SnsAction(alarmTopic))
  }
}
