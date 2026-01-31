package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// NewAWSConfigはregionとendpointを受け取り、AWS configを返します。
// endpointが空の場合はAWSデフォルトエンドポイントを使用します。
func NewAWSConfig(ctx context.Context, region, endpoint string) (aws.Config, error) {
	opts := []func(*config.LoadOptions) error{
		config.WithRegion(region),
	}
	if endpoint != "" {
		opts = append(opts, config.WithBaseEndpoint(endpoint))
	}
	return config.LoadDefaultConfig(ctx, opts...)
}

// NewSESConfigは環境に応じたSES用のAWS configを返します。
// production/staging以外はローカル開発用エンドポイントを使用します。
func NewSESConfig(ctx context.Context, env string) (aws.Config, error) {
	endpoint := ""
	if env != "production" && env != "staging" {
		endpoint = "http://localhost:8005"
	}
	return NewAWSConfig(ctx, "ap-northeast-1", endpoint)
}
