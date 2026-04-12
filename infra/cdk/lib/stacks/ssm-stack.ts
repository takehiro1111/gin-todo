import * as rds from 'aws-cdk-lib/aws-rds'
import * as ssm from 'aws-cdk-lib/aws-ssm'
import { Stack, StackProps } from 'aws-cdk-lib'
import { Construct } from 'constructs'
import { ENV } from '@/env/env.js'

export interface SsmStackProps extends StackProps {
  rdsInstance: rds.IDatabaseInstance
}

// 本番用 SSM パラメータ (RDS エンドポイントを動的に設定)
// ローカル開発用は lib/constructs/ssm-local.ts を参照
export class SsmStack extends Stack {
  constructor(scope: Construct, id: string, props: SsmStackProps) {
    super(scope, id, props)

    const params = [
      {
        id: 'PostgresUser',
        name: `${ENV.ssmParameterPrefix}/db/postgres-user`,
        value: ENV.dbUser,
      },
      {
        id: 'PostgresPassword',
        name: `${ENV.ssmParameterPrefix}/db/postgres-password`,
        value: ENV.postgresPassword,
      },
      {
        id: 'PostgresDbName',
        name: `${ENV.ssmParameterPrefix}/db/postgres-db-name`,
        value: ENV.dbName,
      },
      {
        id: 'PostgresDbHost',
        name: `${ENV.ssmParameterPrefix}/db/postgres-db-host`,
        // RDS エンドポイントを動的に取得
        value: props.rdsInstance.dbInstanceEndpointAddress,
      },
      {
        id: 'PostgresDbPort',
        name: `${ENV.ssmParameterPrefix}/db/postgres-db-port`,
        value: String(ENV.dbPort),
      },
      {
        id: 'PostgresSslMode',
        name: `${ENV.ssmParameterPrefix}/db/postgres-ssl-mode`,
        value: 'require',
      },
      {
        id: 'JwtSecretKey',
        name: `${ENV.ssmParameterPrefix}/db/jwt-secret-key`,
        value: ENV.jwtSecretKey,
      },
    ]

    for (const param of params) {
      new ssm.StringParameter(this, param.id, {
        parameterName: param.name,
        stringValue: param.value,
      })
    }
  }
}
