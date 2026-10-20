package provider

import (
	"context"
	"errors"
	"io"
	"net/http"
)

const maxBody = 2 << 20

// Do executa a requisição e devolve status e corpo.
// Erro de rede ou timeout vira *Error temporário.
func Do(ctx context.Context, client *http.Client, req *http.Request) (int, []byte, error) {
	resp, err := client.Do(req.WithContext(ctx))
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return 0, nil, &Error{Kind: KindTemporary, Detail: "timeout"}
		}
		return 0, nil, &Error{Kind: KindTemporary, Detail: "erro_rede"}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return resp.StatusCode, nil, &Error{Kind: KindTemporary, Status: resp.StatusCode, Detail: "corpo"}
	}
	return resp.StatusCode, body, nil
}

// Classify transforma o HTTP status do provedor na decisão da cadeia.
func Classify(status int) error {
	switch {
	case status == http.StatusOK:
		return nil
	case status == http.StatusNotFound:
		return &Error{Kind: KindNotFound, Status: status}
	case status == http.StatusTooManyRequests:
		return &Error{Kind: KindRateLimit, Status: status}
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		return &Error{Kind: KindAuth, Status: status}
	default:
		return &Error{Kind: KindTemporary, Status: status}
	}
}
