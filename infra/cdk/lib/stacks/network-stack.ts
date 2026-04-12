import { Stack, StackProps } from 'aws-cdk-lib'
import { Construct } from 'constructs'
import { VpcModule } from '@/lib/modules/vpc/index.js'
import { ENV } from '@/env/env.js'

export class NetworkStack extends Stack {
  public readonly vpcModule: VpcModule

  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props)

    this.vpcModule = new VpcModule(this, 'Vpc', {
      vpcCidr: ENV.vpcCidr,
      maxAzs: ENV.maxAzs,
      natGateways: ENV.natGateways,
      publicSubnetCidrMask: ENV.publicSubnetCidrMask,
      privateSubnetCidrMask: ENV.privateSubnetCidrMask,
    })
  }
}
