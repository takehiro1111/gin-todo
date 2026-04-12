import * as ec2 from 'aws-cdk-lib/aws-ec2'
import { Construct } from 'constructs'

export interface SecurityModuleProps {
  vpc: ec2.IVpc
  dbPort: number
}

export class SecurityModule extends Construct {
  public readonly albSecurityGroup: ec2.SecurityGroup
  public readonly ecsSecurityGroup: ec2.SecurityGroup
  public readonly rdsSecurityGroup: ec2.SecurityGroup

  constructor(scope: Construct, id: string, props: SecurityModuleProps) {
    super(scope, id)

    // ALB SG
    // Inbound: VPC Origin SG (CloudFront デプロイ後に自動作成)
    this.albSecurityGroup = new ec2.SecurityGroup(this, 'AlbSecurityGroup', {
      vpc: props.vpc,
      description: 'Security group for ALB (internal)',
    })

    // ECS SG (Frontend/Backend 共有)
    this.ecsSecurityGroup = new ec2.SecurityGroup(this, 'EcsSecurityGroup', {
      vpc: props.vpc,
      description: 'Security group for ECS (Frontend/Backend shared)',
    })

    // RDS SG
    this.rdsSecurityGroup = new ec2.SecurityGroup(this, 'RdsSecurityGroup', {
      vpc: props.vpc,
      description: 'Security group for RDS PostgreSQL',
      allowAllOutbound: false,
    })

    // --- SG ルール ---

    // ALB → ECS: 全ポート
    this.ecsSecurityGroup.addIngressRule(
      this.albSecurityGroup,
      ec2.Port.allTraffic(),
      'Allow all traffic from ALB'
    )

    // ECS セルフ: コンテナ間通信 (Service Connect)
    this.ecsSecurityGroup.addIngressRule(
      this.ecsSecurityGroup,
      ec2.Port.allTraffic(),
      'Allow ECS container-to-container communication'
    )

    // ECS → RDS: PostgreSQL
    this.rdsSecurityGroup.addIngressRule(
      this.ecsSecurityGroup,
      ec2.Port.tcp(props.dbPort),
      'Allow PostgreSQL from ECS'
    )
  }
}
