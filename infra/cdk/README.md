# Welcome to your CDK TypeScript project

This is a blank project for CDK development with TypeScript.

The `cdk.json` file tells the CDK Toolkit how to execute your app.

## Useful commands

* `npm run build`   compile typescript to js
* `npm run watch`   watch for changes and compile
* `npm run test`    perform the jest unit tests
* `npx cdk deploy`  deploy this stack to your default AWS account/region
* `npx cdk diff`    compare deployed stack with current state
* `npx cdk synth`   emits the synthesized CloudFormation template

## cdk commands
```zsh
# cdkのデプロイに必要なIAMロールやS3を作成
$ cdk bootstrap
[WARNING] aws-cdk-lib.aws_ssm.StringParameterProps#type is deprecated.
  - type will always be 'String'
  This API will be removed in the next major release.
[WARNING] aws-cdk-lib.aws_ssm.ParameterType is deprecated.
  these types are no longer used
  This API will be removed in the next major release.
[WARNING] aws-cdk-lib.aws_ssm.ParameterType#SECURE_STRING is deprecated.

  This API will be removed in the next major release.
 ⏳  Bootstrapping environment aws://650251692423/us-east-1...
Trusted accounts for deployment: (none)
Trusted accounts for lookup: (none)
Using default execution policy of 'arn:aws:iam::aws:policy/AdministratorAccess'. Pass '--cloudformation-execution-policies' to customize.
CDKToolkit: creating CloudFormation changeset...
 ✅  Environment aws://650251692423/us-east-1 bootstrapped.


```

## delete stack
```zsh
aws cloudformation delete-stack --stack-name GinTodoSsmStack
```
