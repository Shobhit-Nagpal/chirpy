package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	serveMux := http.NewServeMux()

  pwd, err := os.Getwd()
  if err != nil {
    log.Fatal(err)
    return
  }

  serveMux.Handle("/", http.FileServer(http.Dir(pwd)))
  serveMux.Handle("/assets", http.FileServer(http.Dir(fmt.Sprintf("%s/assets", pwd))))

	server := &http.Server{
		Handler: serveMux,
		Addr:    "localhost:8080",
	}

	server.ListenAndServe()
}
