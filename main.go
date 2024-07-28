package main

import (
	//	"flag"
	"github.com/joho/godotenv"
//  "github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"os"
)

func main() {
	//	dbg := flag.Bool("debug", false, "Enable debug mode")
	//	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

  jwt_secret := os.Getenv("JWT_SECRET")

	apiCfg := &apiConfig{
		fileserverHits: 0,
    jwtSecret: []byte(jwt_secret),
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
	serveMux.HandleFunc("POST /api/chirps", apiCfg.handleCreateChirp)
	serveMux.HandleFunc("GET /api/chirps", apiCfg.handleGetChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpId}", handleGetChirpById)

	//User routes
	serveMux.HandleFunc("POST /api/users", handleCreateUser)
	serveMux.HandleFunc("POST /api/login", apiCfg.handleLogin)
	serveMux.HandleFunc("POST /api/refresh", apiCfg.handleRefreshToken)
	serveMux.HandleFunc("POST /api/revoke", apiCfg.handleRevokeToken)
	serveMux.HandleFunc("PUT /api/users", apiCfg.handleUpdateUser)

	server := &http.Server{
		Handler: serveMux,
		Addr:    "localhost:8080",
	}

	server.ListenAndServe()
}
