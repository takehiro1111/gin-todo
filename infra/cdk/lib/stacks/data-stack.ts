import * as ec2 from 'aws-cdk-lib/aws-ec2'
import * as rds from 'aws-cdk-lib/aws-rds'
import * as ses from 'aws-cdk-lib/aws-ses'
import { SecretValue, Stack, StackProps } from 'aws-cdk-lib'
import { Construct } from 'constructs'
import { VpcModule } from '@/lib/modules/vpc/index.js'
import { RdsModule } from '@/lib/modules/rds/index.js'
import { ENV } from '@/env/env.js'

export interface DataStackProps extends StackProps {
  vpcModule: VpcModule
  rdsSecurityGroup: ec2.ISecurityGroup
}

export class DataStack extends Stack {
  public readonly rdsModule: RdsModule

  constructor(scope: Construct, id: string, props: DataStackProps) {
    super(scope, id, props)

    this.rdsModule = new RdsModule(this, 'Rds', {
      vpc: props.vpcModule.vpc,
      subnets: props.vpcModule.privateSubnets,
      securityGroup: props.rdsSecurityGroup,
      databaseName: ENV.dbName,
      credentials: rds.Credentials.fromPassword(
        ENV.dbUser,
        SecretValue.unsafePlainText(ENV.postgresPassword),
      ),
    })

    // SES ドメイン検証 (パスワードリセットメール送信用)
    new ses.EmailIdentity(this, 'SesEmailIdentity', {
      identity: ses.Identity.domain(ENV.appDomain),
    })
  }
}
