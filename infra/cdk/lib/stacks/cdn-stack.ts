import * as acm from 'aws-cdk-lib/aws-certificatemanager'
import * as elbv2 from 'aws-cdk-lib/aws-elasticloadbalancingv2'
import * as route53 from 'aws-cdk-lib/aws-route53'
import * as s3 from 'aws-cdk-lib/aws-s3'
import { CfnOutput, Stack, StackProps } from 'aws-cdk-lib'
import { Construct } from 'constructs'
import { CdnModule } from '@/lib/modules/cdn/index.js'
import { ENV } from '@/env/env.js'

export interface CdnStackProps extends StackProps {
  frontendBucket: s3.IBucket
  alb: elbv2.IApplicationLoadBalancer
}

export class CdnStack extends Stack {
  public readonly cdnModule: CdnModule

  constructor(scope: Construct, id: string, props: CdnStackProps) {
    super(scope, id, props)

    // 既存の HostedZone を参照
    const hostedZone = route53.HostedZone.fromLookup(this, 'HostedZone', {
      domainName: ENV.domainName,
    })

    // CloudFront 用 SSL 証明書 (us-east-1 に作成)
    const certificate = new acm.DnsValidatedCertificate(this, 'Certificate', {
      domainName: ENV.appDomain,
      hostedZone,
      region: 'us-east-1',
    })

    this.cdnModule = new CdnModule(this, 'Cdn', {
      frontendBucket: props.frontendBucket,
      alb: props.alb,
      domainName: ENV.domainName,
      appDomain: ENV.appDomain,
      certificate,
      hostedZone,
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
