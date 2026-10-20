package model

// Cnpj é o contrato normalizado, no mesmo formato que o SMG monta em Externo\Cnpja e Externo\CnpjWs.
type Cnpj struct {
	NmEmpresarial            string              `json:"nmEmpresarial"`
	NmFantasia               *string             `json:"nmFantasia"`
	CdTipoEstabelecimento    string              `json:"cdTipoEstabelecimento"`
	DtAbertura               string              `json:"dtAbertura"`
	FlOptanteSimplesNacional *bool               `json:"flOptanteSimplesNacional"`
	DsSituacaoCadastral      *string             `json:"dsSituacaoCadastral"`
	Atividades               []Atividade         `json:"atividades"`
	Endereco                 Endereco            `json:"endereco"`
	Telefones                []Telefone          `json:"telefones"`
	Emails                   []Email             `json:"emails"`
	InscricoesEstaduais      []InscricaoEstadual `json:"inscricoesEstaduais"`
}

type Atividade struct {
	CdAtividade string `json:"cdAtividade"`
	NmAtividade string `json:"nmAtividade"`
	FlPrincipal bool   `json:"flPrincipal"`
}

type Endereco struct {
	DsNumero         *string  `json:"dsNumero"`
	DsComplemento    *string  `json:"dsComplemento"`
	DsBairro         *string  `json:"dsBairro"`
	CdTipoLogradouro *string  `json:"cdTipoLogradouro"`
	CdTipoBairro     *string  `json:"cdTipoBairro"`
	DsLogradouro     *string  `json:"dsLogradouro"`
	NrCep            *string  `json:"nrCep"`
	IdCidadeIbge     *int     `json:"idCidadeIbge"`
	Cidade           *string  `json:"cidade"`
	VlLatitude       *float64 `json:"vlLatitude"`
	VlLongitude      *float64 `json:"vlLongitude"`
}

type Telefone struct {
	CdIso3Pais string  `json:"cdIso3Pais"`
	Pais       *string `json:"pais"`
	NrTelefone string  `json:"nrTelefone"`
}

type Email struct {
	DsEmail string `json:"dsEmail"`
}

type InscricaoEstadual struct {
	NrInscricaoEstadual string  `json:"nrInscricaoEstadual"`
	CdEstado            *string `json:"cdEstado"`
	IdEstadoIbge        *int    `json:"idEstadoIbge"`
	Estado              *string `json:"estado"`
}

// Cep espelha o retorno do ViaCEP usado pelo SMG.
type Cep struct {
	Cep            string `json:"cep"`
	TipoLogradouro string `json:"tipoLogradouro"`
	Logradouro     string `json:"logradouro"`
	Complemento    string `json:"complemento"`
	Bairro         string `json:"bairro"`
	Localidade     string `json:"localidade"`
	Uf             string `json:"uf"`
	Ibge           string `json:"ibge"`
	Ddd            string `json:"ddd"`
}

func EmptyCnpj() Cnpj {
	return Cnpj{
		Atividades:          []Atividade{},
		Telefones:           []Telefone{},
		Emails:              []Email{},
		InscricoesEstaduais: []InscricaoEstadual{},
	}
}
