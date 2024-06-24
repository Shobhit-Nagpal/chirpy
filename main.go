package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func handleReadiness(w http.ResponseWriter, req *http.Request) {
  w.Header().Set("Content-Type", "text/plain; charset=utf-8")
  io.WriteString(w, "OK")
  w.WriteHeader(http.StatusOK)
}

func main() {
	serveMux := http.NewServeMux()

  pwd, err := os.Getwd()
  if err != nil {
    log.Fatal(err)
    return
  }

  serveMux.Handle("/app/*", http.StripPrefix("/app/", http.FileServer(http.Dir(pwd))))
  serveMux.Handle("/assets", http.FileServer(http.Dir(fmt.Sprintf("%s/assets", pwd))))
  serveMux.HandleFunc("/healthz", handleReadiness)

	server := &http.Server{
		Handler: serveMux,
		Addr:    "localhost:8080",
	}

	server.ListenAndServe()
}
