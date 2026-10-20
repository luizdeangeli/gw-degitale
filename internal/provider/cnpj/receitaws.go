package cnpj

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gw.degitale.com.br/internal/model"
	"gw.degitale.com.br/internal/provider"
)

type ReceitaWS struct {
	client  *http.Client
	baseURL string
	token   string
	days    int
	timeout time.Duration
}

func NewReceitaWS(client *http.Client, baseURL, token string, days int, timeout time.Duration) *ReceitaWS {
	return &ReceitaWS{
		client:  client,
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		days:    days,
		timeout: timeout,
	}
}

func (r *ReceitaWS) Name() string { return "receitaws" }

func (r *ReceitaWS) Lookup(ctx context.Context, cnpj string) (any, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	u := fmt.Sprintf("%s/cnpj/%s/days/%d?fallback=cacheOnError", r.baseURL, cnpj, r.days)
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return nil, &provider.Error{Kind: provider.KindTemporary, Detail: "requisicao"}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+r.token)
	req.Header.Set("User-Agent", "degitale-gw/1.0")

	status, body, err := provider.Do(ctx, r.client, req)
	if err != nil {
		return nil, err
	}
	if err := provider.Classify(status); err != nil {
		return nil, err
	}
	out, err := MapReceitaWS(body)
	if err != nil {
		if provider.IsKind(err, provider.KindNotFound) {
			return nil, err
		}
		return nil, &provider.Error{Kind: provider.KindTemporary, Detail: "json"}
	}
	return out, nil
}

type receitaWSPayload struct {
	Status                string               `json:"status"`
	Message               string               `json:"message"`
	Tipo                  string               `json:"tipo"`
	Nome                  string               `json:"nome"`
	Fantasia              string               `json:"fantasia"`
	Abertura              string               `json:"abertura"`
	Situacao              string               `json:"situacao"`
	Logradouro            string               `json:"logradouro"`
	Numero                string               `json:"numero"`
	Complemento           string               `json:"complemento"`
	Cep                   string               `json:"cep"`
	Bairro                string               `json:"bairro"`
	Email                 string               `json:"email"`
	Telefone              string               `json:"telefone"`
	AtividadePrincipal    []receitaWSAtividade `json:"atividade_principal"`
	AtividadesSecundarias []receitaWSAtividade `json:"atividades_secundarias"`
	Simples               *struct {
		Optante bool `json:"optante"`
	} `json:"simples"`
}

type receitaWSAtividade struct {
	Code string `json:"code"`
	Text string `json:"text"`
}

func MapReceitaWS(body []byte) (model.Cnpj, error) {
	var payload receitaWSPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return model.Cnpj{}, err
	}
	if strings.EqualFold(strings.TrimSpace(payload.Status), "ERROR") {
		return model.Cnpj{}, receitawsStatus(payload.Message)
	}
	if payload.Status != "" && !strings.EqualFold(payload.Status, "OK") && strings.TrimSpace(payload.Nome) == "" {
		return model.Cnpj{}, &provider.Error{Kind: provider.KindTemporary, Detail: "json"}
	}

	out := model.EmptyCnpj()
	out.NmEmpresarial = strings.TrimSpace(payload.Nome)
	out.NmFantasia = model.StrPtr(payload.Fantasia)
	if strings.TrimSpace(payload.Tipo) == "" {
		out.CdTipoEstabelecimento = "MATRIZ"
	} else {
		out.CdTipoEstabelecimento = strings.ToUpper(strings.TrimSpace(payload.Tipo))
	}
	out.DtAbertura = dataBR(payload.Abertura)
	out.DsSituacaoCadastral = model.StrPtr(payload.Situacao)
	if payload.Simples != nil {
		opt := payload.Simples.Optante
		out.FlOptanteSimplesNacional = &opt
	}

	for _, at := range payload.AtividadePrincipal {
		appendAtividade(&out, at, true)
	}
	for _, at := range payload.AtividadesSecundarias {
		appendAtividade(&out, at, false)
	}

	out.Endereco.DsNumero = model.StrPtr(payload.Numero)
	out.Endereco.DsComplemento = model.StrPtr(payload.Complemento)
	out.Endereco.DsBairro = model.StrPtr(payload.Bairro)
	out.Endereco.DsLogradouro = model.StrPtr(payload.Logradouro)
	out.Endereco.NrCep = model.StrPtr(onlyDigits(payload.Cep))
	out.Endereco.CdTipoLogradouro, out.Endereco.DsLogradouro = model.SplitTipo(out.Endereco.CdTipoLogradouro, out.Endereco.DsLogradouro)
	out.Endereco.CdTipoBairro, out.Endereco.DsBairro = model.SplitTipo(out.Endereco.CdTipoBairro, out.Endereco.DsBairro)

	for _, tel := range phones(payload.Telefone) {
		out.Telefones = append(out.Telefones, model.Telefone{CdIso3Pais: "BRA", NrTelefone: tel})
	}
	if email := strings.TrimSpace(payload.Email); email != "" {
		out.Emails = append(out.Emails, model.Email{DsEmail: email})
	}
	return out, nil
}

func receitawsStatus(message string) error {
	m := strings.ToLower(strings.TrimSpace(message))
	if strings.Contains(m, "inválido") || strings.Contains(m, "invalido") || strings.Contains(m, "rejeitado") || strings.Contains(m, "não encontr") || strings.Contains(m, "nao encontr") {
		return &provider.Error{Kind: provider.KindNotFound, Status: http.StatusNotFound}
	}
	return &provider.Error{Kind: provider.KindTemporary, Detail: "erro"}
}

func appendAtividade(out *model.Cnpj, at receitaWSAtividade, principal bool) {
	text := strings.TrimSpace(at.Text)
	if strings.EqualFold(text, "Não informada") || strings.EqualFold(text, "Nao informada") {
		return
	}
	code := activityCode(at.Code)
	if code == "" && text == "" {
		return
	}
	out.Atividades = append(out.Atividades, model.Atividade{
		CdAtividade: code,
		NmAtividade: text,
		FlPrincipal: principal,
	})
}

func activityCode(code string) string {
	d := onlyDigits(code)
	if d == "" || strings.Trim(d, "0") == "" {
		return ""
	}
	return d
}

func dataBR(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 10 && s[2] == '/' && s[5] == '/' {
		return s[6:10] + "-" + s[3:5] + "-" + s[0:2]
	}
	return model.DateOnly(s)
}

func phones(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == '/' || r == ',' || r == ';'
	})
	var out []string
	for _, part := range parts {
		d := onlyDigits(part)
		if len(d) >= 10 {
			out = append(out, d)
		}
	}
	return out
}

func onlyDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}
