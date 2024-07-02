package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/Shobhit-Nagpal/chirpy/internal/database"
)

func handleCreateChirp(w http.ResponseWriter, req *http.Request) {
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
		Id   int    `json:"id"`
		Body string `json:"body"`
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

	//Insert into db here
	path, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	db, err := database.NewDB(path + "/database.json")
	if err != nil {
		log.Printf("Error creating DB connection: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't connect to DB")
		return
	}

	chirp, err := db.CreateChirp(cleanedMsg)
	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}

	response := CleanedResponse{
		Id:   chirp.Id,
		Body: chirp.Body,
	}

	err = respondWithJSON(w, http.StatusCreated, response)
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

func handleGetChirps(w http.ResponseWriter, req *http.Request) {

	path, err := os.Getwd()
	if err != nil {
		log.Printf("Error getting current directory: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	db, err := database.NewDB(path + "/database.json")
	if err != nil {
		log.Printf("Error creating DB connection: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't connect to DB")
		return
	}

	chirps, err := db.GetChirps()
	if err != nil {
		log.Printf("Error getting chirps: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Couldn't get chirps")
		return
	}

	err = respondWithJSON(w, http.StatusOK, chirps)
	if err != nil {
		log.Printf("Error encoding to json: %s", err)
		err = respondWithError(w, http.StatusInternalServerError, "Something went wrong")
	}
	return
}
