package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

type apiConfig struct {
	fileserverHits int
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits++
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handleReset(w http.ResponseWriter, req *http.Request) {
	cfg.fileserverHits = 0
	w.WriteHeader(http.StatusOK)
}

func (cfg *apiConfig) handleMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-9")
	io.WriteString(w, fmt.Sprintf("Hits: %d", cfg.fileserverHits))
	w.WriteHeader(http.StatusOK)
}

func handleReadiness(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-9")
	io.WriteString(w, "OK")
	w.WriteHeader(http.StatusOK)
}

func main() {
	apiCfg := &apiConfig{
		fileserverHits: 0,
	}
	serveMux := http.NewServeMux()

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		return
	}

	serveMux.Handle("GET /app/*", apiCfg.middlewareMetricsInc(http.StripPrefix("/app/", http.FileServer(http.Dir(pwd)))))
	serveMux.HandleFunc("GET /api/healthz", handleReadiness)
	serveMux.HandleFunc("GET /api/metrics", apiCfg.handleMetrics)
	serveMux.HandleFunc("GET /api/reset", apiCfg.handleReset)

	server := &http.Server{
		Handler: serveMux,
		Addr:    "localhost:8080",
	}

	server.ListenAndServe()
}
