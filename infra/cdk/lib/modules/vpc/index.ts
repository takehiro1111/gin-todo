import * as ec2 from 'aws-cdk-lib/aws-ec2'
import { Construct } from 'constructs'

export interface VpcModuleProps {
  vpcCidr: string
  maxAzs: number
  natGateways: number
  publicSubnetCidrMask: number
  privateSubnetCidrMask: number
}

export class VpcModule extends Construct {
  public readonly vpc: ec2.Vpc
  public readonly publicSubnets: ec2.ISubnet[]
  public readonly privateSubnets: ec2.ISubnet[]

  constructor(scope: Construct, id: string, props: VpcModuleProps) {
    super(scope, id)

    this.vpc = new ec2.Vpc(this, 'Vpc', {
      ipAddresses: ec2.IpAddresses.cidr(props.vpcCidr),
      maxAzs: props.maxAzs,
      natGateways: props.natGateways,
      subnetConfiguration: [
        {
          cidrMask: props.publicSubnetCidrMask,
          name: 'Public',
          subnetType: ec2.SubnetType.PUBLIC,
        },
        {
          cidrMask: props.privateSubnetCidrMask,
          name: 'Private',
          subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
        },
      ],
    })

    this.publicSubnets = this.vpc.publicSubnets
    this.privateSubnets = this.vpc.privateSubnets
  }
}
