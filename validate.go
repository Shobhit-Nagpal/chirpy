package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

func handleValidateChirp(w http.ResponseWriter, req *http.Request) {
	type ReqBody struct {
		Body string `json:"body"`
	}

	type Err struct {
		Error string `json:"error"`
	}

	type Response struct {
		Valid bool `json:"valid"`
	}

	type CleanedResponse struct {
		CleanedBody string `json:"cleaned_body"`
	}

	bodyFromReq, err := io.ReadAll(req.Body)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	reqBody := ReqBody{}

	err = json.Unmarshal(bodyFromReq, &reqBody)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if len(reqBody.Body) > 140 {
		respondWithError(w, http.StatusBadRequest, "Chirp is too long")
		return
	}

	cleanedMsg := cleanMessage(reqBody.Body)

	response := CleanedResponse{
		CleanedBody: cleanedMsg,
	}

	err = respondWithJSON(w, http.StatusOK, response)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}
	return
}

func cleanMessage(msg string) string {
	words := strings.Split(msg, " ")
	for idx, word := range words {
		switch strings.ToLower(word) {
		case "kerfuffle":
      words[idx] = "****"
		case "sharbert":
      words[idx] = "****"
		case "fornax":
      words[idx] = "****"
		}
	}

  cleanMsg := strings.Join(words, " ")
	return cleanMsg
}
