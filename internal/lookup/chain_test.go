package lookup

import (
	"context"
	"errors"
	"testing"

	"gw.degitale.com.br/internal/provider"
)

type stub struct {
	name  string
	err   error
	data  any
	calls int
}

func (s *stub) Name() string { return s.name }

func (s *stub) Lookup(context.Context, string) (any, error) {
	s.calls++
	if s.err != nil {
		return nil, s.err
	}
	return s.data, nil
}

func TestFallbackOn429StopsOn404(t *testing.T) {
	first := &stub{name: "cnpjws", err: &provider.Error{Kind: provider.KindRateLimit, Status: 429}}
	second := &stub{name: "cnpja", err: &provider.Error{Kind: provider.KindNotFound, Status: 404}}
	third := &stub{name: "outro", data: "nao deve"}

	_, err := New([]provider.Lookup{first, second, third}).Run(context.Background(), "1")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("err=%v", err)
	}
	if first.calls != 1 || second.calls != 1 || third.calls != 0 {
		t.Fatalf("calls %d %d %d", first.calls, second.calls, third.calls)
	}
}

func TestUsesNextAfterRateLimit(t *testing.T) {
	first := &stub{name: "cnpjws", err: &provider.Error{Kind: provider.KindRateLimit, Status: 429}}
	second := &stub{name: "cnpja", data: "ok"}
	res, err := New([]provider.Lookup{first, second}).Run(context.Background(), "1")
	if err != nil {
		t.Fatal(err)
	}
	if res.Provider != "cnpja" || res.Data != "ok" {
		t.Fatalf("%+v", res)
	}
	if len(res.Attempts) != 1 || res.Attempts[0].Status != "429" {
		t.Fatalf("attempts %+v", res.Attempts)
	}
}

func TestExhausted(t *testing.T) {
	a := &stub{name: "cnpjws", err: &provider.Error{Kind: provider.KindTemporary, Detail: "timeout"}}
	b := &stub{name: "cnpja", err: &provider.Error{Kind: provider.KindAuth, Status: 401}}
	res, err := New([]provider.Lookup{a, b}).Run(context.Background(), "1")
	if !errors.Is(err, ErrExhausted) {
		t.Fatal(err)
	}
	if len(res.Attempts) != 2 || res.Attempts[0].Status != "timeout" || res.Attempts[1].Status != "401" {
		t.Fatalf("%+v", res.Attempts)
	}
}
