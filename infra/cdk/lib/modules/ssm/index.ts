import * as ssm from 'aws-cdk-lib/aws-ssm'
import { Construct } from 'constructs'

export interface SsmParameter {
  id: string
  name: string
  value: string
}

export interface SsmModuleProps {
  parameters: SsmParameter[]
}

export class SsmModule extends Construct {
  public readonly parameters: ssm.StringParameter[]

  constructor(scope: Construct, id: string, props: SsmModuleProps) {
    super(scope, id)

    this.parameters = props.parameters.map(
      (param) =>
        new ssm.StringParameter(this, param.id, {
          parameterName: param.name,
          stringValue: param.value,
        })
    )
  }
}
