package cnpj

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gw.degitale.com.br/internal/model"
	"gw.degitale.com.br/internal/provider"
)

type Cnpja struct {
	client  *http.Client
	baseURL string
	apiKey  string
	timeout time.Duration
}

func NewCnpja(client *http.Client, baseURL, apiKey string, timeout time.Duration) *Cnpja {
	return &Cnpja{
		client:  client,
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		timeout: timeout,
	}
}

func (c *Cnpja) Name() string { return "cnpja" }

func (c *Cnpja) Lookup(ctx context.Context, cnpj string) (any, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	url := c.baseURL + "/office/" + cnpj + "?simples=true&simplesHistory=true&registrations=BR&geocoding=true"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, &provider.Error{Kind: provider.KindTemporary, Detail: "requisicao"}
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", c.apiKey)
	req.Header.Set("User-Agent", "degitale-gw/1.0")

	status, body, err := provider.Do(ctx, c.client, req)
	if err != nil {
		return nil, err
	}
	if err := provider.Classify(status); err != nil {
		return nil, err
	}
	out, err := MapCnpja(body)
	if err != nil {
		return nil, &provider.Error{Kind: provider.KindTemporary, Detail: "json"}
	}
	return out, nil
}

type cnpjaPayload struct {
	Company *struct {
		Name    string `json:"name"`
		Simples *struct {
			Optant bool `json:"optant"`
		} `json:"simples"`
	} `json:"company"`
	Alias   *string `json:"alias"`
	Founded string  `json:"founded"`
	Head    *bool   `json:"head"`
	Status  *struct {
		Text string `json:"text"`
	} `json:"status"`
	Address *struct {
		Municipality int      `json:"municipality"`
		Street       string   `json:"street"`
		Number       string   `json:"number"`
		Details      *string  `json:"details"`
		District     string   `json:"district"`
		Zip          string   `json:"zip"`
		Latitude     *float64 `json:"latitude"`
		Longitude    *float64 `json:"longitude"`
	} `json:"address"`
	Phones []struct {
		Area   string `json:"area"`
		Number string `json:"number"`
	} `json:"phones"`
	Emails []struct {
		Address string `json:"address"`
	} `json:"emails"`
	MainActivity *struct {
		ID   flexID `json:"id"`
		Text string `json:"text"`
	} `json:"mainActivity"`
	SideActivities []struct {
		ID   flexID `json:"id"`
		Text string `json:"text"`
	} `json:"sideActivities"`
	Registrations []struct {
		State  string `json:"state"`
		Number string `json:"number"`
	} `json:"registrations"`
}

func MapCnpja(body []byte) (model.Cnpj, error) {
	var payload cnpjaPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		return model.Cnpj{}, err
	}
	out := model.EmptyCnpj()
	if payload.Company != nil {
		out.NmEmpresarial = strings.TrimSpace(payload.Company.Name)
		// A API do CNPJá expõe company.simples.optant. O PHP confere um campo
		// aninhado simples.simples que não existe nesse provedor.
		if payload.Company.Simples != nil {
			opt := payload.Company.Simples.Optant
			out.FlOptanteSimplesNacional = &opt
		}
	}
	out.NmFantasia = model.StrPtr(deref(payload.Alias))
	if payload.Head != nil && !*payload.Head {
		out.CdTipoEstabelecimento = "FILIAL"
	} else {
		out.CdTipoEstabelecimento = "MATRIZ"
	}
	out.DtAbertura = model.DateOnly(payload.Founded)
	if payload.Status != nil {
		out.DsSituacaoCadastral = model.StrPtr(payload.Status.Text)
	}

	if payload.MainActivity != nil && (payload.MainActivity.ID != "" || payload.MainActivity.Text != "") {
		out.Atividades = append(out.Atividades, model.Atividade{
			CdAtividade: string(payload.MainActivity.ID),
			NmAtividade: payload.MainActivity.Text,
			FlPrincipal: true,
		})
	}
	for _, as := range payload.SideActivities {
		out.Atividades = append(out.Atividades, model.Atividade{
			CdAtividade: string(as.ID),
			NmAtividade: as.Text,
			FlPrincipal: false,
		})
	}

	if payload.Address != nil {
		out.Endereco.DsNumero = model.StrPtr(payload.Address.Number)
		if payload.Address.Details != nil {
			out.Endereco.DsComplemento = model.StrPtr(*payload.Address.Details)
		}
		out.Endereco.DsBairro = model.StrPtr(payload.Address.District)
		out.Endereco.DsLogradouro = model.StrPtr(payload.Address.Street)
		out.Endereco.NrCep = model.StrPtr(payload.Address.Zip)
		out.Endereco.IdCidadeIbge = model.IntPtr(payload.Address.Municipality)
		out.Endereco.VlLatitude = payload.Address.Latitude
		out.Endereco.VlLongitude = payload.Address.Longitude
		out.Endereco.CdTipoLogradouro, out.Endereco.DsLogradouro = model.SplitTipo(out.Endereco.CdTipoLogradouro, out.Endereco.DsLogradouro)
		out.Endereco.CdTipoBairro, out.Endereco.DsBairro = model.SplitTipo(out.Endereco.CdTipoBairro, out.Endereco.DsBairro)
	}

	for _, t := range payload.Phones {
		numero := strings.TrimSpace(t.Area + t.Number)
		if numero == "" {
			continue
		}
		out.Telefones = append(out.Telefones, model.Telefone{
			CdIso3Pais: "BRA",
			NrTelefone: numero,
		})
	}
	for _, e := range payload.Emails {
		if strings.TrimSpace(e.Address) == "" {
			continue
		}
		out.Emails = append(out.Emails, model.Email{DsEmail: strings.TrimSpace(e.Address)})
	}
	for _, ie := range payload.Registrations {
		if strings.TrimSpace(ie.Number) == "" {
			continue
		}
		out.InscricoesEstaduais = append(out.InscricoesEstaduais, model.InscricaoEstadual{
			NrInscricaoEstadual: strings.TrimSpace(ie.Number),
			CdEstado:            model.StrPtr(ie.State),
		})
	}
	return out, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// flexID aceita número (CNPJá) ou string (CNPJ.ws) e grava sempre como texto.
type flexID string

func (f *flexID) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if len(b) == 0 || string(b) == "null" {
		*f = ""
		return nil
	}
	if b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*f = flexID(strings.TrimSpace(s))
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		return fmt.Errorf("id invalido")
	}
	*f = flexID(n.String())
	return nil
}
