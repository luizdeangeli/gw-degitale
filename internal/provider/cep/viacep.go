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

type ViaCep struct {
	client  *http.Client
	baseURL string
	timeout time.Duration
}

func NewViaCep(client *http.Client, baseURL string, timeout time.Duration) *ViaCep {
	return &ViaCep{
		client:  client,
		baseURL: strings.TrimRight(baseURL, "/"),
		timeout: timeout,
	}
}

func (v *ViaCep) Name() string { return "viacep" }

func (v *ViaCep) Lookup(ctx context.Context, cep string) (any, error) {
	ctx, cancel := context.WithTimeout(ctx, v.timeout)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, v.baseURL+"/ws/"+cep+"/json/", nil)
	if err != nil {
		return nil, &provider.Error{Kind: provider.KindTemporary, Detail: "requisicao"}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "degitale-gw/1.0")

	status, body, err := provider.Do(ctx, v.client, req)
	if err != nil {
		return nil, err
	}
	if status == http.StatusBadRequest {
		return nil, &provider.Error{Kind: provider.KindNotFound, Status: status}
	}
	if err := provider.Classify(status); err != nil {
		return nil, err
	}
	out, err := MapViaCep(body)
	if err != nil {
		if provider.IsKind(err, provider.KindNotFound) {
			return nil, err
		}
		return nil, &provider.Error{Kind: provider.KindTemporary, Detail: "json"}
	}
	return out, nil
}

type viaCepPayload struct {
	Erro        any    `json:"erro"`
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Ddd         string `json:"ddd"`
}

func MapViaCep(body []byte) (model.Cep, error) {
	var payload viaCepPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return model.Cep{}, err
	}
	if erroTrue(payload.Erro) {
		return model.Cep{}, &provider.Error{Kind: provider.KindNotFound, Status: http.StatusNotFound}
	}
	out := model.Cep{
		Cep:         payload.Cep,
		Complemento: payload.Complemento,
		Bairro:      payload.Bairro,
		Localidade:  payload.Localidade,
		Uf:          payload.Uf,
		Ibge:        payload.Ibge,
		Ddd:         payload.Ddd,
	}
	tipo, resto := model.SplitTipo(nil, model.StrPtr(payload.Logradouro))
	if tipo != nil {
		out.TipoLogradouro = *tipo
	}
	if resto != nil {
		out.Logradouro = *resto
	}
	return out, nil
}

func erroTrue(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		s := strings.ToLower(strings.TrimSpace(t))
		return s == "true" || s == "1"
	default:
		return false
	}
}
