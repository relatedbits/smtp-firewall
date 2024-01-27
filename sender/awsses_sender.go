package sender

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"

	"github.com/relatedbits/smtp-firewall/model"
)

type AWSSESSender struct {
	Timeout int16
	client  *sesv2.Client
}

func NewAWSSESSender(timeout int16) *AWSSESSender {
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	return &AWSSESSender{
		Timeout: timeout,
		client:  sesv2.NewFromConfig(cfg),
	}
}

func (s *AWSSESSender) Send(email *model.Email) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(s.Timeout)*time.Second)
	defer cancel()

	_, err := s.client.SendEmail(ctx, &sesv2.SendEmailInput{
		Content: &types.EmailContent{
			Raw: &types.RawMessage{Data: email.Data},
		},
	})
	return err
}
