package httpapi

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gw.degitale.com.br/internal/auth"
	"gw.degitale.com.br/internal/cache"
	"gw.degitale.com.br/internal/lookup"
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

func newServer(cnpj, cep provider.Lookup) *Server {
	tokens, err := auth.Parse("smg:segredo")
	if err != nil {
		panic(err)
	}
	return &Server{
		Auth:    tokens,
		Cnpj:    lookup.New([]provider.Lookup{cnpj}),
		Cep:     lookup.New([]provider.Lookup{cep}),
		Cache:   cache.New(),
		TTLCnpj: time.Hour,
		TTLCep:  time.Hour,
	}
}

func get(t *testing.T, h http.Handler, path, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestDocsPublic(t *testing.T) {
	h := newServer(&stub{name: "cnpjws"}, &stub{name: "viacep"}).Handler()
	rec := get(t, h, "/docs", "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "redoc") {
		t.Fatal(rec.Code, rec.Body.String())
	}
	rec = get(t, h, "/openapi.yaml", "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), "/v1/cnpj/{cnpj}") {
		t.Fatal(rec.Code, rec.Body.String())
	}
}

func TestHealthPublic(t *testing.T) {
	h := newServer(&stub{name: "cnpjws"}, &stub{name: "viacep"}).Handler()
	rec := get(t, h, "/health", "")
	if rec.Code != 200 {
		t.Fatal(rec.Code)
	}
}

func TestAuthAndValidation(t *testing.T) {
	h := newServer(&stub{name: "cnpjws", data: map[string]string{"nmEmpresarial": "X"}}, &stub{name: "viacep"}).Handler()
	if rec := get(t, h, "/v1/cnpj/47960950000121", ""); rec.Code != 401 {
		t.Fatal(rec.Code)
	}
	if rec := get(t, h, "/v1/cnpj/47960950000121", "errado"); rec.Code != 401 {
		t.Fatal(rec.Code)
	}
	if rec := get(t, h, "/v1/cnpj/123", "segredo"); rec.Code != 400 {
		t.Fatal(rec.Code)
	}
	if rec := get(t, h, "/v1/cep/123", "segredo"); rec.Code != 400 {
		t.Fatal(rec.Code)
	}
}

func TestCnpjSuccessAndCache(t *testing.T) {
	prov := &stub{name: "cnpjws", data: map[string]string{"nmEmpresarial": "ACME"}}
	h := newServer(prov, &stub{name: "viacep"}).Handler()
	rec := get(t, h, "/v1/cnpj/47.960.950.0001-21", "segredo")
	if rec.Code != 200 {
		t.Fatal(rec.Body.String())
	}
	body := decode(t, rec)
	if body["provider"] != "cnpjws" || body["cached"] != false {
		t.Fatalf("%v", body)
	}
	rec = get(t, h, "/v1/cnpj/47960950000121", "segredo")
	body = decode(t, rec)
	if body["cached"] != true {
		t.Fatalf("segunda chamada deveria vir do cache: %v", body)
	}
	if prov.calls != 1 {
		t.Fatalf("calls=%d", prov.calls)
	}
}

func TestNotFoundAndExhausted(t *testing.T) {
	h := newServer(
		&stub{name: "cnpjws", err: &provider.Error{Kind: provider.KindNotFound, Status: 404}},
		&stub{name: "viacep", err: &provider.Error{Kind: provider.KindRateLimit, Status: 429}},
	).Handler()
	rec := get(t, h, "/v1/cnpj/47960950000121", "segredo")
	if rec.Code != 404 {
		t.Fatal(rec.Code, rec.Body.String())
	}
	rec = get(t, h, "/v1/cep/01001000", "segredo")
	if rec.Code != 503 {
		t.Fatal(rec.Code, rec.Body.String())
	}
	body := decode(t, rec)
	if body["error"] != "provedores_indisponiveis" {
		t.Fatal(body)
	}
}

func decode(t *testing.T, rec *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	raw, _ := io.ReadAll(rec.Body)
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		t.Fatal(err)
	}
	return body
}
