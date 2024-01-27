package filter

import (
	"strings"

	mapset "github.com/deckarep/golang-set/v2"

	"github.com/relatedbits/smtp-firewall/model"
)

type BadDomainFilter struct {
	blacklist mapset.Set[string]
}

func NewBadDomainFilter(domains []string) *BadDomainFilter {
	output := &BadDomainFilter{
		blacklist: mapset.NewSet[string](domains...),
	}
	return output
}

func (b *BadDomainFilter) CanSend(email *model.Email) bool {
	for _, v := range email.To {
		// Assume `email.To` contains a valid Email address
		to := strings.Split(v, "@")
		domain := to[len(to)-1]
		if domain == "" || b.blacklist.ContainsOne(domain) {
			return false
		}
	}

	return true
}
