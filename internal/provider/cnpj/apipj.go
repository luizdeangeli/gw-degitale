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

type ApiPJ struct {
	client  *http.Client
	baseURL string
	token   string
	timeout time.Duration
}

func NewApiPJ(client *http.Client, baseURL, token string, timeout time.Duration) *ApiPJ {
	return &ApiPJ{
		client:  client,
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
		timeout: timeout,
	}
}

func (a *ApiPJ) Name() string { return "apipj" }

func (a *ApiPJ) Lookup(ctx context.Context, cnpj string) (any, error) {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, a.baseURL+"/cnpj/"+cnpj, nil)
	if err != nil {
		return nil, &provider.Error{Kind: provider.KindTemporary, Detail: "requisicao"}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.token)
	req.Header.Set("User-Agent", "degitale-gw/1.0")

	status, body, err := provider.Do(ctx, a.client, req)
	if err != nil {
		return nil, err
	}
	// 404 aqui é ausência no lote mensal, não prova de que o CNPJ não existe.
	if status == http.StatusNotFound {
		return nil, &provider.Error{Kind: provider.KindTemporary, Status: status}
	}
	if err := provider.Classify(status); err != nil {
		return nil, err
	}
	out, err := MapApiPJ(body)
	if err != nil {
		return nil, &provider.Error{Kind: provider.KindTemporary, Detail: "json"}
	}
	return out, nil
}

type apiPJPayload struct {
	RazaoSocial         string `json:"razao_social"`
	NomeFantasia        string `json:"nome_fantasia"`
	MatrizFilial        string `json:"matriz_filial"`
	DataInicioAtividade string `json:"data_inicio_atividade"`
	SituacaoCadastral   *struct {
		Descricao string `json:"descricao"`
	} `json:"situacao_cadastral"`
	CnaePrincipal *struct {
		Codigo    flexID `json:"codigo"`
		Descricao string `json:"descricao"`
	} `json:"cnae_principal"`
	CnaesSecundarios []struct {
		Codigo    flexID `json:"codigo"`
		Descricao string `json:"descricao"`
	} `json:"cnaes_secundarios"`
	Endereco *struct {
		TipoLogradouro string `json:"tipo_logradouro"`
		Logradouro     string `json:"logradouro"`
		Numero         string `json:"numero"`
		Complemento    string `json:"complemento"`
		Bairro         string `json:"bairro"`
		Cep            string `json:"cep"`
		Municipio      *struct {
			CodigoIbge flexID `json:"codigo_ibge"`
		} `json:"municipio"`
	} `json:"endereco"`
	Telefones []struct {
		DDD    string `json:"ddd"`
		Numero string `json:"numero"`
	} `json:"telefones"`
	Email   string `json:"email"`
	Simples *struct {
		Optante bool `json:"optante"`
	} `json:"simples"`
}

func MapApiPJ(body []byte) (model.Cnpj, error) {
	var payload apiPJPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return model.Cnpj{}, err
	}
	out := model.EmptyCnpj()
	out.NmEmpresarial = strings.TrimSpace(payload.RazaoSocial)
	out.NmFantasia = model.StrPtr(payload.NomeFantasia)
	if strings.TrimSpace(payload.MatrizFilial) == "" {
		out.CdTipoEstabelecimento = "MATRIZ"
	} else {
		out.CdTipoEstabelecimento = strings.ToUpper(strings.TrimSpace(payload.MatrizFilial))
	}
	out.DtAbertura = model.DateOnly(payload.DataInicioAtividade)
	if payload.SituacaoCadastral != nil {
		out.DsSituacaoCadastral = model.StrPtr(payload.SituacaoCadastral.Descricao)
	}
	if payload.Simples != nil {
		opt := payload.Simples.Optante
		out.FlOptanteSimplesNacional = &opt
	}
	if payload.CnaePrincipal != nil {
		appendAtividade(&out, receitaWSAtividade{Code: string(payload.CnaePrincipal.Codigo), Text: payload.CnaePrincipal.Descricao}, true)
	}
	for _, at := range payload.CnaesSecundarios {
		appendAtividade(&out, receitaWSAtividade{Code: string(at.Codigo), Text: at.Descricao}, false)
	}
	if payload.Endereco != nil {
		end := payload.Endereco
		out.Endereco.DsNumero = model.StrPtr(end.Numero)
		out.Endereco.DsComplemento = model.StrPtr(end.Complemento)
		out.Endereco.DsBairro = model.StrPtr(end.Bairro)
		out.Endereco.CdTipoLogradouro = model.StrPtr(end.TipoLogradouro)
		out.Endereco.DsLogradouro = model.StrPtr(end.Logradouro)
		out.Endereco.NrCep = model.StrPtr(onlyDigits(end.Cep))
		if end.Municipio != nil {
			if n, err := strconv.Atoi(string(end.Municipio.CodigoIbge)); err == nil {
				out.Endereco.IdCidadeIbge = model.IntPtr(n)
			}
		}
		out.Endereco.CdTipoLogradouro, out.Endereco.DsLogradouro = model.SplitTipo(out.Endereco.CdTipoLogradouro, out.Endereco.DsLogradouro)
		out.Endereco.CdTipoBairro, out.Endereco.DsBairro = model.SplitTipo(out.Endereco.CdTipoBairro, out.Endereco.DsBairro)
	}
	for _, tel := range payload.Telefones {
		numero := strings.TrimSpace(tel.DDD + tel.Numero)
		if numero == "" {
			continue
		}
		out.Telefones = append(out.Telefones, model.Telefone{CdIso3Pais: "BRA", NrTelefone: onlyDigits(numero)})
	}
	if email := strings.TrimSpace(payload.Email); email != "" {
		out.Emails = append(out.Emails, model.Email{DsEmail: email})
	}
	return out, nil
}
