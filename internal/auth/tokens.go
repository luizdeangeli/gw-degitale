package auth

import (
	"crypto/subtle"
	"fmt"
	"strings"
)

type token struct {
	name   string
	secret string
}

// Tokens autentica chamadas pelo segredo e devolve o nome do sistema para o log.
type Tokens struct {
	list []token
}

func Parse(raw string) (*Tokens, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, fmt.Errorf("GW_TOKENS vazio")
	}

	var list []token
	seenName := map[string]struct{}{}
	seenSecret := map[string]struct{}{}

	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, secret, ok := strings.Cut(part, ":")
		name = strings.TrimSpace(name)
		secret = strings.TrimSpace(secret)
		if !ok || name == "" || secret == "" {
			return nil, fmt.Errorf("token invalido em GW_TOKENS (use nome:segredo)")
		}
		if _, dup := seenName[name]; dup {
			return nil, fmt.Errorf("nome de token duplicado: %s", name)
		}
		if _, dup := seenSecret[secret]; dup {
			return nil, fmt.Errorf("segredo de token duplicado")
		}
		seenName[name] = struct{}{}
		seenSecret[secret] = struct{}{}
		list = append(list, token{name: name, secret: secret})
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("GW_TOKENS vazio")
	}
	return &Tokens{list: list}, nil
}

// Name percorre todos os tokens de propósito, para o tempo não revelar qual casou.
func (t *Tokens) Name(secret string) (string, bool) {
	if t == nil || secret == "" {
		return "", false
	}
	in := []byte(secret)
	var found string
	ok := false
	for _, item := range t.list {
		if subtle.ConstantTimeCompare([]byte(item.secret), in) == 1 {
			found = item.name
			ok = true
		}
	}
	return found, ok
}
