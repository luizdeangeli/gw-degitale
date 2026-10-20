package provider

import (
	"context"
	"fmt"
)

type Kind int

const (
	KindNotFound Kind = iota + 1
	KindRateLimit
	KindAuth
	KindTemporary
)

// Error classifica a falha de um provedor para a cadeia decidir se para ou segue.
type Error struct {
	Kind   Kind
	Status int
	Detail string
}

func (e *Error) Error() string {
	if e.Status != 0 {
		return fmt.Sprintf("provedor status %d", e.Status)
	}
	if e.Detail != "" {
		return e.Detail
	}
	return "falha de provedor"
}

func IsKind(err error, kind Kind) bool {
	pe, ok := err.(*Error)
	return ok && pe.Kind == kind
}

func StatusLabel(err error) string {
	pe, ok := err.(*Error)
	if !ok {
		return "erro"
	}
	if pe.Status != 0 {
		return fmt.Sprintf("%d", pe.Status)
	}
	if pe.Detail != "" {
		return pe.Detail
	}
	return "erro"
}

// Lookup consulta um provedor. key já vem só com dígitos.
type Lookup interface {
	Name() string
	Lookup(ctx context.Context, key string) (any, error)
}
