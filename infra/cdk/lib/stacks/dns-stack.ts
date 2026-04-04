import * as route53 from 'aws-cdk-lib/aws-route53'
import { CfnOutput, Fn, Stack, StackProps } from 'aws-cdk-lib'
import { Construct } from 'constructs'
import { ENV } from '@/env/env.js'

export class DnsStack extends Stack {
  public readonly hostedZone: route53.PublicHostedZone

  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props)

    this.hostedZone = new route53.PublicHostedZone(this, 'HostedZone', {
      zoneName: ENV.domainName,
    })

    new CfnOutput(this, 'NameServers', {
      value: Fn.join(', ', this.hostedZone.hostedZoneNameServers!),
      description: 'NS records — ドメインレジストラにこの値を設定してください',
    })
  }
}
