package httpapi

import (
	"embed"
	"net/http"
)

//go:embed openapi.yaml docs.html
var docsFS embed.FS

func (s *Server) openapi(w http.ResponseWriter, _ *http.Request) {
	writeStatic(w, "application/yaml; charset=utf-8", "openapi.yaml")
}

func (s *Server) docs(w http.ResponseWriter, _ *http.Request) {
	writeStatic(w, "text/html; charset=utf-8", "docs.html")
}

func writeStatic(w http.ResponseWriter, contentType, name string) {
	body, err := docsFS.ReadFile(name)
	if err != nil {
		http.Error(w, "documentacao indisponivel", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
