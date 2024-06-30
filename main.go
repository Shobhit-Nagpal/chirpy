package main

import (
	"log"
	"net/http"
	"os"
)

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
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handleAdminMetrics)
	serveMux.HandleFunc("GET /api/metrics", apiCfg.handleMetrics)
	serveMux.HandleFunc("GET /api/reset", apiCfg.handleReset)
	serveMux.HandleFunc("POST /api/validate_chirp", handleValidateChirp)

	server := &http.Server{
		Handler: serveMux,
		Addr:    "localhost:8080",
	}

	server.ListenAndServe()
}
