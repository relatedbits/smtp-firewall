package smtpfirewall

import (
	"io"

	"github.com/emersion/go-smtp"

	"github.com/relatedbits/smtp-firewall/filter"
	"github.com/relatedbits/smtp-firewall/mailqueue"
	"github.com/relatedbits/smtp-firewall/model"
)

type Backend struct {
	Filters   *[]filter.Filter
	MailQueue mailqueue.MailQueue
}

func (bkd *Backend) NewSession(c *smtp.Conn) (smtp.Session, error) {
	return &session{
		Emain: &model.Email{},

		Filters:   *bkd.Filters,
		MailQueue: bkd.MailQueue,
	}, nil
}

func (bkd *Backend) Shutdown() {
}

type session struct {
	Emain *model.Email

	Filters   []filter.Filter
	MailQueue mailqueue.MailQueue
}

func (s *session) AuthPlain(username, password string) error {
	return nil
}

func (s *session) Mail(from string, opts *smtp.MailOptions) error {
	s.Emain.From = from
	return nil
}

func (s *session) Rcpt(to string, opts *smtp.RcptOptions) error {
	s.Emain.To = append(s.Emain.To, to)
	return nil
}

func (s *session) Data(r io.Reader) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	s.Emain.Data = data

	// Discard if cannot send
	if s.CanSend() {
		s.MailQueue.Append(s.Emain)
	}

	return nil
}

func (s *session) Reset() {
	s.Emain = &model.Email{}
}

func (s *session) Logout() error {
	return nil
}

func (s *session) CanSend() bool {
	output := true
	for _, filter := range s.Filters {
		output = filter.CanSend(s.Emain)
	}
	return output
}
