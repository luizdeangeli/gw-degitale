package cnpj

import "testing"

func TestMapCnpja(t *testing.T) {
	body := []byte(`{
	  "company": {"name": "MAGAZINE LUIZA S/A", "simples": {"optant": false}},
	  "alias": "Magazine Luiza",
	  "founded": "1966-10-24T00:00:00.000Z",
	  "head": false,
	  "status": {"text": "Ativa"},
	  "address": {
	    "municipality": 3516200,
	    "street": "Rua Voluntarios da Franca",
	    "number": "1465",
	    "details": null,
	    "district": "Centro",
	    "zip": "14400490",
	    "latitude": -20.5,
	    "longitude": -47.4
	  },
	  "phones": [{"area": "16", "number": "37112002"}],
	  "emails": [{"address": "fiscal@example.com"}],
	  "mainActivity": {"id": 4713004, "text": "Lojas"},
	  "sideActivities": [{"id": 3329501, "text": "Montagem"}],
	  "registrations": [{"state": "SP", "number": "310035324119"}]
	}`)
	out, err := MapCnpja(body)
	if err != nil {
		t.Fatal(err)
	}
	if out.NmEmpresarial != "MAGAZINE LUIZA S/A" || out.CdTipoEstabelecimento != "FILIAL" {
		t.Fatalf("%+v", out)
	}
	if out.NmFantasia == nil || *out.NmFantasia != "Magazine Luiza" {
		t.Fatalf("fantasia %+v", out.NmFantasia)
	}
	if out.DtAbertura != "1966-10-24" {
		t.Fatal(out.DtAbertura)
	}
	if out.FlOptanteSimplesNacional == nil || *out.FlOptanteSimplesNacional {
		t.Fatal("simples")
	}
	if len(out.Atividades) != 2 || out.Atividades[0].CdAtividade != "4713004" || !out.Atividades[0].FlPrincipal {
		t.Fatalf("%+v", out.Atividades)
	}
	if out.Endereco.CdTipoLogradouro == nil || *out.Endereco.CdTipoLogradouro != "Rua" {
		t.Fatalf("tipo %+v", out.Endereco.CdTipoLogradouro)
	}
	if out.Endereco.DsLogradouro == nil || *out.Endereco.DsLogradouro != "Voluntarios da Franca" {
		t.Fatalf("logradouro %+v", out.Endereco.DsLogradouro)
	}
	if out.Endereco.IdCidadeIbge == nil || *out.Endereco.IdCidadeIbge != 3516200 {
		t.Fatal(out.Endereco.IdCidadeIbge)
	}
	if len(out.Telefones) != 1 || out.Telefones[0].NrTelefone != "1637112002" || out.Telefones[0].CdIso3Pais != "BRA" {
		t.Fatalf("%+v", out.Telefones)
	}
	if len(out.InscricoesEstaduais) != 1 || out.InscricoesEstaduais[0].CdEstado == nil || *out.InscricoesEstaduais[0].CdEstado != "SP" {
		t.Fatalf("%+v", out.InscricoesEstaduais)
	}
}

func TestMapCnpjWs(t *testing.T) {
	body := []byte(`{
	  "razao_social": "GOMES & MARINO LTDA",
	  "simples": {"simples": "Sim"},
	  "estabelecimento": {
	    "tipo": "Matriz",
	    "nome_fantasia": "GEOSOLOS",
	    "situacao_cadastral": "Ativa",
	    "data_inicio_atividade": "2008-12-05",
	    "tipo_logradouro": "AVENIDA",
	    "logradouro": "GOV. BENTO",
	    "numero": "1034",
	    "complemento": "  BLOCO 01  ",
	    "bairro": "ZONA 07",
	    "cep": "87030010",
	    "ddd1": "44",
	    "telefone1": "32263929",
	    "ddd2": "",
	    "telefone2": "99578078",
	    "email": "evitoria@teracom.com.br",
	    "atividade_principal": {"id": "4321500", "descricao": "Eletrica"},
	    "atividades_secundarias": [{"id": "4312600", "descricao": "Sondagens"}],
	    "cidade": {"ibge_id": 4115200},
	    "inscricoes_estaduais": [{"inscricao_estadual": "123", "estado": {"ibge_id": 41}}]
	  }
	}`)
	out, err := MapCnpjWs(body)
	if err != nil {
		t.Fatal(err)
	}
	if out.NmEmpresarial != "GOMES & MARINO LTDA" || out.CdTipoEstabelecimento != "MATRIZ" {
		t.Fatalf("%+v", out)
	}
	if out.FlOptanteSimplesNacional == nil || !*out.FlOptanteSimplesNacional {
		t.Fatal("simples")
	}
	if out.Endereco.CdTipoLogradouro == nil || *out.Endereco.CdTipoLogradouro != "AVENIDA" {
		t.Fatalf("tipo %+v", out.Endereco.CdTipoLogradouro)
	}
	if out.Endereco.DsLogradouro == nil || *out.Endereco.DsLogradouro != "GOV. BENTO" {
		t.Fatalf("logradouro %+v", out.Endereco.DsLogradouro)
	}
	if out.Endereco.CdTipoBairro == nil || *out.Endereco.CdTipoBairro != "ZONA" || out.Endereco.DsBairro == nil || *out.Endereco.DsBairro != "07" {
		t.Fatalf("bairro %+v %+v", out.Endereco.CdTipoBairro, out.Endereco.DsBairro)
	}
	if out.Endereco.DsComplemento == nil || *out.Endereco.DsComplemento != "BLOCO 01" {
		t.Fatalf("complemento %+v", out.Endereco.DsComplemento)
	}
	if len(out.Telefones) != 1 || out.Telefones[0].NrTelefone != "4432263929" {
		t.Fatalf("%+v", out.Telefones)
	}
	if out.Atividades[0].CdAtividade != "4321500" || out.Atividades[0].FlPrincipal != true {
		t.Fatalf("%+v", out.Atividades)
	}
	if out.InscricoesEstaduais[0].IdEstadoIbge == nil || *out.InscricoesEstaduais[0].IdEstadoIbge != 41 {
		t.Fatalf("%+v", out.InscricoesEstaduais)
	}
	if out.Endereco.VlLatitude != nil {
		t.Fatal("latitude deveria ser nula")
	}
}
