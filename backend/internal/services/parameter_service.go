package services

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func GetParameters(ssmClient *ssm.Client, ctx context.Context) (*ssm.GetParametersOutput, error) {
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

	return result, nil
}
