package doc

import "testing"

func TestDocuments(t *testing.T) {
	cnpj, err := CNPJ("47.960.950/0001-21")
	if err != nil || cnpj != "47960950000121" {
		t.Fatalf("%s %v", cnpj, err)
	}
	if _, err := CNPJ("123"); err == nil {
		t.Fatal("cnpj curto")
	}
	cep, err := CEP("14400-490")
	if err != nil || cep != "14400490" {
		t.Fatalf("%s %v", cep, err)
	}
	if _, err := CEP("1440049"); err == nil {
		t.Fatal("cep curto")
	}
}
