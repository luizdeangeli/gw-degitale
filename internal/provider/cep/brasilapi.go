package cep

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gw.degitale.com.br/internal/model"
	"gw.degitale.com.br/internal/provider"
)

type BrasilAPI struct {
	client  *http.Client
	baseURL string
	timeout time.Duration
}

func NewBrasilAPI(client *http.Client, baseURL string, timeout time.Duration) *BrasilAPI {
	return &BrasilAPI{
		client:  client,
		baseURL: strings.TrimRight(baseURL, "/"),
		timeout: timeout,
	}
}

func (b *BrasilAPI) Name() string { return "brasilapi" }

func (b *BrasilAPI) Lookup(ctx context.Context, cep string) (any, error) {
	ctx, cancel := context.WithTimeout(ctx, b.timeout)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, b.baseURL+"/cep/v2/"+cep, nil)
	if err != nil {
		return nil, &provider.Error{Kind: provider.KindTemporary, Detail: "requisicao"}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "degitale-gw/1.0")

	status, body, err := provider.Do(ctx, b.client, req)
	if err != nil {
		return nil, err
	}
	if err := provider.Classify(status); err != nil {
		return nil, err
	}
	out, err := MapBrasilAPI(body)
	if err != nil {
		return nil, &provider.Error{Kind: provider.KindTemporary, Detail: "json"}
	}
	return out, nil
}

type brasilAPIPayload struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Ibge         *struct {
		City string `json:"city"`
	} `json:"ibge"`
}

func MapBrasilAPI(body []byte) (model.Cep, error) {
	var payload brasilAPIPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return model.Cep{}, err
	}
	out := model.Cep{
		Cep:        formatCEP(payload.Cep),
		Bairro:     payload.Neighborhood,
		Localidade: payload.City,
		Uf:         payload.State,
	}
	if payload.Ibge != nil {
		out.Ibge = payload.Ibge.City
	}
	tipo, resto := model.SplitTipo(nil, model.StrPtr(payload.Street))
	if tipo != nil {
		out.TipoLogradouro = *tipo
	}
	if resto != nil {
		out.Logradouro = *resto
	}
	return out, nil
}

func formatCEP(s string) string {
	d := digits(s)
	if len(d) != 8 {
		return strings.TrimSpace(s)
	}
	return d[:5] + "-" + d[5:]
}

func digits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
