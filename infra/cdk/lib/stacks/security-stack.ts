import { Stack, StackProps } from 'aws-cdk-lib'
import { Construct } from 'constructs'
import { VpcModule } from '@/lib/modules/vpc/index.js'
import { SecurityModule } from '@/lib/modules/security/index.js'
import { ENV } from '@/env/env.js'

export interface SecurityStackProps extends StackProps {
  vpcModule: VpcModule
}

export class SecurityStack extends Stack {
  public readonly securityModule: SecurityModule

  constructor(scope: Construct, id: string, props: SecurityStackProps) {
    super(scope, id, props)

    this.securityModule = new SecurityModule(this, 'Security', {
      vpc: props.vpcModule.vpc,
      dbPort: ENV.dbPort,
    })
  }
}
