package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adrianbrad/queue"
	"github.com/emersion/go-smtp"
	"github.com/spf13/viper"
	"github.com/zekroTJA/timedmap"
	"go.uber.org/ratelimit"

	smtpfirewall "github.com/relatedbits/smtp-firewall"
	"github.com/relatedbits/smtp-firewall/filter"
	"github.com/relatedbits/smtp-firewall/mailqueue"
	"github.com/relatedbits/smtp-firewall/model"
	"github.com/relatedbits/smtp-firewall/sender"
)

func main() {
	bkd := &smtpfirewall.Backend{
		Filters: newFilters(config),
		MailQueue: mailqueue.NewRatelimitQueue(
			newSender(config),
			queue.NewBlocking[*model.Email]([]*model.Email{}, queue.WithCapacity(config.GetInt("queue.capacity"))),
			ratelimit.New(config.GetInt("ratelimit")),
			timedmap.New(time.Duration(config.GetInt64("timedmap.cleanup_per_n_second"))*time.Second),
			int16(config.GetInt64("timedmap.cooldown_seconds_for_single_recipient")),
			os.Stdout,
		),
	}
	bkd.MailQueue.Serve()

	s := smtp.NewServer(bkd)

	s.Addr = config.GetString("smtp_server.addr")
	s.AllowInsecureAuth = config.GetBool("smtp_server.allow_insecure_auth")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Println("Starting SMTP server at", s.Addr)
		if err := s.ListenAndServe(); err != nil && err != smtp.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutdown Server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		fmt.Println("SMTP Server Shutdown:", err)
	}
	bkd.MailQueue.Shutdown(ctx)
}

func newFilters(config *viper.Viper) *[]filter.Filter {
	output := []filter.Filter{}

	if config.GetBool("filter.bad_domain.enabled") {
		bd := filter.NewBadDomainFilter(readLinesFromFile("bad_domains.txt"))
		output = append(output, bd)
	}

	return &output
}

func readLinesFromFile(filename string) []string {
	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var output []string
	s := bufio.NewScanner(f)
	for s.Scan() {
		if t := s.Text(); t != "" {
			output = append(output, t)
		}
	}
	if err = s.Err(); err != nil {
		log.Fatal(err)
	}

	return output
}

func newSender(config *viper.Viper) (output sender.Sender) {
	senderType := config.GetString("sender.type")

	switch senderType {
	case "awsses":
		output = sender.NewAWSSESSender(int16(config.GetInt(("sender.awsses.timeout"))))

	case "mailjet":
		baseurl := config.GetString("sender.mailjet.baseurl")
		keyPub := config.GetString("sender.mailjet.apikey_public")
		keyPri := config.GetString("sender.mailjet.apikey_private")

		output = sender.NewMailjetSender(keyPub, keyPri, baseurl)

	case "smtp":
		output = sender.NewSMTPSender(config.GetString("sender.smtp.addr"), nil)

	default:
		log.Fatal("unknown Sender Type:", senderType)
	}
	return
}
