package sender

import (
	"github.com/relatedbits/smtp-firewall/model"
)

type Sender interface {
	Send(email *model.Email) error
}
