package mailqueue

import (
	"context"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/adrianbrad/queue"
	"github.com/zekroTJA/timedmap"
	"go.uber.org/ratelimit"

	"github.com/relatedbits/smtp-firewall/model"
	"github.com/relatedbits/smtp-firewall/sender"
)

type RatelimitQueue struct {
	Sender                     sender.Sender
	Queue                      *queue.Blocking[*model.Email]
	Ratelimit                  ratelimit.Limiter
	Timedmap                   *timedmap.TimedMap
	CooldownForSingleRecipient int16

	Debug     io.Writer
	done      chan struct{}
	queueDone chan struct{}
}

func NewRatelimitQueue(sender sender.Sender, queue *queue.Blocking[*model.Email], ratelimit ratelimit.Limiter, timedmap *timedmap.TimedMap, cooldownForSingleRecipient int16, debug io.Writer) *RatelimitQueue {
	return &RatelimitQueue{
		Sender:                     sender,
		Queue:                      queue,
		Ratelimit:                  ratelimit,
		Timedmap:                   timedmap,
		CooldownForSingleRecipient: cooldownForSingleRecipient,

		Debug:     debug,
		done:      make(chan struct{}, 1),
		queueDone: make(chan struct{}, 1),
	}
}

func (rq *RatelimitQueue) Append(email *model.Email) {
	go func(email *model.Email) {
		rq.Queue.OfferWait(email)
	}(email)
}

func (rq *RatelimitQueue) Serve() {
	go func() {
		for {
			select {
			case <-rq.done:
				if rq.Queue.IsEmpty() {
					close(rq.queueDone)
					return
				}
			default:
			}

			rq.Ratelimit.Take()
			email := rq.Queue.GetWait()

			if email == nil {
				continue
			}

			send := true
			for _, v := range email.To {
				if rq.Timedmap.GetValue(v) != nil {
					// Discard the Email
					send = false
					continue
				}
			}
			if !send {
				continue
			}

			if err := rq.Sender.Send(email); err != nil {
				fmt.Fprintln(rq.Debug, err)
				continue
			}

			for _, v := range email.To {
				rq.Timedmap.Set(v, true, time.Duration(rq.CooldownForSingleRecipient)*time.Second)
			}
		}
	}()
}

func (rq *RatelimitQueue) Shutdown(ctx context.Context) {
	select {
	case <-rq.done:
		return
	default:
		close(rq.done)
	}

	rq.Append(nil) // To end the `queue.GetWait()`

	select {
	case <-ctx.Done():
		log.Fatal(ctx.Err())
		return
	case <-rq.queueDone:
		return
	}
}
