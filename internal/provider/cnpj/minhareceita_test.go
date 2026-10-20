package cnpj

import "testing"

func TestMapMinhaReceita(t *testing.T) {
	body := []byte(`{
	  "razao_social": "OPEN KNOWLEDGE BRASIL",
	  "nome_fantasia": "REDE PELO CONHECIMENTO LIVRE",
	  "descricao_identificador_matriz_filial": "MATRIZ",
	  "data_inicio_atividade": "2013-10-03",
	  "descricao_situacao_cadastral": "ATIVA",
	  "opcao_pelo_simples": false,
	  "cnae_fiscal": 9430800,
	  "cnae_fiscal_descricao": "Atividades de associacoes",
	  "cnaes_secundarios": [{"codigo": 9493600, "descricao": "Cultura"}],
	  "descricao_tipo_de_logradouro": "AVENIDA",
	  "logradouro": "PAULISTA",
	  "numero": "37",
	  "complemento": "ANDAR 4",
	  "bairro": "BELA VISTA",
	  "cep": "01311902",
	  "codigo_municipio_ibge": 3550308,
	  "ddd_telefone_1": "1123851939",
	  "ddd_telefone_2": "",
	  "email": null
	}`)
	out, err := MapMinhaReceita(body)
	if err != nil {
		t.Fatal(err)
	}
	if out.NmEmpresarial != "OPEN KNOWLEDGE BRASIL" || out.DtAbertura != "2013-10-03" {
		t.Fatalf("%+v", out)
	}
	if out.FlOptanteSimplesNacional == nil || *out.FlOptanteSimplesNacional {
		t.Fatal("simples")
	}
	if out.Atividades[0].CdAtividade != "9430800" || !out.Atividades[0].FlPrincipal {
		t.Fatalf("%+v", out.Atividades)
	}
	if out.Endereco.CdTipoLogradouro == nil || *out.Endereco.CdTipoLogradouro != "AVENIDA" {
		t.Fatalf("tipo %+v", out.Endereco.CdTipoLogradouro)
	}
	if out.Endereco.DsLogradouro == nil || *out.Endereco.DsLogradouro != "PAULISTA" {
		t.Fatalf("logradouro %+v", out.Endereco.DsLogradouro)
	}
	if out.Endereco.IdCidadeIbge == nil || *out.Endereco.IdCidadeIbge != 3550308 {
		t.Fatalf("ibge %+v", out.Endereco.IdCidadeIbge)
	}
	if len(out.Telefones) != 1 || out.Telefones[0].NrTelefone != "1123851939" {
		t.Fatalf("%+v", out.Telefones)
	}
}
