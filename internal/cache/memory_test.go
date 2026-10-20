package cache

import (
	"testing"
	"time"
)

func TestTTL(t *testing.T) {
	m := New()
	cur := time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC)
	m.now = func() time.Time { return cur }

	m.Set("cnpj:1", "cnpjws", "dados", time.Hour)
	provider, data, ok := m.Get("cnpj:1")
	if !ok || provider != "cnpjws" || data != "dados" {
		t.Fatalf("cache miss inesperado: %s %v %v", provider, data, ok)
	}

	cur = cur.Add(2 * time.Hour)
	if _, _, ok := m.Get("cnpj:1"); ok {
		t.Fatal("item expirado ainda presente")
	}
}
