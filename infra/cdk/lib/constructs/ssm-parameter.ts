import * as ssm from "aws-cdk-lib/aws-ssm";
import { Stack, StackProps } from "aws-cdk-lib";
import { Construct } from "constructs";
import { ENV } from "../../env/env";

export class SsmParameterStack extends Stack {
  constructor(scope: Construct, id: string, props?: StackProps) {
    super(scope, id, props);

    const stringParams = [
      {
        id: "PostgresUser",
        name: "/gin-todo/db/postgres-user",
        value: "gin",
      },
      {
        id: "PostgresPassword",
        name: "/gin-todo/db/postgres-password",
        value: ENV.postgresPassword,
      },
      {
        id: "PostgresDbName",
        name: "/gin-todo/db/postgres-db-name",
        value: "gin-todo",
      },
      {
        id: "PostgresDbHost",
        name: "/gin-todo/db/postgres-db-host",
        value: "localhost",
      },
      {
        id: "PostgresDbPort",
        name: "/gin-todo/db/postgres-db-port",
        value: "5432",
      },
      {
        id: "PostgresSslMode",
        name: "/gin-todo/db/postgres-ssl-mode",
        value: "disable",
      },
      {
        id: "JwtSecretKey",
        name: "/gin-todo/db/jwt-secret-key",
        value: ENV.jwtSecretKey,
      },
    ];

    for (const param of stringParams) {
      new ssm.StringParameter(this, param.id, {
        parameterName: param.name,
        stringValue: param.value,
      });
    }
  }
}
