package model

import "testing"

func TestSplitTipo(t *testing.T) {
	street := "Rua Voluntarios da Franca"
	tipo, resto := SplitTipo(nil, &street)
	if *tipo != "Rua" || *resto != "Voluntarios da Franca" {
		t.Fatalf("%q %q", *tipo, *resto)
	}

	bairro := "Centro"
	tipo, resto = SplitTipo(nil, &bairro)
	if *tipo != "Centro" || *resto != "" {
		t.Fatalf("%q %q", *tipo, *resto)
	}

	already := "AVENIDA"
	valor := "GOV BENTO"
	tipo, resto = SplitTipo(&already, &valor)
	if *tipo != "AVENIDA" || *resto != "GOV BENTO" {
		t.Fatalf("nao deveria repartir: %q %q", *tipo, *resto)
	}
}
