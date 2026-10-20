package cnpj

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"gw.degitale.com.br/internal/model"
	"gw.degitale.com.br/internal/provider"
)

type CnpjWs struct {
	client  *http.Client
	baseURL string
	timeout time.Duration
}

func NewCnpjWs(client *http.Client, baseURL string, timeout time.Duration) *CnpjWs {
	return &CnpjWs{
		client:  client,
		baseURL: strings.TrimRight(baseURL, "/"),
		timeout: timeout,
	}
}

func (c *CnpjWs) Name() string { return "cnpjws" }

func (c *CnpjWs) Lookup(ctx context.Context, cnpj string) (any, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/cnpj/"+cnpj, nil)
	if err != nil {
		return nil, &provider.Error{Kind: provider.KindTemporary, Detail: "requisicao"}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "degitale-gw/1.0")

	status, body, err := provider.Do(ctx, c.client, req)
	if err != nil {
		return nil, err
	}
	if err := provider.Classify(status); err != nil {
		return nil, err
	}
	out, err := MapCnpjWs(body)
	if err != nil {
		return nil, &provider.Error{Kind: provider.KindTemporary, Detail: "json"}
	}
	return out, nil
}

type cnpjWsPayload struct {
	RazaoSocial string `json:"razao_social"`
	Simples     *struct {
		Simples string `json:"simples"`
	} `json:"simples"`
	Estabelecimento *struct {
		Tipo                string  `json:"tipo"`
		NomeFantasia        *string `json:"nome_fantasia"`
		SituacaoCadastral   string  `json:"situacao_cadastral"`
		DataInicioAtividade string  `json:"data_inicio_atividade"`
		TipoLogradouro      string  `json:"tipo_logradouro"`
		Logradouro          string  `json:"logradouro"`
		Numero              string  `json:"numero"`
		Complemento         string  `json:"complemento"`
		Bairro              string  `json:"bairro"`
		Cep                 string  `json:"cep"`
		DDD1                string  `json:"ddd1"`
		Telefone1           string  `json:"telefone1"`
		DDD2                string  `json:"ddd2"`
		Telefone2           string  `json:"telefone2"`
		Email               string  `json:"email"`
		AtividadePrincipal  *struct {
			ID        flexID `json:"id"`
			Descricao string `json:"descricao"`
		} `json:"atividade_principal"`
		AtividadesSecundarias []struct {
			ID        flexID `json:"id"`
			Descricao string `json:"descricao"`
		} `json:"atividades_secundarias"`
		Cidade *struct {
			IbgeID int `json:"ibge_id"`
		} `json:"cidade"`
		InscricoesEstaduais []struct {
			InscricaoEstadual string `json:"inscricao_estadual"`
			Estado            *struct {
				IbgeID int `json:"ibge_id"`
			} `json:"estado"`
		} `json:"inscricoes_estaduais"`
	} `json:"estabelecimento"`
}

func MapCnpjWs(body []byte) (model.Cnpj, error) {
	var payload cnpjWsPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return model.Cnpj{}, err
	}
	out := model.EmptyCnpj()
	out.NmEmpresarial = strings.TrimSpace(payload.RazaoSocial)
	if payload.Simples != nil && strings.TrimSpace(payload.Simples.Simples) != "" {
		opt := strings.TrimSpace(payload.Simples.Simples) == "Sim"
		out.FlOptanteSimplesNacional = &opt
	}

	est := payload.Estabelecimento
	if est == nil {
		return out, nil
	}

	out.NmFantasia = model.StrPtr(deref(est.NomeFantasia))
	if strings.TrimSpace(est.Tipo) == "" {
		out.CdTipoEstabelecimento = "MATRIZ"
	} else {
		out.CdTipoEstabelecimento = strings.ToUpper(strings.TrimSpace(est.Tipo))
	}
	out.DtAbertura = model.DateOnly(est.DataInicioAtividade)
	out.DsSituacaoCadastral = model.StrPtr(est.SituacaoCadastral)

	if est.AtividadePrincipal != nil && (est.AtividadePrincipal.ID != "" || est.AtividadePrincipal.Descricao != "") {
		out.Atividades = append(out.Atividades, model.Atividade{
			CdAtividade: string(est.AtividadePrincipal.ID),
			NmAtividade: est.AtividadePrincipal.Descricao,
			FlPrincipal: true,
		})
	}
	for _, as := range est.AtividadesSecundarias {
		out.Atividades = append(out.Atividades, model.Atividade{
			CdAtividade: string(as.ID),
			NmAtividade: as.Descricao,
			FlPrincipal: false,
		})
	}

	out.Endereco.DsNumero = model.StrPtr(est.Numero)
	out.Endereco.DsComplemento = model.StrPtr(est.Complemento)
	out.Endereco.DsBairro = model.StrPtr(est.Bairro)
	out.Endereco.CdTipoLogradouro = model.StrPtr(est.TipoLogradouro)
	out.Endereco.DsLogradouro = model.StrPtr(est.Logradouro)
	out.Endereco.NrCep = model.StrPtr(est.Cep)
	if est.Cidade != nil {
		out.Endereco.IdCidadeIbge = model.IntPtr(est.Cidade.IbgeID)
	}
	out.Endereco.CdTipoLogradouro, out.Endereco.DsLogradouro = model.SplitTipo(out.Endereco.CdTipoLogradouro, out.Endereco.DsLogradouro)
	out.Endereco.CdTipoBairro, out.Endereco.DsBairro = model.SplitTipo(out.Endereco.CdTipoBairro, out.Endereco.DsBairro)

	if tel := joinPhone(est.DDD1, est.Telefone1); tel != "" {
		out.Telefones = append(out.Telefones, model.Telefone{CdIso3Pais: "BRA", NrTelefone: tel})
	}
	if tel := joinPhone(est.DDD2, est.Telefone2); tel != "" {
		out.Telefones = append(out.Telefones, model.Telefone{CdIso3Pais: "BRA", NrTelefone: tel})
	}
	if email := strings.TrimSpace(est.Email); email != "" {
		out.Emails = append(out.Emails, model.Email{DsEmail: email})
	}
	for _, ie := range est.InscricoesEstaduais {
		if strings.TrimSpace(ie.InscricaoEstadual) == "" {
			continue
		}
		item := model.InscricaoEstadual{NrInscricaoEstadual: strings.TrimSpace(ie.InscricaoEstadual)}
		if ie.Estado != nil {
			item.IdEstadoIbge = model.IntPtr(ie.Estado.IbgeID)
		}
		out.InscricoesEstaduais = append(out.InscricoesEstaduais, item)
	}
	return out, nil
}

func joinPhone(ddd, numero string) string {
	ddd = strings.TrimSpace(ddd)
	numero = strings.TrimSpace(numero)
	if ddd == "" || numero == "" {
		return ""
	}
	return ddd + numero
}
