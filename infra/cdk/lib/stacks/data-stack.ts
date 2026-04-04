import * as rds from 'aws-cdk-lib/aws-rds'
import * as s3 from 'aws-cdk-lib/aws-s3'
import * as ses from 'aws-cdk-lib/aws-ses'
import { RemovalPolicy, SecretValue, Stack, StackProps } from 'aws-cdk-lib'
import { Construct } from 'constructs'
import { VpcModule } from '@/lib/modules/vpc/index.js'
import { RdsModule } from '@/lib/modules/rds/index.js'
import { ENV } from '@/env/env.js'

export interface DataStackProps extends StackProps {
  vpcModule: VpcModule
}

export class DataStack extends Stack {
  public readonly rdsModule: RdsModule
  public readonly frontendBucket: s3.Bucket
  public readonly sesIdentityArn: string

  constructor(scope: Construct, id: string, props: DataStackProps) {
    super(scope, id, props)

    this.rdsModule = new RdsModule(this, 'Rds', {
      vpc: props.vpcModule.vpc,
      subnets: props.vpcModule.privateSubnets,
      databaseName: ENV.dbName,
      // SSM Parameter Store と同じパスワードを使用
      credentials: rds.Credentials.fromPassword(
        ENV.dbUser,
        SecretValue.unsafePlainText(ENV.postgresPassword),
      ),
    })

    this.frontendBucket = new s3.Bucket(this, 'FrontendBucket', {
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      encryption: s3.BucketEncryption.S3_MANAGED,
      enforceSSL: true,
      removalPolicy: RemovalPolicy.RETAIN,
      autoDeleteObjects: false,
    })

    // SES ドメイン検証 (パスワードリセットメール送信用)
    const sesIdentity = new ses.EmailIdentity(this, 'SesEmailIdentity', {
      identity: ses.Identity.domain(ENV.appDomain),
    })
    this.sesIdentityArn = sesIdentity.emailIdentityArn
  }
}
