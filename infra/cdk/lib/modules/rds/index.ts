import * as ec2 from 'aws-cdk-lib/aws-ec2'
import * as rds from 'aws-cdk-lib/aws-rds'
import { Duration, RemovalPolicy } from 'aws-cdk-lib'
import { Construct } from 'constructs'

export interface RdsModuleProps {
  vpc: ec2.IVpc
  subnets: ec2.ISubnet[]
  databaseName: string
  credentials: rds.Credentials
  instanceType?: ec2.InstanceType
}

export class RdsModule extends Construct {
  public readonly instance: rds.DatabaseInstance
  public readonly securityGroup: ec2.SecurityGroup

  constructor(scope: Construct, id: string, props: RdsModuleProps) {
    super(scope, id)

    this.securityGroup = new ec2.SecurityGroup(this, 'SecurityGroup', {
      vpc: props.vpc,
      description: 'Security group for RDS PostgreSQL',
      allowAllOutbound: false,
    })

    this.instance = new rds.DatabaseInstance(this, 'Instance', {
      engine: rds.DatabaseInstanceEngine.postgres({
        version: rds.PostgresEngineVersion.VER_16_4,
      }),
      instanceType:
        props.instanceType ??
        ec2.InstanceType.of(ec2.InstanceClass.T3, ec2.InstanceSize.MICRO),
      vpc: props.vpc,
      vpcSubnets: { subnets: props.subnets },
      securityGroups: [this.securityGroup],
      databaseName: props.databaseName,
      credentials: props.credentials,
      multiAz: false,
      allocatedStorage: 20,
      maxAllocatedStorage: 100,
      storageEncrypted: true,
      backupRetention: Duration.days(7),
      deletionProtection: true,
      publiclyAccessible: false,
      removalPolicy: RemovalPolicy.RETAIN,
    })
  }
}
