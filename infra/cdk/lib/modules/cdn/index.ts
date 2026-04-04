import * as acm from 'aws-cdk-lib/aws-certificatemanager'
import * as cloudfront from 'aws-cdk-lib/aws-cloudfront'
import * as origins from 'aws-cdk-lib/aws-cloudfront-origins'
import * as elbv2 from 'aws-cdk-lib/aws-elasticloadbalancingv2'
import * as route53 from 'aws-cdk-lib/aws-route53'
import * as route53Targets from 'aws-cdk-lib/aws-route53-targets'
import * as s3 from 'aws-cdk-lib/aws-s3'
import { Duration } from 'aws-cdk-lib'
import { Construct } from 'constructs'

export interface CdnModuleProps {
  frontendBucket: s3.IBucket
  alb: elbv2.IApplicationLoadBalancer
  domainName: string
  appDomain: string
  certificate: acm.ICertificate
  hostedZone: route53.IHostedZone
}

export class CdnModule extends Construct {
  public readonly distribution: cloudfront.Distribution

  constructor(scope: Construct, id: string, props: CdnModuleProps) {
    super(scope, id)

    const oac = new cloudfront.S3OriginAccessControl(this, 'OAC', {
      signing: cloudfront.Signing.SIGV4_NO_OVERRIDE,
    })

    const s3Origin = origins.S3BucketOrigin.withOriginAccessControl(
      props.frontendBucket as s3.Bucket,
      { originAccessControl: oac }
    )

    // VPC Origin: CloudFront → internal ALB (PrivateLink 経由)
    // デプロイ後に CloudFront が VPC Origin 用の SG を自動作成する
    const vpcOrigin = origins.VpcOrigin.withApplicationLoadBalancer(props.alb, {
      protocolPolicy: cloudfront.OriginProtocolPolicy.HTTP_ONLY,
      httpPort: 80,
    })

    this.distribution = new cloudfront.Distribution(this, 'Distribution', {
      domainNames: [props.appDomain],
      certificate: props.certificate,
      defaultBehavior: {
        origin: s3Origin,
        viewerProtocolPolicy:
          cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
        cachePolicy: cloudfront.CachePolicy.CACHING_OPTIMIZED,
      },
      additionalBehaviors: {
        '/api/ws/*': {
          origin: vpcOrigin,
          viewerProtocolPolicy:
            cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
          allowedMethods: cloudfront.AllowedMethods.ALLOW_ALL,
          cachePolicy: cloudfront.CachePolicy.CACHING_DISABLED,
          originRequestPolicy:
            cloudfront.OriginRequestPolicy.ALL_VIEWER_EXCEPT_HOST_HEADER,
        },
        '/api/*': {
          origin: vpcOrigin,
          viewerProtocolPolicy:
            cloudfront.ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
          allowedMethods: cloudfront.AllowedMethods.ALLOW_ALL,
          cachePolicy: cloudfront.CachePolicy.CACHING_DISABLED,
          originRequestPolicy:
            cloudfront.OriginRequestPolicy.ALL_VIEWER_EXCEPT_HOST_HEADER,
        },
      },
      defaultRootObject: 'index.html',
      errorResponses: [
        {
          httpStatus: 403,
          responseHttpStatus: 200,
          responsePagePath: '/index.html',
          ttl: Duration.minutes(5),
        },
        {
          httpStatus: 404,
          responseHttpStatus: 200,
          responsePagePath: '/index.html',
          ttl: Duration.minutes(5),
        },
      ],
    })

    // Route53: todo.takehiro1111.com → CloudFront
    new route53.ARecord(this, 'AliasRecord', {
      zone: props.hostedZone,
      recordName: props.appDomain,
      target: route53.RecordTarget.fromAlias(
        new route53Targets.CloudFrontTarget(this.distribution)
      ),
    })

    new route53.AaaaRecord(this, 'AliasRecordAAAA', {
      zone: props.hostedZone,
      recordName: props.appDomain,
      target: route53.RecordTarget.fromAlias(
        new route53Targets.CloudFrontTarget(this.distribution)
      ),
    })
  }
}
