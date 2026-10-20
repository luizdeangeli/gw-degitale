package model

import "strings"

func StrPtr(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	v := s
	return &v
}

func IntPtr(n int) *int {
	if n == 0 {
		return nil
	}
	v := n
	return &v
}

// SplitTipo reproduz o strtok do SMG: o primeiro termo vira o tipo e o resto fica no valor.
// Se não houver espaço, o valor inteiro vira tipo e o restante fica vazio.
func SplitTipo(tipo, valor *string) (*string, *string) {
	if valor == nil || strings.TrimSpace(*valor) == "" {
		return tipo, valor
	}
	if tipo != nil && strings.TrimSpace(*tipo) != "" {
		return tipo, valor
	}
	s := strings.TrimSpace(*valor)
	head, tail, found := strings.Cut(s, " ")
	h := head
	if !found {
		empty := ""
		return &h, &empty
	}
	t := strings.TrimSpace(tail)
	return &h, &t
}

func DateOnly(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 10 && s[4] == '-' && s[7] == '-' {
		return s[:10]
	}
	return s
}
