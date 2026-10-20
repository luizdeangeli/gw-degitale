package cep

import "testing"

func TestMapBrasilAPI(t *testing.T) {
	out, err := MapBrasilAPI([]byte(`{
	  "cep": "01001000",
	  "state": "SP",
	  "city": "São Paulo",
	  "neighborhood": "Sé",
	  "street": "Praça da Sé",
	  "ibge": {"city": "3550308", "state": "35"}
	}`))
	if err != nil {
		t.Fatal(err)
	}
	if out.Cep != "01001-000" || out.Uf != "SP" || out.Localidade != "São Paulo" || out.Ibge != "3550308" {
		t.Fatalf("%+v", out)
	}
	if out.TipoLogradouro != "Praça" || out.Logradouro != "da Sé" {
		t.Fatalf("tipo=%q logradouro=%q", out.TipoLogradouro, out.Logradouro)
	}
}
