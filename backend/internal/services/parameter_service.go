package services

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func GetDBAuthenticate(ssmClient *ssm.Client, ctx context.Context) (*ssm.GetParametersOutput, error) {
	paramInput := &ssm.GetParametersInput{
		Names: []string{
			PostgresUser,
			PostgresPassword,
			PostgresDBName,
			PostgresDBHost,
			PostgresDBPort,
			PostgresSslMode,
		},
		WithDecryption: aws.Bool(false), // SecureStringは使用してない。
	}

	result, err := ssmClient.GetParameters(ctx, paramInput)
	if err != nil {
		return nil, err
	}

	if len(result.InvalidParameters) > 0 {
		return nil, fmt.Errorf("parameters not found in SSM: %v", result.InvalidParameters)
	}

	return result, nil
}
