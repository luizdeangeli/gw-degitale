package auth

import "testing"

func TestParseAndName(t *testing.T) {
	tokens, err := Parse(" smg:abc , sig:def ")
	if err != nil {
		t.Fatal(err)
	}
	name, ok := tokens.Name("def")
	if !ok || name != "sig" {
		t.Fatalf("name=%s ok=%v", name, ok)
	}
	if _, ok := tokens.Name("nope"); ok {
		t.Fatal("aceitou segredo invalido")
	}
}

func TestParseRejects(t *testing.T) {
	for _, raw := range []string{"", "semseparador", "nome:", ":segredo", "a:1,a:2", "a:1,b:1"} {
		if _, err := Parse(raw); err == nil {
			t.Fatalf("aceitou %q", raw)
		}
	}
}
