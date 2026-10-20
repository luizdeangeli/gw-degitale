package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port             string
	TokensRaw        string
	CnpjChain        []string
	CepChain         []string
	CnpjaBaseURL     string
	CnpjaAPIKey      string
	CnpjwsBaseURL    string
	ReceitaWSBaseURL string
	ReceitaWSToken   string
	ReceitaWSDays    int
	ApiPJBaseURL     string
	ApiPJToken       string
	ViaCepBaseURL    string
	HTTPTimeout      time.Duration
	CacheTTLCnpj     time.Duration
	CacheTTLCep      time.Duration
}

func Load() (Config, error) {
	timeout, err := parseDuration("HTTP_TIMEOUT", "8s")
	if err != nil {
		return Config{}, err
	}
	ttlCnpj, err := parseDuration("CACHE_TTL_CNPJ", "24h")
	if err != nil {
		return Config{}, err
	}
	ttlCep, err := parseDuration("CACHE_TTL_CEP", "168h")
	if err != nil {
		return Config{}, err
	}

	cnpjChain := parseList(getenv("CNPJ_CHAIN", "cnpjws,apipj,receitaws,cnpja"))
	cepChain := parseList(getenv("CEP_CHAIN", "viacep"))
	if len(cnpjChain) == 0 {
		return Config{}, fmt.Errorf("CNPJ_CHAIN vazia")
	}
	if len(cepChain) == 0 {
		return Config{}, fmt.Errorf("CEP_CHAIN vazia")
	}
	days, err := parseDays("RECEITAWS_DAYS", "30")
	if err != nil {
		return Config{}, err
	}

	return Config{
		Port:             getenv("PORT", "8080"),
		TokensRaw:        os.Getenv("GW_TOKENS"),
		CnpjChain:        cnpjChain,
		CepChain:         cepChain,
		CnpjaBaseURL:     trimSlash(getenv("CNPJA_BASE_URL", "https://api.cnpja.com")),
		CnpjaAPIKey:      strings.TrimSpace(os.Getenv("CNPJA_API_KEY")),
		CnpjwsBaseURL:    trimSlash(getenv("CNPJWS_BASE_URL", "https://publica.cnpj.ws")),
		ReceitaWSBaseURL: trimSlash(getenv("RECEITAWS_BASE_URL", "https://receitaws.com.br/v1")),
		ReceitaWSToken:   strings.TrimSpace(os.Getenv("RECEITAWS_TOKEN")),
		ReceitaWSDays:    days,
		ApiPJBaseURL:     trimSlash(getenv("APIPJ_BASE_URL", "https://apipj.com.br/v1")),
		ApiPJToken:       strings.TrimSpace(os.Getenv("APIPJ_TOKEN")),
		ViaCepBaseURL:    trimSlash(getenv("VIACEP_BASE_URL", "https://viacep.com.br")),
		HTTPTimeout:      timeout,
		CacheTTLCnpj:     ttlCnpj,
		CacheTTLCep:      ttlCep,
	}, nil
}

func getenv(key, fallback string) string {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	return v
}

func parseDuration(key, fallback string) (time.Duration, error) {
	raw := getenv(key, fallback)
	d, err := time.ParseDuration(raw)
	if err != nil {
		return 0, fmt.Errorf("%s invalido: %s", key, raw)
	}
	if d <= 0 {
		return 0, fmt.Errorf("%s deve ser maior que zero", key)
	}
	return d, nil
}

func parseDays(key, fallback string) (int, error) {
	raw := getenv(key, fallback)
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, fmt.Errorf("%s invalido: %s", key, raw)
	}
	return n, nil
}

func parseList(raw string) []string {
	var out []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.ToLower(strings.TrimSpace(part))
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func trimSlash(s string) string {
	return strings.TrimRight(strings.TrimSpace(s), "/")
}
