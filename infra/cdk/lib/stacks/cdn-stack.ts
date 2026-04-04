import * as acm from 'aws-cdk-lib/aws-certificatemanager'
import * as elbv2 from 'aws-cdk-lib/aws-elasticloadbalancingv2'
import * as route53 from 'aws-cdk-lib/aws-route53'
import * as s3 from 'aws-cdk-lib/aws-s3'
import { CfnOutput, RemovalPolicy, Stack, StackProps } from 'aws-cdk-lib'
import { Construct } from 'constructs'
import { CdnModule } from '@/lib/modules/cdn/index.js'
import { ENV } from '@/env/env.js'

export interface CdnStackProps extends StackProps {
  alb: elbv2.IApplicationLoadBalancer
  hostedZone: route53.IHostedZone
}

export class CdnStack extends Stack {
  public readonly cdnModule: CdnModule

  constructor(scope: Construct, id: string, props: CdnStackProps) {
    super(scope, id, props)

    // 静的アセット用 S3 バケット (CloudFront OAC 経由でのみアクセス)
    const staticAssetsBucket = new s3.Bucket(this, 'StaticAssetsBucket', {
      blockPublicAccess: s3.BlockPublicAccess.BLOCK_ALL,
      encryption: s3.BucketEncryption.S3_MANAGED,
      enforceSSL: true,
      removalPolicy: RemovalPolicy.RETAIN,
      autoDeleteObjects: false,
    })

    // CloudFront 用 SSL 証明書 (us-east-1 に作成)
    const certificate = new acm.DnsValidatedCertificate(this, 'Certificate', {
      domainName: ENV.appDomain,
      hostedZone: props.hostedZone,
      region: 'us-east-1',
    })

    this.cdnModule = new CdnModule(this, 'Cdn', {
      staticAssetsBucket,
      alb: props.alb,
      domainName: ENV.domainName,
      appDomain: ENV.appDomain,
      certificate,
      hostedZone: props.hostedZone,
    })

    new CfnOutput(this, 'AppUrl', {
      value: `https://${ENV.appDomain}`,
      description: 'Application URL',
    })

    new CfnOutput(this, 'DistributionId', {
      value: this.cdnModule.distribution.distributionId,
      description: 'CloudFront distribution ID',
    })
  }
}
