package mailqueue

import (
	"context"

	"github.com/relatedbits/smtp-firewall/model"
)

type MailQueue interface {
	Append(email *model.Email)
	Serve()
	Shutdown(ctx context.Context)
}
