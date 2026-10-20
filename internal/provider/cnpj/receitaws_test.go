package cnpj

import (
	"testing"

	"gw.degitale.com.br/internal/provider"
)

func TestMapReceitaWS(t *testing.T) {
	body := []byte(`{
	  "status": "OK",
	  "tipo": "MATRIZ",
	  "nome": "MAGAZINE LUIZA S/A",
	  "fantasia": "MAGAZINE LUIZA",
	  "abertura": "24/10/1966",
	  "situacao": "ATIVA",
	  "atividade_principal": [{"code": "47.13-0-04", "text": "Lojas"}],
	  "atividades_secundarias": [{"code": "00.00-0-00", "text": "Não informada"}],
	  "logradouro": "RUA VOLUNTARIOS DA FRANCA",
	  "numero": "1465",
	  "complemento": "",
	  "cep": "14.400-490",
	  "bairro": "CENTRO",
	  "email": "fiscal@example.com",
	  "telefone": "(16) 3711-2002 / (16) 99999-0000",
	  "simples": {"optante": false}
	}`)
	out, err := MapReceitaWS(body)
	if err != nil {
		t.Fatal(err)
	}
	if out.NmEmpresarial != "MAGAZINE LUIZA S/A" || out.CdTipoEstabelecimento != "MATRIZ" {
		t.Fatalf("%+v", out)
	}
	if out.DtAbertura != "1966-10-24" {
		t.Fatal(out.DtAbertura)
	}
	if out.FlOptanteSimplesNacional == nil || *out.FlOptanteSimplesNacional {
		t.Fatal("simples")
	}
	if len(out.Atividades) != 1 || out.Atividades[0].CdAtividade != "4713004" || !out.Atividades[0].FlPrincipal {
		t.Fatalf("%+v", out.Atividades)
	}
	if out.Endereco.CdTipoLogradouro == nil || *out.Endereco.CdTipoLogradouro != "RUA" {
		t.Fatalf("tipo %+v", out.Endereco.CdTipoLogradouro)
	}
	if out.Endereco.DsLogradouro == nil || *out.Endereco.DsLogradouro != "VOLUNTARIOS DA FRANCA" {
		t.Fatalf("logradouro %+v", out.Endereco.DsLogradouro)
	}
	if out.Endereco.NrCep == nil || *out.Endereco.NrCep != "14400490" {
		t.Fatalf("cep %+v", out.Endereco.NrCep)
	}
	if len(out.Telefones) != 2 || out.Telefones[0].NrTelefone != "1637112002" {
		t.Fatalf("%+v", out.Telefones)
	}
	if len(out.Emails) != 1 || out.Emails[0].DsEmail != "fiscal@example.com" {
		t.Fatalf("%+v", out.Emails)
	}
}

func TestMapReceitaWSRejeitado(t *testing.T) {
	_, err := MapReceitaWS([]byte(`{"status":"ERROR","message":"CNPJ rejeitado pela Receita Federal"}`))
	if !provider.IsKind(err, provider.KindNotFound) {
		t.Fatal(err)
	}
}
