package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gw.degitale.com.br/internal/auth"
	"gw.degitale.com.br/internal/cache"
	"gw.degitale.com.br/internal/config"
	"gw.degitale.com.br/internal/httpapi"
	"gw.degitale.com.br/internal/lookup"
	"gw.degitale.com.br/internal/provider"
	"gw.degitale.com.br/internal/provider/cep"
	"gw.degitale.com.br/internal/provider/cnpj"
)

func main() {
	log.SetFlags(log.LstdFlags | log.LUTC)
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	tokens, err := auth.Parse(cfg.TokensRaw)
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{
		Timeout: cfg.HTTPTimeout + time.Second,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: cfg.HTTPTimeout,
			MaxIdleConns:          32,
			IdleConnTimeout:       90 * time.Second,
		},
	}

	cnpjChain, err := buildCnpj(cfg, client)
	if err != nil {
		log.Fatal(err)
	}
	cepChain, err := buildCep(cfg, client)
	if err != nil {
		log.Fatal(err)
	}

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           (&httpapi.Server{Auth: tokens, Cnpj: cnpjChain, Cep: cepChain, Cache: cache.New(), TTLCnpj: cfg.CacheTTLCnpj, TTLCep: cfg.CacheTTLCep}).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      time.Duration(longest(len(cnpjChain.Names()), len(cepChain.Names())))*cfg.HTTPTimeout + 5*time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Printf("ouvindo :%s cnpj=%s cep=%s", cfg.Port, strings.Join(cnpjChain.Names(), ","), strings.Join(cepChain.Names(), ","))
		errCh <- srv.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shut, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := srv.Shutdown(shut); err != nil {
			log.Fatal(err)
		}
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}
}

func buildCnpj(cfg config.Config, client *http.Client) (*lookup.Chain, error) {
	var list []provider.Lookup
	for _, name := range cfg.CnpjChain {
		switch name {
		case "cnpjws":
			list = append(list, cnpj.NewCnpjWs(client, cfg.CnpjwsBaseURL, cfg.HTTPTimeout))
		case "minhareceita":
			list = append(list, cnpj.NewMinhaReceita(client, cfg.MinhaReceitaBaseURL, cfg.HTTPTimeout))
		case "apipj":
			if cfg.ApiPJToken == "" {
				log.Printf("apipj sem APIPJ_TOKEN, fora da cadeia")
				continue
			}
			list = append(list, cnpj.NewApiPJ(client, cfg.ApiPJBaseURL, cfg.ApiPJToken, cfg.HTTPTimeout))
		case "receitaws":
			if cfg.ReceitaWSToken == "" {
				log.Printf("receitaws sem RECEITAWS_TOKEN, fora da cadeia")
				continue
			}
			list = append(list, cnpj.NewReceitaWS(client, cfg.ReceitaWSBaseURL, cfg.ReceitaWSToken, cfg.ReceitaWSDays, cfg.HTTPTimeout))
		case "cnpja":
			if cfg.CnpjaAPIKey == "" {
				log.Printf("cnpja sem CNPJA_API_KEY, fora da cadeia")
				continue
			}
			list = append(list, cnpj.NewCnpja(client, cfg.CnpjaBaseURL, cfg.CnpjaAPIKey, cfg.HTTPTimeout))
		default:
			return nil, fmt.Errorf("provedor de cnpj desconhecido: %s", name)
		}
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("cadeia de cnpj vazia")
	}
	return lookup.New(list), nil
}

func buildCep(cfg config.Config, client *http.Client) (*lookup.Chain, error) {
	var list []provider.Lookup
	for _, name := range cfg.CepChain {
		switch name {
		case "viacep":
			list = append(list, cep.NewViaCep(client, cfg.ViaCepBaseURL, cfg.HTTPTimeout))
		case "brasilapi":
			list = append(list, cep.NewBrasilAPI(client, cfg.BrasilAPIBaseURL, cfg.HTTPTimeout))
		default:
			return nil, fmt.Errorf("provedor de cep desconhecido: %s", name)
		}
	}
	if len(list) == 0 {
		return nil, fmt.Errorf("cadeia de cep vazia")
	}
	return lookup.New(list), nil
}

func longest(a, b int) int {
	if a > b {
		return a
	}
	return b
}
