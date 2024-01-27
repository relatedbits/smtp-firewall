package filter

import "github.com/relatedbits/smtp-firewall/model"

type Filter interface {
	CanSend(email *model.Email) bool
}
