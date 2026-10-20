package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"gw.degitale.com.br/internal/auth"
	"gw.degitale.com.br/internal/cache"
	"gw.degitale.com.br/internal/doc"
	"gw.degitale.com.br/internal/lookup"
)

type clientKey struct{}

type Server struct {
	Auth    *auth.Tokens
	Cnpj    *lookup.Chain
	Cep     *lookup.Chain
	Cache   *cache.Memory
	TTLCnpj time.Duration
	TTLCep  time.Duration
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /v1/cnpj/{cnpj}", s.auth(s.cnpj))
	mux.HandleFunc("GET /v1/cep/{cep}", s.auth(s.cep))
	return mux
}

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthBody{Ok: true})
}

func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		secret, ok := bearer(r.Header.Get("Authorization"))
		if !ok {
			writeError(w, http.StatusUnauthorized, "nao_autorizado")
			return
		}
		name, ok := s.Auth.Name(secret)
		if !ok {
			writeError(w, http.StatusUnauthorized, "nao_autorizado")
			return
		}
		ctx := context.WithValue(r.Context(), clientKey{}, name)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) cnpj(w http.ResponseWriter, r *http.Request) {
	digits, err := doc.CNPJ(r.PathValue("cnpj"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "documento_invalido")
		return
	}
	s.consult(w, r, "cnpj", "cnpj:"+digits, s.TTLCnpj, s.Cnpj, digits)
}

func (s *Server) cep(w http.ResponseWriter, r *http.Request) {
	digits, err := doc.CEP(r.PathValue("cep"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "documento_invalido")
		return
	}
	s.consult(w, r, "cep", "cep:"+digits, s.TTLCep, s.Cep, digits)
}

func (s *Server) consult(w http.ResponseWriter, r *http.Request, tipo, cacheKey string, ttl time.Duration, chain *lookup.Chain, digits string) {
	client := clientName(r.Context())
	if providerName, data, ok := s.Cache.Get(cacheKey); ok {
		log.Printf("consulta tipo=%s client=%s resultado=ok provider=%s cached=true", tipo, client, providerName)
		writeOK(w, providerName, true, data)
		return
	}

	res, err := chain.Run(r.Context(), digits)
	for _, attempt := range res.Attempts {
		log.Printf("fallback tipo=%s client=%s provider=%s status=%s", tipo, client, attempt.Provider, attempt.Status)
	}
	if err == nil {
		s.Cache.Set(cacheKey, res.Provider, res.Data, ttl)
		log.Printf("consulta tipo=%s client=%s resultado=ok provider=%s cached=false", tipo, client, res.Provider)
		writeOK(w, res.Provider, false, res.Data)
		return
	}
	if errors.Is(err, lookup.ErrNotFound) {
		log.Printf("consulta tipo=%s client=%s resultado=nao_encontrado", tipo, client)
		writeError(w, http.StatusNotFound, "nao_encontrado")
		return
	}
	log.Printf("consulta tipo=%s client=%s resultado=provedores_indisponiveis", tipo, client)
	writeJSON(w, http.StatusServiceUnavailable, exhaustedBody{
		Ok:         false,
		Error:      "provedores_indisponiveis",
		Tentativas: res.Attempts,
	})
}

func bearer(header string) (string, bool) {
	const prefix = "Bearer "
	if len(header) <= len(prefix) || !strings.EqualFold(header[:len(prefix)], prefix) {
		return "", false
	}
	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", false
	}
	return token, true
}

func clientName(ctx context.Context) string {
	name, _ := ctx.Value(clientKey{}).(string)
	if name == "" {
		return "-"
	}
	return name
}

type healthBody struct {
	Ok bool `json:"ok"`
}

type okBody struct {
	Ok       bool   `json:"ok"`
	Provider string `json:"provider"`
	Cached   bool   `json:"cached"`
	Data     any    `json:"data"`
}

type errBody struct {
	Ok    bool   `json:"ok"`
	Error string `json:"error"`
}

type exhaustedBody struct {
	Ok         bool             `json:"ok"`
	Error      string           `json:"error"`
	Tentativas []lookup.Attempt `json:"tentativas"`
}

func writeOK(w http.ResponseWriter, providerName string, cached bool, data any) {
	writeJSON(w, http.StatusOK, okBody{
		Ok:       true,
		Provider: providerName,
		Cached:   cached,
		Data:     data,
	})
}

func writeError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, errBody{Ok: false, Error: code})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(v)
}
