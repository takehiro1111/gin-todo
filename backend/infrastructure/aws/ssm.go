package aws

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func NewSSMClient() (*ssm.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion(DefaultRegion),
	)
	if err != nil {
		return nil, err
	}

	ssmClient := ssm.NewFromConfig(cfg)

	return ssmClient, nil
}
