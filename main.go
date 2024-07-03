package main

import (
//	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
//	dbg := flag.Bool("debug", false, "Enable debug mode")
//	flag.Parse()

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

	//Chirp routes
	serveMux.HandleFunc("POST /api/chirps", handleCreateChirp)
	serveMux.HandleFunc("GET /api/chirps", handleGetChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpId}", handleGetChirpById)

  //User routes
	serveMux.HandleFunc("POST /api/users", handleCreateUser)
	serveMux.HandleFunc("POST /api/login", handleLogin)

	server := &http.Server{
		Handler: serveMux,
		Addr:    "localhost:8080",
	}

	server.ListenAndServe()
}
