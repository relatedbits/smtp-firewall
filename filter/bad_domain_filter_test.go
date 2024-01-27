package filter

import (
	"reflect"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	"github.com/stretchr/testify/require"

	"github.com/relatedbits/smtp-firewall/model"
)

func TestNewBadDomainFilter(t *testing.T) {
	type args struct {
		domains []string
	}
	scenarios := []struct {
		args args
		want *BadDomainFilter
	}{
		{
			args: args{
				domains: []string{"blocked.local"},
			},
			want: &BadDomainFilter{
				blacklist: mapset.NewSet[string]("blocked.local"),
			},
		},
		{
			args: args{},
			want: &BadDomainFilter{
				blacklist: mapset.NewSet[string](),
			},
		},
	}
	for _, s := range scenarios {
		if got := NewBadDomainFilter(s.args.domains); !reflect.DeepEqual(got, s.want) {
			require.Equal(t, s.want, got)
		}
	}
}

func TestBadDomainFilter_CanSend(t *testing.T) {
	type fields struct {
		blacklist mapset.Set[string]
	}
	type args struct {
		email *model.Email
	}
	scenarios := []struct {
		fields fields
		args   args
		want   bool
	}{
		{
			fields: fields{
				blacklist: mapset.NewSet[string]("blocked.local"),
			},
			args: args{
				email: &model.Email{
					To: []string{"alice@blocked.local"},
				},
			},
			want: false,
		},
		{
			fields: fields{
				blacklist: mapset.NewSet[string]("blocked.local"),
			},
			args: args{
				email: &model.Email{
					To: []string{"alice@pass.local"},
				},
			},
			want: true,
		},
		{
			fields: fields{
				blacklist: mapset.NewSet[string](),
			},
			args: args{
				email: &model.Email{
					To: []string{"alice@pass.local"},
				},
			},
			want: true,
		},
	}
	for _, s := range scenarios {
		b := &BadDomainFilter{
			blacklist: s.fields.blacklist,
		}
		if got := b.CanSend(s.args.email); got != s.want {
			require.Equal(t, s.want, got, "blacklist %s, to %s", s.fields.blacklist.ToSlice(), s.args.email.To)
		}
	}
}
