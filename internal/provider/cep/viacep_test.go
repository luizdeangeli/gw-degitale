package cep

import (
	"testing"

	"gw.degitale.com.br/internal/provider"
)

func TestMapViaCep(t *testing.T) {
	out, err := MapViaCep([]byte(`{
	  "cep": "01001-000",
	  "logradouro": "Praça da Sé",
	  "complemento": "lado ímpar",
	  "bairro": "Sé",
	  "localidade": "São Paulo",
	  "uf": "SP",
	  "ibge": "3550308",
	  "ddd": "11"
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if out.Ibge != "3550308" || out.Localidade != "São Paulo" || out.Ddd != "11" {
		t.Fatalf("%+v", out)
	}
}

func TestMapViaCepErro(t *testing.T) {
	_, err := MapViaCep([]byte(`{"erro": true}`))
	if !provider.IsKind(err, provider.KindNotFound) {
		t.Fatal(err)
	}
	_, err = MapViaCep([]byte(`{"erro": "true"}`))
	if !provider.IsKind(err, provider.KindNotFound) {
		t.Fatal(err)
	}
}
