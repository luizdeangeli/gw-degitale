package cnpj

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gw.degitale.com.br/internal/model"
	"gw.degitale.com.br/internal/provider"
)

type MinhaReceita struct {
	client  *http.Client
	baseURL string
	timeout time.Duration
}

func NewMinhaReceita(client *http.Client, baseURL string, timeout time.Duration) *MinhaReceita {
	return &MinhaReceita{
		client:  client,
		baseURL: strings.TrimRight(baseURL, "/"),
		timeout: timeout,
	}
}

func (m *MinhaReceita) Name() string { return "minhareceita" }

func (m *MinhaReceita) Lookup(ctx context.Context, cnpj string) (any, error) {
	ctx, cancel := context.WithTimeout(ctx, m.timeout)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, m.baseURL+"/"+cnpj, nil)
	if err != nil {
		return nil, &provider.Error{Kind: provider.KindTemporary, Detail: "requisicao"}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "degitale-gw/1.0")

	status, body, err := provider.Do(ctx, m.client, req)
	if err != nil {
		return nil, err
	}
	// Base de dados abertos: 404 não encerra a cadeia.
	if status == http.StatusNotFound {
		return nil, &provider.Error{Kind: provider.KindTemporary, Status: status}
	}
	if err := provider.Classify(status); err != nil {
		return nil, err
	}
	out, err := MapMinhaReceita(body)
	if err != nil {
		return nil, &provider.Error{Kind: provider.KindTemporary, Detail: "json"}
	}
	return out, nil
}

type minhaReceitaPayload struct {
	RazaoSocial      string  `json:"razao_social"`
	NomeFantasia     *string `json:"nome_fantasia"`
	Tipo             string  `json:"descricao_identificador_matriz_filial"`
	DataInicio       string  `json:"data_inicio_atividade"`
	Situacao         string  `json:"descricao_situacao_cadastral"`
	OpcaoPeloSimples *bool   `json:"opcao_pelo_simples"`
	CnaeFiscal       flexID  `json:"cnae_fiscal"`
	CnaeDescricao    string  `json:"cnae_fiscal_descricao"`
	CnaesSecundarios []struct {
		Codigo    flexID `json:"codigo"`
		Descricao string `json:"descricao"`
	} `json:"cnaes_secundarios"`
	TipoLogradouro string  `json:"descricao_tipo_de_logradouro"`
	Logradouro     string  `json:"logradouro"`
	Numero         string  `json:"numero"`
	Complemento    string  `json:"complemento"`
	Bairro         string  `json:"bairro"`
	Cep            string  `json:"cep"`
	CodigoIbge     flexID  `json:"codigo_municipio_ibge"`
	Telefone1      string  `json:"ddd_telefone_1"`
	Telefone2      string  `json:"ddd_telefone_2"`
	Email          *string `json:"email"`
}

func MapMinhaReceita(body []byte) (model.Cnpj, error) {
	var payload minhaReceitaPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return model.Cnpj{}, err
	}
	out := model.EmptyCnpj()
	out.NmEmpresarial = strings.TrimSpace(payload.RazaoSocial)
	out.NmFantasia = model.StrPtr(deref(payload.NomeFantasia))
	if strings.TrimSpace(payload.Tipo) == "" {
		out.CdTipoEstabelecimento = "MATRIZ"
	} else {
		out.CdTipoEstabelecimento = strings.ToUpper(strings.TrimSpace(payload.Tipo))
	}
	out.DtAbertura = model.DateOnly(payload.DataInicio)
	out.DsSituacaoCadastral = model.StrPtr(payload.Situacao)
	if payload.OpcaoPeloSimples != nil {
		opt := *payload.OpcaoPeloSimples
		out.FlOptanteSimplesNacional = &opt
	}
	appendAtividade(&out, receitaWSAtividade{Code: string(payload.CnaeFiscal), Text: payload.CnaeDescricao}, true)
	for _, at := range payload.CnaesSecundarios {
		appendAtividade(&out, receitaWSAtividade{Code: string(at.Codigo), Text: at.Descricao}, false)
	}
	out.Endereco.DsNumero = model.StrPtr(payload.Numero)
	out.Endereco.DsComplemento = model.StrPtr(payload.Complemento)
	out.Endereco.DsBairro = model.StrPtr(payload.Bairro)
	out.Endereco.CdTipoLogradouro = model.StrPtr(payload.TipoLogradouro)
	out.Endereco.DsLogradouro = model.StrPtr(payload.Logradouro)
	out.Endereco.NrCep = model.StrPtr(onlyDigits(payload.Cep))
	if n, err := strconv.Atoi(string(payload.CodigoIbge)); err == nil {
		out.Endereco.IdCidadeIbge = model.IntPtr(n)
	}
	out.Endereco.CdTipoLogradouro, out.Endereco.DsLogradouro = model.SplitTipo(out.Endereco.CdTipoLogradouro, out.Endereco.DsLogradouro)
	out.Endereco.CdTipoBairro, out.Endereco.DsBairro = model.SplitTipo(out.Endereco.CdTipoBairro, out.Endereco.DsBairro)
	for _, tel := range []string{payload.Telefone1, payload.Telefone2} {
		d := onlyDigits(tel)
		if len(d) < 10 {
			continue
		}
		out.Telefones = append(out.Telefones, model.Telefone{CdIso3Pais: "BRA", NrTelefone: d})
	}
	if payload.Email != nil && strings.TrimSpace(*payload.Email) != "" {
		out.Emails = append(out.Emails, model.Email{DsEmail: strings.TrimSpace(*payload.Email)})
	}
	return out, nil
}
