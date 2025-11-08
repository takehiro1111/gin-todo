#!/usr/bin/env node
import * as cdk from 'aws-cdk-lib';
import { SsmParameterStack } from '../lib/constructs/ssm-parameter';
import { ENV } from "../env/env";

const app = new cdk.App();

// GinTodoSsmStack = CloudFormationのスタック名
new SsmParameterStack(app, 'GinTodoSsmStack', {
  env: {
    account: ENV.accountID,
    region: ENV.defaultRegion,
  }
});
