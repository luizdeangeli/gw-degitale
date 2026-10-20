package doc

import "fmt"

func Digits(s string) string {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= '0' && c <= '9' {
			out = append(out, c)
		}
	}
	return string(out)
}

func CNPJ(s string) (string, error) {
	d := Digits(s)
	if len(d) != 14 {
		return "", fmt.Errorf("cnpj invalido")
	}
	return d, nil
}

func CEP(s string) (string, error) {
	d := Digits(s)
	if len(d) != 8 {
		return "", fmt.Errorf("cep invalido")
	}
	return d, nil
}
