package lookup

import (
	"context"
	"errors"

	"gw.degitale.com.br/internal/provider"
)

var (
	ErrNotFound  = errors.New("nao encontrado")
	ErrExhausted = errors.New("provedores indisponiveis")
)

type Attempt struct {
	Provider string `json:"provider"`
	Status   string `json:"status"`
}

type Result struct {
	Provider string
	Data     any
	Attempts []Attempt
}

type Chain struct {
	providers []provider.Lookup
}

func New(providers []provider.Lookup) *Chain {
	return &Chain{providers: providers}
}

func (c *Chain) Names() []string {
	names := make([]string, 0, len(c.providers))
	for _, p := range c.providers {
		names = append(names, p.Name())
	}
	return names
}

// Run percorre a cadeia. 404 encerra. 429, 5xx, timeout, rede e 401/403 seguem para o próximo.
func (c *Chain) Run(ctx context.Context, key string) (Result, error) {
	attempts := []Attempt{}
	for _, p := range c.providers {
		data, err := p.Lookup(ctx, key)
		if err == nil {
			return Result{Provider: p.Name(), Data: data, Attempts: attempts}, nil
		}
		if provider.IsKind(err, provider.KindNotFound) {
			return Result{}, ErrNotFound
		}
		attempts = append(attempts, Attempt{
			Provider: p.Name(),
			Status:   provider.StatusLabel(err),
		})
		if ctx.Err() != nil {
			break
		}
	}
	return Result{Attempts: attempts}, ErrExhausted
}
