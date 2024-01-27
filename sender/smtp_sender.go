package sender

import (
	"bytes"

	"github.com/emersion/go-sasl"
	"github.com/emersion/go-smtp"

	"github.com/relatedbits/smtp-firewall/model"
)

type SMTPSender struct {
	Addr   string
	client *sasl.Client
}

func NewSMTPSender(addr string, client *sasl.Client) *SMTPSender {
	return &SMTPSender{
		Addr:   addr,
		client: client,
	}
}

func (s *SMTPSender) Send(email *model.Email) error {
	if s.client == nil {
		c, err := smtp.Dial(s.Addr)
		if err != nil {
			return err
		}
		return c.SendMail(email.From, email.To, bytes.NewReader(email.Data))
	} else {
		return smtp.SendMailTLS(s.Addr, *s.client, email.From, email.To, bytes.NewReader(email.Data))
	}
}
