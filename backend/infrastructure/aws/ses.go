package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

type Mail interface {
	SendEmail(ctx context.Context, title, content, to, from string) error
}

type mail struct {
	client *sesv2.Client
}

func NewMail(cfg aws.Config) Mail {
	return &mail{
		client: sesv2.NewFromConfig(cfg),
	}
}

func (m *mail) SendEmail(ctx context.Context, title string, content string, to string, from string) error {
	_, err := m.client.SendEmail(ctx, &sesv2.SendEmailInput{
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		FromEmailAddress: aws.String(from),
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data: aws.String(title),
				},
				Body: &types.Body{
					Text: &types.Content{
						Data: aws.String(content),
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
