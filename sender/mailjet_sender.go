package sender

import (
	"bytes"
	"fmt"

	"github.com/DusanKasan/parsemail"
	mailjet "github.com/mailjet/mailjet-apiv3-go"

	"github.com/relatedbits/smtp-firewall/model"
)

type MailjetSender struct {
	client *mailjet.Client
}

func NewMailjetSender(mjApikeyPublic string, mjApikeyPrivate string, baseURL string) *MailjetSender {
	var client *mailjet.Client

	if baseURL != "" {
		client = mailjet.NewMailjetClient(mjApikeyPublic, mjApikeyPrivate, baseURL)
	} else {
		client = mailjet.NewMailjetClient(mjApikeyPublic, mjApikeyPrivate)
	}

	return &MailjetSender{
		client: client,
	}
}

func (s *MailjetSender) Send(email *model.Email) error {
	req, err := s.emailToMJMessagesV31(email)
	if err != nil {
		return err
	}

	_, err = s.client.SendMailV31(req)
	return err
}

func (s *MailjetSender) emailToMJMessagesV31(email *model.Email) (*mailjet.MessagesV31, error) {
	e, err := parsemail.Parse((bytes.NewReader(email.Data)))
	if err != nil {
		return nil, err
	}

	from := e.From[0]
	if from == nil {
		return nil, fmt.Errorf("no From address")
	}

	var to mailjet.RecipientsV31
	for _, a := range e.To {
		to = append(to, mailjet.RecipientV31{
			Name:  a.Name,
			Email: a.Address,
		})
	}

	info := mailjet.InfoMessagesV31{
		From: &mailjet.RecipientV31{
			Email: from.Address,
			Name:  from.Name,
		},
		To:       &to,
		Subject:  e.Subject,
		TextPart: e.TextBody,
		HTMLPart: e.HTMLBody,
	}

	msg := &mailjet.MessagesV31{
		Info: []mailjet.InfoMessagesV31{info},
	}

	return msg, nil
}
